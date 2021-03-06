package main

/* minio sdk
shows upload page.
allows file upload to s3 from web
*/

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var s3Client *minio.Client

func main() {
	setupClient()
	createBucket("uploads")
	// uploadFileFromStream()
	// downloadFileAsStream()
	setupRoutes()

}
func setupClient() {
	endpoint := "localhost:9000" //"play.min.io" // "s3.amazonaws.com"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	// Initialize minio client object.
	minioS3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%#v\n", minioS3Client) // minioClient is now setup
	s3Client = minioS3Client

}

func createBucket(bucketName string) {
	ctx := context.Background()

	// Make a new bucket called mymusic.
	// bucketName := "mymusic"
	location := "us-east-1"

	err := s3Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := s3Client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
}

func uploadFileStreamToS3(reader io.Reader, bucketName string, objectName string, size int64) (info minio.UploadInfo, err error) {

	// size = int64(-1)
	//For size input as -1 PutObject does a multipart Put operation until input stream reaches EOF. Maximum object size -5TiB.
	// WARNING:
	// Passing down '-1' will use memory and these cannot be reused for best outcomes for PutObject(), pass the size always.

	uploadInfo, err := s3Client.PutObject(context.Background(), bucketName, objectName,
		reader, size,
		// minio.PutObjectOptions{ContentType: "application/octet-stream", PartSize: uint64(50 << 20)} // partsize will limit max size
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Successfully uploaded bytes: ", uploadInfo)

	return uploadInfo, err
}

//-------------------------------------

func setupRoutes() {
	http.HandleFunc("/", landingPageHandler)
	http.HandleFunc("/upload", uploadFileHandler)
	err := http.ListenAndServe(":6060", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Landing page Hit")
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
						<form
						enctype="multipart/form-data"
						action="http://localhost:6060/upload"
						method="post"
						>
						<input type="file" name="myFile" />
						<input type="submit" value="upload" />
						</form>
					</body>
					</html>
					`
	fmt.Fprintf(w, template)
}
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, fileHeader, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
	fmt.Printf("File Size: %+v\n", fileHeader.Size)
	fmt.Printf("MIME Header: %+v\n", fileHeader.Header)

	// Create the uploads folder if it doesn't
	// already exist
	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uniqueFileName := fmt.Sprintf("%s_%d%s", filepath.Base(fileHeader.Filename), time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))

	/*** STREAMING READ from /tmp/multipart-** file autogenerated by r.FormFile("myFile") ****/
	// _, err = io.Copy(tempFile, file)
	pr := &Progress{
		TotalSize: fileHeader.Size,
	}
	// _, err = io.Copy(tempFile, io.TeeReader(file, pr))
	// uploadInfo, err := uploadFileStreamToS3(file, "uploads", uniqueFileName, fileHeader.Size)
	uploadInfo, err := uploadFileStreamToS3(io.TeeReader(file, pr), "uploads", uniqueFileName, fileHeader.Size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File to s3\n%s\n", uploadInfo.Key)
}

//--START:Progress Helper Struct---------------

// Progress is used to track the progress of a file upload.
// It implements the io.Writer interface so it can be passed
// to an io.TeeReader()
type Progress struct {
	TotalSize int64
	BytesRead int64
}

// Write is used to satisfy the io.Writer interface.
// Instead of writing somewhere, it simply aggregates
// the total bytes on each read
func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pr.BytesRead += int64(n)
	pr.Print()
	return
}

// Print displays the current progress of the file upload
// each time Write is called
func (pr *Progress) Print() {

	percentRead := (pr.BytesRead * 100) / pr.TotalSize
	fmt.Printf("File upload in progress:[%d%%]  %d/%d Bytes\n", percentRead, pr.BytesRead, pr.TotalSize)
	//Use the special verb %%, which consumes no argument, to write a literal percent sign:
	//fmt.Printf("%d %%", 50) // prints "50 %"

	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")
		return
	}
}

//--END:Progress Helper Struct-------------
