package ai

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

type Client struct {
	endpoint string
	location string
	project  string
}

// リクエスト用の構造体を追加
type generateImageRequest struct {
	Instances []struct {
		Prompt string `json:"prompt"`
	} `json:"instances"`
	Parameters struct {
		SampleCount int    `json:"sampleCount"`
		AspectRatio string `json:"aspectRatio"`
	} `json:"parameters"`
}

func NewClient(project, location string) *Client {
	return &Client{
		endpoint: fmt.Sprintf("%s-aiplatform.googleapis.com:443", location),
		location: location,
		project:  project,
	}
}

func (c *Client) GenerateImage(ctx context.Context, prompt string) ([]byte, error) {
	// ランダムにフィクスチャーファイルを選択
	fileNum := rand.Intn(3)
	filename := fmt.Sprintf("fake%d.png", fileNum)

	// フィクスチャーファイルのパスを構築
	fixturePath := filepath.Join("fixture", "generated", filename)

	// ファイルを読み込む
	imageBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		return nil, fmt.Errorf("フィクスチャーファイルの読み込みに失敗: %w", err)
	}

	return imageBytes, nil
}
