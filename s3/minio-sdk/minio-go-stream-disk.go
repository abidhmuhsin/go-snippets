package main

// minio sdk to stream file upload download from local disk to s3

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var s3Client *minio.Client

func main() {
	setupClient()
	createBucket("mymusic-streaming")
	uploadFileFromStream()
	downloadFileAsStream()

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

func uploadFileFromStream() {

	bucketName := "mymusic-streaming"
	objectName := "golden-oldies.txt" // filename in s3
	filePathToRead := "./s3_upload_test.txt"

	file, err := os.Open(filePathToRead)
	if err != nil {
		fmt.Println(err, filePathToRead)
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	uploadInfo, err := s3Client.PutObject(context.Background(), bucketName, objectName, file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Successfully uploaded bytes: ", uploadInfo)
}

func downloadFileAsStream() {
	bucketName := "mymusic-streaming"

	// Download the txt file
	objectName := "golden-oldies.txt" // filename in s3
	filePathToWrite := "./s3_download_test.txt"

	object, err := s3Client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
	localFile, err := os.Create(filePathToWrite)
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err = io.Copy(localFile, object); err != nil {
		fmt.Println(err, objectName)
		return
	}
	log.Printf("Successfully downloaded %s to %s\n", objectName, filePathToWrite)
}
