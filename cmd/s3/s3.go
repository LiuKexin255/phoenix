package main

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// ListBuckets 检索所有桶
func ListBuckets(ctx context.Context, client *s3.Client) ([]types.Bucket, error) {
	resp, err := client.ListBuckets(ctx, nil)
	if err != nil {
		return nil, err
	}
	return resp.Buckets, nil
}

// CreateBucket 创建桶
func CreateBucket(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return err
	}
	return nil
}

// IsExistBucket 查询桶是否存在
func IsExistBucket(ctx context.Context, client *s3.Client, bucket string) (bool, error) {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

// ListObjects 列出桶的所有文件
func ListObjects(ctx context.Context, client *s3.Client, bucket string) ([]types.Object, error) {
	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, err
	}
	return resp.Contents, nil
}

// IsExistObject 查询文件是否存在
func IsExistObject(ctx context.Context, client *s3.Client, bucket string, key string) (bool, error) {
	_, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

// UploadFile 上传文件
func UploadFile(ctx context.Context, client *s3.Client, bucket string, key string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return err
	}
	return nil
}

// DownloadFile 下载文件
func DownloadFile(ctx context.Context, client *s3.Client, bucket string, key string, path string) error {
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, os.ModePerm)
}
