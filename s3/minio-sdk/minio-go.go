package main

//basic minio sdk client setup and create bucket
//docker run -p 9000:9000 minio/minio server /data

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var s3Client *minio.Client

func main() {
	setupClient()
	createBucket("mymusic")

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
