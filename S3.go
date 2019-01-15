package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"os"
	"path"
	"strconv"
)

func _listFilesInS3Bucket(bucketName, prefix string) *s3.ListObjectsOutput {
	svc := s3.New(_getSession(), _getAwsConfig())

	var marker = new(string)
	var result = new(s3.ListObjectsOutput)

	for marker != nil && len(result.Contents) < int(maxResult) {

		if marker != nil && *marker == "" {
			marker = nil
		}

		params := &s3.ListObjectsInput{
			Bucket:  aws.String(bucketName),
			MaxKeys: &maxResult,
			Marker:marker,
		}

		if prefix != "" {
			params.Prefix = aws.String(prefix)
		}

		resp, err := svc.ListObjects(params)
		assertError(err)

		result.Contents = append(resp.Contents)

		if verbose {
			fmt.Println("Fetched", len(result.Contents), "items.")
		}

		marker = resp.NextMarker
	}

	return result
}

func _listS3Buckets() *s3.ListBucketsOutput {
	svc := s3.New(_getSession(), _getAwsConfig())

	params := &s3.ListBucketsInput{

	}

	resp, err := svc.ListBuckets(params)
	assertError(err)
	return resp
}

func _copyFileFromS3Bucket(bucket, filename string) *s3.GetObjectOutput {
	svc := s3.New(_getSession(), _getAwsConfig())

	params := &s3.GetObjectInput{
		Key:    aws.String(filename),
		Bucket: aws.String(bucket),
	}

	resp, err := svc.GetObject(params)
	assertError(err)
	return resp
}

func ListS3Buckets() {
	buckets := _listS3Buckets()

	for i := 0; i < len(buckets.Buckets); i++ {
		bucket := buckets.Buckets[i]
		if verboseLevel > 0 {
			fmt.Printf("%s   %s\n", bucket.CreationDate.Format("2006-01-02 15:04:05 -0700 MST"), *bucket.Name)
		} else {
			fmt.Println(*bucket.Name)
		}
	}
}

func ListFilesInS3Bucket(bucketName, prefix string) {
	files := _listFilesInS3Bucket(bucketName, prefix)

	for i := 0; i < len(files.Contents); i++ {
		file := files.Contents[i]

		if verboseLevel == 1 {
			// https://writer-lambda-releases.s3.amazonaws.com/ImageMetadata-develop.zip
			fmt.Printf("https://%s.s3.amazonaws.com/%s\n", bucketName, *file.Key)
		} else if verboseLevel == 2 {
			// -rw-r--r--  1 tobias  staff      1392 Sep 28 15:21 Clusters.go
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

	out, err := os.Create(outFile)

	assertError(err)
	defer result.Body.Close()

	fmt.Printf("Copying %s/%s to %s... ", bucketName, filename, output)
	written, err := io.Copy(out, result.Body)
	assertError(err)
	fmt.Println("Done writing " + strconv.FormatInt(written, 10) + " bytes")
}
