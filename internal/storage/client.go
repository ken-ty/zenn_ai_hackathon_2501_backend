package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type Client struct {
	bucket  *storage.BucketHandle
	baseURL string
}

func NewClient(ctx context.Context, bucketName string) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}

	bucket := client.Bucket(bucketName)
	baseURL := fmt.Sprintf("gs://%s", bucketName)

	return &Client{
		bucket:  bucket,
		baseURL: baseURL,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, path string, content io.Reader) error {
	obj := c.bucket.Object(path)
	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, content); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func (c *Client) GetFile(ctx context.Context, path string) (io.Reader, error) {
	obj := c.bucket.Object(path)
	return obj.NewReader(ctx)
}
