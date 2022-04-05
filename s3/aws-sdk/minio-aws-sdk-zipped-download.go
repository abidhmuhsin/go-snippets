package main

/*
lists and downloads all files from a given bucket.
Use minio docker with a bucket having some pre uploaded files.

Method1:
Zips them via channel as soon as each file is downloaded.
New file download needs to wait until prev downloaded file is zipped.
One task at time processed, either download or zip.

Method2:
Zips them via channel as soon as each file is downloaded.
zipping prev files and downloading next file will run in parellel.
zipping will wait for next file to finish downloading if all files already downloaded done adding to zip.
Download and zip are non blocking wrt each other.
Zip will be waiting asynchronously if no files to zip in downloaded queue
*/
import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var s3Client *s3.S3
var newSession *session.Session

func main() {

	printMemoryEveryNseconds(5)
	fmt.Println(time.Now())

	s3Config()
	filesToDownload := listObjects("uploads")
	// downloadFiles("uploads", filesToDownload)
	// downloadFilesAndZipOneByOne("uploads", filesToDownload)
	downloadFilesAndZipViaChannel("uploads", filesToDownload)

	fmt.Println(time.Now())
}

func downloadFiles(bucketName string, objectKeys []string) {

	replaceExisting := true

	// Create download folder
	err := os.MkdirAll("./downloads", os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create folder", err)
		return
	}

	for _, fileKey := range objectKeys {
		// Create one file per object
		filepath := "./downloads/" + fileKey
		if _, err := os.Stat(filepath); err == nil {
			// If file exists
			if replaceExisting == true {
				e := os.Remove(filepath)
				if e != nil {
					fmt.Printf("Couldn't delete existing file: %s, skipping..\n", fileKey)
					continue
				}
			} else {
				// if no replace skip
				fmt.Printf("Skipping existing file: %s\n", fileKey)
				continue
			}
		}
		file, err := os.Create(filepath) //create file at path
		if err != nil {
			fmt.Println("Failed to create file", err)
		}
		defer file.Close()

		fmt.Printf("Download file: %s\n", fileKey)
		// Download from s3 into filestream created.
		// downloader := s3manager.NewDownloader(newSession, func(d *s3manager.Downloader) {
		// 	d.PartSize = 64 * 1024 * 1024 // 64MB per part
		// 	d.Concurrency = 1 // 1= will be slow single connection for direct streaming without seeking or chunks.
		// })
		downloader := s3manager.NewDownloader(newSession)
		_, err = downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(fileKey),
			})
		if err != nil {
			fmt.Println("Failed to download file", err)
			return
		}
	}
	fmt.Println("Download completed at ", "./downloads/*")
}

func downloadFilesAndZipOneByOne(bucketName string, objectKeys []string) {

	//Create temp dir for downloading & zipping
	tempDir, err := ioutil.TempDir("./", "downloads-*") //create folder at path
	if err != nil {
		fmt.Println(err)
	}
	// defer os.RemoveAll(tempDir) -- remove temp dir once zip file processed or uploaded

	// Create a zip file
	file, err := ioutil.TempFile(tempDir, "s3download-*.zip") //create file at path
	if err != nil {
		fmt.Println("Failed to create file", err)
	}
	defer file.Close()
	zipFile := zip.NewWriter(file)
	defer zipFile.Close()

	// Download one by one
	for _, fileKey := range objectKeys {

		tempFile, err := ioutil.TempFile(tempDir, "s3-tmp-*")
		if err != nil {
			fmt.Println("Failed to create file", err)
		}
		// defer tempFile.Close() -- we will just remove file as soon as added to zip

		fmt.Printf("Download file: %s\n", fileKey)
		// Download from s3 into tmp filestream created.
		downloader := s3manager.NewDownloader(newSession)
		_, err = downloader.Download(tempFile,
			&s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(fileKey),
			})
		if err != nil {
			fmt.Println("Failed to download file", err)
			return
		}
		// zipWriter can only write one at a time due to zip algorithms.No parellel processing
		// go func(){ -- wont work -- zip corrupted
		zipWriter, _ := zipFile.Create(fileKey)
		fmt.Printf("Adding %s to zip\n", fileKey)
		io.Copy(zipWriter, tempFile)
		// Added to zip, remove tempFile
		os.Remove(tempFile.Name())
		// }()
	}
	fmt.Println("Download completed at ", "./downloads/*")
}

func downloadFilesAndZipViaChannel(bucketName string, objectKeys []string) {

	//Create temp dir for downloading & zipping
	tempDir, err := ioutil.TempDir("./", "downloads-*") //create folder at path
	if err != nil {
		fmt.Println(err)
	}
	defer os.RemoveAll(tempDir) //-- optionally remove temp dir once zip file processed or uploaded

	files := make(chan *os.File)
	zipFilePath := tempDir + "/download.zip"
	wg := AsyncZipWriter(zipFilePath, files)

	// Download one by one
	for _, fileKey := range objectKeys {

		tempFile, err := os.Create(tempDir + "/" + fileKey) // ioutil.TempFile(tempDir, "*_"+fileKey)
		if err != nil {
			fmt.Println("Failed to create file", err)
		}
		// --defer tempFile.Close() -- we will just remove file as soon as added to zip

		fmt.Printf("Downloading file: %s...\n", fileKey)
		// Download from s3 into tmp filestream created.
		downloader := s3manager.NewDownloader(newSession)
		_, err = downloader.Download(tempFile,
			&s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(fileKey),
			})
		if err != nil {
			fmt.Println("Failed to download file", err)
			return
		}
		// zipWriter can only write one at a time due to zip algorithms.No parellel processing
		// So send downloaded files to channel that will process one by one and delete
		files <- tempFile
		// Sent to zip channel and continue downloading next file.
		// Zipping and downloading will happen in parellel.
		// --os.Remove(tempFile.Name()) ,-- remove tempFile inside channel receiver after zipping..not here.
	}
	fmt.Println("Downloads completed at ", tempDir)

	// Once we're done sending the files, we can close the channel.
	// close(files) channel close will cause ZipWriter to break out of the loop,
	//  close the file, and unblock the next mutex:
	close(files)
	//Just Wait from exiting main program until our waitgroup is done.ie all files are zipped.
	wg.Wait()
	fmt.Println("Zipping completed at ", zipFilePath)

	uploadToS3("zippedfiles", zipFilePath)

}

func s3Config() {
	const accessKey = "minioadmin" //your_access_key
	const secretKey = "minioadmin" //your_secret_key

	// Configure to use MinIO Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String("http://localhost:9000"),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession = session.New(s3Config)

	s3Client = s3.New(newSession)
}

func listObjects(bucketName string) (fileKeysList []string) {
	fileKeysList = []string{}

	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName), // Required
		// Delimiter:    aws.String("Delimiter"),
		// EncodingType: aws.String("EncodingType"),
		// Marker:       aws.String("Marker"),
		// MaxKeys:      aws.Int64(1),
		// Prefix:       aws.String("Prefix"),
	}
	resp, err := s3Client.ListObjects(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}
	// Pretty-print the response data.
	// fmt.Println(resp.Contents)

	for _, item := range resp.Contents {
		fmt.Printf("Filename: %s, Size:%dBytes\n", aws.StringValue(item.Key), aws.Int64Value(item.Size))
		fileKeysList = append(fileKeysList, aws.StringValue(item.Key))
	}
	// fmt.Println(fileKeysList)
	return fileKeysList

}

func uploadToS3(bucketName string, filepath string) {

	cparams := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName), // Required
	}
	// Create a new bucket using the CreateBucket call.
	_, err := s3Client.CreateBucket(cparams)
	if err != nil {
		// Message from an error.
		fmt.Println(err.Error())
		// return -- process bucket existing error
	}

	fmt.Printf("S3 upload started\n")
	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(newSession, func(u *s3manager.Uploader) {
		// u.PartSize = 5 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		// u.Concurrency = 2            // default is 5
	})

	//open the file
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("failed to open file %q, %v", filepath, err)
		return
	}
	defer f.Close()

	fileName, _ := os.Stat(f.Name()) // To get proper filename part alone from file stream f.

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName.Name()),
		Body:   f,
	})

	//in case it fails to upload
	if err != nil {
		fmt.Printf("failed to upload file, %v", err)
		return
	}
	fmt.Printf("File uploaded to, %s.  size:%d bytes\n", result.Location, fileName.Size())
}

/*
	- AsyncZipWriter accepts a zipfile location to create new zipfile and a channel to receive files.
	- The channel listens for new files that will be added to zip in order as and when received.
	- It creates and returns a WaitGroup[1] whose  wg.Done() will be called by deferring until the channel is closed.
	and zipping loop is exited.
	- Returned wg ref can be used by the caller to wait for compression channel to be completed.
*/
func AsyncZipWriter(zipFilePath string, files chan *os.File) *sync.WaitGroup {
	f, err := os.Create(zipFilePath)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	zw := zip.NewWriter(f)
	go func() {
		// Note the order (LIFO):
		defer wg.Done() // 2. signal that we're done
		defer f.Close() // 1. close the file
		var err error
		var fw io.Writer
		fmt.Println("*******AsyncZipWriter: Waiting for files on channel**********")
		for f := range files {
			// Loop until channel is closed.
			fmt.Printf("Adding to zip %s\n", f.Name())
			fileName, _ := os.Stat(f.Name()) // To get proper filename part alone from file stream f.
			if fw, err = zw.Create(fileName.Name()); err != nil {
				panic(err)
			}
			io.Copy(fw, f)
			if err = f.Close(); err != nil {
				panic(err)
			}
			fmt.Printf("Added to zip %s\n", f.Name())
			// Remove zipped temp file
			os.Remove(f.Name())
		}
		// The zip writer must be closed *before* f.Close() is called!
		if err = zw.Close(); err != nil {
			panic(err)
		}
		fmt.Println("*******AsyncZipWriter: Zipping all done**********")
	}()
	return &wg
}

//--------------------------------------------
func printMemoryEveryNseconds(seconds int64) {
	ticker := time.NewTicker(time.Duration(seconds) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				PrintMemUsage()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v", m.NumGC)
	fmt.Printf("\t%s \n", time.Now())
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
