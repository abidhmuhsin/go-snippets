/*
aws sdk v2
Upload Page allows user to upload new files to s3 bucket.
List page retrieves and shows all available filenames in given bucket
Clicking on download against filename allows to download the file.
Clicking on signed link generates a signed download link from s3 and displays it.

Docker
docker pull minio/minio
docker run -p 9000:9000 minio/minio server /data
Note: /data (D:\data) is the root directory for MinIO. All the buckets and files will be created into this location only.

Default admin User id/access key: minioadmin and password/secret key: minioadmin

The service will be available on http://localhost:9000 or http://${hostname}:9000.
Port 9000 is default port. However, you can configure it using ./minio server /data --address ":9000"

*/

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	AWS_S3_REGION = ""        // Region
	AWS_S3_BUCKET = "uploads" // Bucket
)

// We will be using this client everywhere in our code
var awsS3Client *s3.Client

func main() {
	configS3()
	createBucket()

	http.HandleFunc("/", handlerHome)
	http.HandleFunc("/upload", handlerUpload)          // Upload: /upload (upload file named "file")
	http.HandleFunc("/download", handlerDownload)      // Download: /download?key={key of the object}&filename={name for the new file}
	http.HandleFunc("/presigned", handlerPresignedUrl) // Presigned:  /presigned?key={key of the object}
	http.HandleFunc("/list", handlerList)              // List: /list?prefix={prefix}&delimeter={delimeter}
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// configS3 creates the S3 client
func _configS3() {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AWS_S3_REGION))
	if err != nil {
		log.Fatal(err)
	}

	awsS3Client = s3.NewFromConfig(cfg)
}

// configS3 creates the S3 client
func __configS3() {
	const accessKey = "minioadmin" //your_access_key
	const secretKey = "minioadmin" //your_secret_key
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(AWS_S3_BUCKET))
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	awsS3Client = s3.NewFromConfig(cfg)

	log.Printf("awsS3Client: %v", awsS3Client)
}

func configS3() {
	const defaultRegion = "us-east-1"
	//https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/endpoints/
	staticResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               "http://localhost:9000", // or where ever you ran minio
			SigningRegion:     defaultRegion,
			HostnameImmutable: true,
		}, nil
	})

	cfg := aws.Config{
		Region:                      defaultRegion,
		Credentials:                 credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", ""),
		EndpointResolverWithOptions: staticResolver,
	}

	awsS3Client = s3.NewFromConfig(cfg)

	// log.Printf("awsS3Client: %v", awsS3Client)
}

func showError(w http.ResponseWriter, r *http.Request, status int, message string) {
	http.Error(w, message, status)
}

func createBucket() {
	cparams := &s3.CreateBucketInput{
		Bucket: aws.String(AWS_S3_BUCKET), // Required
	}
	_, err := awsS3Client.CreateBucket(context.TODO(), cparams)
	if err != nil {
		fmt.Printf("create bucket err", err.Error())
	}
}

//-----------------------------------------
func handlerList(w http.ResponseWriter, r *http.Request) {

	// There aren't really any folders in S3, but we can emulate them by using "/" in the key names
	// of the objects. In case we want to listen the contents of a "folder" in S3, what we really need
	// to do is to list all objects which have a certain prefix.
	prefix := r.URL.Query().Get("prefix")
	delimeter := r.URL.Query().Get("delimeter")

	paginator := s3.NewListObjectsV2Paginator(awsS3Client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(AWS_S3_BUCKET),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String(delimeter),
	})

	w.Header().Set("Content-Type", "text/html")

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			// Error handling goes here
			log.Printf("err: %v", err)
		}
		for _, obj := range page.Contents {
			template := `
			<li>
			File :   %s 
			&nbsp;&nbsp; 
			<a href='/download?key=%s'>Download</a>	
			&nbsp;|&nbsp; 		
			<a href='/presigned?key=%s'>Get presigned url</a>
			</li>
			
			`
			// Do whatever you need with each object "obj"
			fmt.Fprintf(w, template, *obj.Key, *obj.Key, *obj.Key)
		}
	}

	return
}

/*
	This function implements s3 upload which uploads the file from form by name(myFile)
	eg usage: 	http://localhost:8000/download?key=file_in_bucket.jpg
*/
func handlerUpload(w http.ResponseWriter, r *http.Request) {

	fmt.Println("upload handler...")

	r.ParseMultipartForm(10 << 20)

	// Get a file from the form input name "file"
	file, header, err := r.FormFile("myFile")
	if err != nil {
		showError(w, r, http.StatusInternalServerError, "Something went wrong retrieving the file from the form")
		return
	}
	defer file.Close()

	filename := header.Filename

	fmt.Printf("Uploading .. %s", filename)
	uploader := manager.NewUploader(awsS3Client)
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		// Do your error handling here
		showError(w, r, http.StatusInternalServerError, "Something went wrong uploading the file")
		return
	}

	fmt.Fprintf(w, "Successfully uploaded to %q\n", AWS_S3_BUCKET)
	return

}

/*
	This function implements s3 download which downloads the file with given s3 key, urlparam
	and serves it via http response.Bucket name is provided globally.
	eg usage: 	http://localhost:8000/download?key=file_in_bucket.jpg
*/
func handlerDownload(w http.ResponseWriter, r *http.Request) {

	fmt.Println("download handler...")

	// We get the name of the file on the URL
	key := r.URL.Query().Get("key")

	// We get the name of the file on the URL
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		filename = key
	}

	// Create the file
	newFile, err := os.Create(filename)
	if err != nil {
		showError(w, r, http.StatusBadRequest, "Something went wrong creating the local file")
	}
	defer newFile.Close()
	defer os.Remove(newFile.Name()) // since we download via browser delete once handler completes

	fmt.Printf("Downloading .. %s", key)
	downloader := manager.NewDownloader(awsS3Client)
	_, err = downloader.Download(context.TODO(), newFile, &s3.GetObjectInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(key),
	})

	if err != nil {
		showError(w, r, http.StatusBadRequest, "Something went wrong retrieving the file from S3")
		return
	}

	w.Header().Set("Content-Disposition", "inline; filename="+filename)
	http.ServeFile(w, r, newFile.Name())
}

func handlerPresignedUrl(w http.ResponseWriter, r *http.Request) {

	fmt.Println("presigned url handler...")
	EXPIRY_SECONDS := 10

	// We get the name of the file on the URL
	key := r.URL.Query().Get("key")

	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(key),
	}

	psClient := s3.NewPresignClient(awsS3Client)

	// presignedReq, err := psClient.PresignGetObject(context.TODO(), getObjInput)
	presignedReq, err := psClient.PresignGetObject(context.TODO(), getObjInput, func(opt *s3.PresignOptions) {
		opt.Expires = time.Duration(EXPIRY_SECONDS) * time.Second // Set signed url expiry to 10seconds
	})
	if err != nil {
		log.Println("Failed to sign request", err)
	}

	// log.Printf("The URL is %s\n", presignedReq.URL)
	fmt.Fprintf(w, "<div>Presigned URL for %s valid for %d seconds <br><br><a href='%s' target='_blank'>%s</a><div>", key, EXPIRY_SECONDS, presignedReq.URL, presignedReq.URL)

}

func handlerHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("home handler...")

	const template = `
					<!DOCTYPE html>
					<html lang="en">
					<head>
						<meta charset="UTF-8" />
						<meta name="viewport" content="width=device-width, initial-scale=1.0" />
						<meta http-equiv="X-UA-Compatible" content="ie=edge" />
						<title>Document</title>
					</head>
					<body>
						<div>
							<h2> S3 upload/list/download using aws sdk v2 </h2>
							<h3>Bucket [%s]</h3>
							<br>
							
							<a href="/list">List all files </a>
							<br><br>

							<form
							enctype="multipart/form-data"
							action="http://localhost:8000/upload"
							method="post"
							>
							<input type="file" name="myFile" />
							<input type="submit" value="upload" />
							</form>
						</div>
					</body>
					</html>
					`
	fmt.Fprintf(w, template, AWS_S3_BUCKET)
}
