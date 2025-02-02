package config

import (
	"fmt"
	"os"
)

// Config はアプリケーションの設定を保持する構造体
type Config struct {
	ProjectID  string
	Location   string
	BucketName string
	Port       string
}

// Load は環境変数から設定を読み込む
func Load() (*Config, error) {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("PROJECT_ID environment variable is not set")
	}

	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("BUCKET_NAME environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // デフォルトポート
	}

	return &Config{
		ProjectID:  projectID,
		Location:   "us-central1",
		BucketName: bucketName,
		Port:       port,
	}, nil
}

// GetPort はポート番号を:8080の形式で返します
func (c *Config) GetPort() string {
	return ":" + c.Port
}

// Validate は設定値の検証を行います
func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("ProjectID is required")
	}
	if c.BucketName == "" {
		return fmt.Errorf("BucketName is required")
	}
	if c.Location == "" {
		return fmt.Errorf("Location is required")
	}
	if c.Port == "" {
		return fmt.Errorf("Port is required")
	}
	return nil
}
