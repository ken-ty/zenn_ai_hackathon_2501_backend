package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"zenn_ai_hackathon_2501_backend/internal/models"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
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

func (c *Client) UpdateQuestions(ctx context.Context, questions models.QuestionsResponse) error {
	obj := c.bucket.Object("metadata/questions.json")
	writer := obj.NewWriter(ctx)

	if err := json.NewEncoder(writer).Encode(questions); err != nil {
		return fmt.Errorf("json.Encode: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

// GenerateSignedURL は、指定されたオブジェクトの署名付きURLを生成します。
// このURLは、オブジェクトの読み取り専用アクセスを許可し、有効期限が15分間です。
func (c *Client) GenerateSignedURL(ctx context.Context, objectPath string) (string, error) {
	log.Printf("Starting to generate signed URL for: %s", objectPath)

	// キーファイルを直接読み込む
	credBytes, err := os.ReadFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		log.Printf("Error reading credentials file: %v", err)
		return "", fmt.Errorf("failed to read credentials: %v", err)
	}
	log.Printf("Successfully read credentials file")

	conf, err := google.JWTConfigFromJSON(credBytes)
	if err != nil {
		log.Printf("Error parsing JWT config: %v", err)
		return "", fmt.Errorf("failed to parse JWT config: %v", err)
	}
	log.Printf("Successfully parsed JWT config, using email: %s", conf.Email)

	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(15 * time.Minute),
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
	}

	url, err := c.bucket.SignedURL(objectPath, opts)
	if err != nil {
		log.Printf("Error generating signed URL: %v", err)
		return "", err
	}
	log.Printf("Successfully generated signed URL for: %s", objectPath)

	return url, nil
}
