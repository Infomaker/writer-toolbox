package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"strconv"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"os"
	"path"
)

func _listFilesInS3Bucket(bucketName, prefix string) *s3.ListObjectsOutput {
	svc := s3.New(session.New(), _getAwsConfig())

	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}

	if prefix != "" {
		params.Prefix = aws.String(prefix)
	}

	resp, err := svc.ListObjects(params);

	if err != nil {
		errState(err.Error())
	}

	return resp;
}

func _listS3Buckets() *s3.ListBucketsOutput {
	svc := s3.New(session.New(), _getAwsConfig())

	params := &s3.ListBucketsInput{

	}

	resp, err := svc.ListBuckets(params);

	if err != nil {
		errState(err.Error())
	}

	return resp
}


func _copyFileFromS3Bucket(bucket, filename string) *s3.GetObjectOutput {
	svc := s3.New(session.New(), _getAwsConfig())

	params := &s3.GetObjectInput{
		Key: aws.String(filename),
		Bucket: aws.String(bucket),
	}


	resp, err := svc.GetObject(params)
	if err != nil {
		errState(err.Error())
	}

	return resp;
}

func ListS3Buckets() {
	buckets := _listS3Buckets()

	for i := 0; i < len(buckets.Buckets); i++ {
		bucket := buckets.Buckets[i];

		if (verboseLevel > 0) {
			fmt.Printf("%s   %s\n", bucket.CreationDate.Format("2006-01-02 15:04:05 -0700 MST"), *bucket.Name)
		} else {
			fmt.Println(*bucket.Name)
		}
	}
}

func ListFilesInS3Bucket(bucketName, prefix string) {
	files := _listFilesInS3Bucket(bucketName, prefix)

	// -rw-r--r--  1 tobias  staff      1392 Sep 28 15:21 Clusters.go

	for i := 0; i < len(files.Contents); i++ {
		file := files.Contents[i]
		if (verboseLevel > 0) {
			fmt.Printf("%s  %16.d %s %s\n", *file.Owner.DisplayName, *file.Size, file.LastModified.Format("2006-01-02 15:04:05-0700"), *file.Key)
		} else {
			fmt.Println(*file.Key)
		}

	}
}

func CopyFileFromS3Bucket(bucketName, filename, output string) {

	mode := GetFileMode(output)

	if !mode.IsDir() {
		errUsage("Output '" + output + "' must be directory")
	}

	result := _copyFileFromS3Bucket(bucketName, filename)

	outFile := path.Join(output, filename)

	out, err :=os.Create(outFile)

	if err != nil {
		errState(err.Error())
	}


	defer result.Body.Close()

	fmt.Printf("Copying %s/%s to %s... ", bucketName, filename, output)
	written, err := io.Copy(out, result.Body)
	if err != nil {
		errState(err.Error())
	}
	fmt.Println("Done writing " + strconv.FormatInt(written, 10) + " bytes")
}