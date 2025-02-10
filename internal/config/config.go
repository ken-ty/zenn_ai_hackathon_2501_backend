package config

import (
	"fmt"
	"os"
)

// Config はアプリケーションの設定を保持します
type Config struct {
	ProjectID  string
	BucketName string
	Location   string
	Port       string
	Debug      bool
}

// Load は環境変数から設定を読み込みます
func Load() (*Config, error) {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("PROJECT_ID is required")
	}

	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("BUCKET_NAME is required")
	}

	location := os.Getenv("LOCATION")
	if location == "" {
		location = "us-central1"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	debug := os.Getenv("DEBUG") == "true"

	return &Config{
		ProjectID:  projectID,
		BucketName: bucketName,
		Location:   location,
		Port:       port,
		Debug:      debug,
	}, nil
}

// GetPort はポート番号を返します
func (c *Config) GetPort() string {
	return fmt.Sprintf(":%s", c.Port)
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
