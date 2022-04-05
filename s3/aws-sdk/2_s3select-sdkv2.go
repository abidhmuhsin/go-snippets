/*
https://eforexcel.com/wp/downloads-18-sample-csv-files-data-sets-for-testing-sales/
Region	Country	Item Type	Sales Channel	Order Priority	Order Date	Order ID	Ship Date	Units Sold	Unit Price	Unit Cost	Total Revenue	Total Cost	Total Profit
Australia and Oceania	Tuvalu	Baby Food	Offline	H	5/28/2010	669165933	6/27/2010	9925	255.28	159.42	2533654	1582243.5	951410.5
Central America and the Caribbean	Grenada	Cereal	Online	C	8/22/2012	963881480	9/15/2012	2804	205.7	117.11	576782.8	328376.44	248406.36
Europe	Russia	Office Supplies	Offline	L	5/2/2014	341417157	5/8/2014	1779	651.21	524.96	1158502.59	933903.84	224598.75

*/

package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// We will be using this client everywhere in our code
var awsS3Client *s3.Client

func main() {
	configS3()
	s3select()

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

func s3select() {

	params := &s3.SelectObjectContentInput{
		Bucket:         aws.String("csv-bucket"),
		Key:            aws.String("1000000 Sales Records.csv"),
		ExpressionType: types.ExpressionTypeSql,
		// Expression:     aws.String("SELECT * FROM S3Object WHERE Country = Russia LIMIT 5"),
		Expression: aws.String("SELECT Country,\"Units Sold\" FROM S3Object WHERE cast(\"Units Sold\" as int) = 1"),
		// Expression: aws.String("SELECT * FROM S3Object s LIMIT 5"),
		InputSerialization: &types.InputSerialization{
			CSV: &types.CSVInput{
				FileHeaderInfo: types.FileHeaderInfoUse,
				// CompressionType: types.CompressionTypeGzip,
			},
		},
		OutputSerialization: &types.OutputSerialization{
			CSV: &types.CSVOutput{},
		},
	}

	resp, err := awsS3Client.SelectObjectContent(context.TODO(), params)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Printf("%+v\n", resp.ResultMetadata)

	selectStream := resp.GetStream()
	defer selectStream.Close()

	for event := range selectStream.Events() {
		// fmt.Println(event)
		switch v := event.(type) {
		case *types.SelectObjectContentEventStreamMemberRecords:
			//is a byte slice of select records ??
			fmt.Println(string(v.Value.Payload))
		case *types.SelectObjectContentEventStreamMemberStats:
			// s3.StatsEvent contains information on the data that’s processed
			fmt.Println("Processed", v.Value.Details.BytesProcessed, "bytes")
			fmt.Printf("--Stats %+v\n", v.Value.Details)
		case *types.SelectObjectContentEventStreamMemberEnd:
			// s3.EndEvent
			fmt.Println("SelectObjectContent completed")
		default:
			fmt.Printf("selectStream EVENT -------- %+v\n", v)
		}
	}

	if err := selectStream.Err(); err != nil {
		fmt.Errorf("failed to read from SelectObjectContent EventStream, %v", err)
	}
}
