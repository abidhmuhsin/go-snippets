package main

// Upload download delete via minio sdk

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var s3Client *minio.Client

func main() {
	setupClient()
	createBucket("mymusic")
	uploadFile()
	downloadFile()
	// deleteObject()
	deleteMultipleObjects()

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

func uploadFile() {
	ctx := context.Background()

	bucketName := "mymusic"

	// Upload the txt file
	objectName := "golden-oldies.txt" // filename in s3. Use "folder1/folder2/filename.txt to use folders"
	filePathToRead := "./s3_upload_test.txt"
	contentType := "application/text"

	// Upload the zip file with FPutObject
	info, err := s3Client.FPutObject(ctx, bucketName, objectName, filePathToRead, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
}

func downloadFile() {
	bucketName := "mymusic"

	objectName := "golden-oldies.txt" // filename in s3
	filePathToWrite := "./s3_download_test.txt"

	err := s3Client.FGetObject(context.Background(), bucketName, objectName, filePathToWrite, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("Successfully downloaded %s\n", objectName)
}

func deleteObject() {

	bucketName := "mymusic"
	objectName := "golden-oldies.txt"

	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
		// VersionID:        "myversionid",
	}
	err := s3Client.RemoveObject(context.Background(), bucketName, objectName, opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("Successfully deleted %s\n", objectName)
}

func deleteMultipleObjects() {

	bucketName := "mymusic"
	prefixFolderName := "" // may use "" or "/" for root and  "folder1/folder2" for a specific folder.

	listObjectOptions := minio.ListObjectsOptions{
		Prefix:    prefixFolderName,
		Recursive: true,
	}
	objectsCh := make(chan minio.ObjectInfo)

	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range s3Client.ListObjects(context.Background(), bucketName, listObjectOptions) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			fmt.Println("To delete -- ", object.Key)
			objectsCh <- object
		}
	}()

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	for rErr := range s3Client.RemoveObjects(context.Background(), bucketName, objectsCh, opts) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

	log.Printf("Bulk Deletion completed for prefix:%s\n", prefixFolderName)
}
