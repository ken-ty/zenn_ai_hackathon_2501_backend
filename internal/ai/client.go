package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	endpoint string
	location string
	project  string
}

func NewClient(project, location string) *Client {
	return &Client{
		endpoint: fmt.Sprintf("%s-aiplatform.googleapis.com:443", location),
		location: location,
		project:  project,
	}
}

func (c *Client) GenerateImage(ctx context.Context, originalImage io.Reader) (io.Reader, error) {
	// 画像をバイト配列に変換
	imageBytes, err := io.ReadAll(originalImage)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}

	// base64エンコード
	base64Input := base64.StdEncoding.EncodeToString(imageBytes)

	// リクエストJSONの構築（インデントを整理）
	requestJSON := fmt.Sprintf(`{
		"instances": [{
			"image": "%s",
			"mode": "variation"
		}]
	}`, base64Input)

	// アクセストークンの取得
	cmd := exec.Command("gcloud", "auth", "print-access-token")
	token, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	// curlコマンドの実行とjqによる処理
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/imagegeneration@006:predict",
		c.location, c.project, c.location)

	curlCmd := fmt.Sprintf(`curl -s -X POST \
		-H "Authorization: Bearer %s" \
		-H "Content-Type: application/json; charset=utf-8" \
		-d @- \
		"%s" | tee /dev/stderr | jq -r '.predictions[0].bytesBase64Encoded' | base64 -d`,
		strings.TrimSpace(string(token)),
		endpoint)

	fmt.Printf("Executing API request...\n")
	curl := exec.Command("bash", "-c", curlCmd)
	curl.Stdin = strings.NewReader(requestJSON)
	curl.Stderr = os.Stderr

	output, err := curl.Output()
	if err != nil {
		fmt.Printf("API request failed\n")
		return nil, fmt.Errorf("failed to execute curl: %v", err)
	}
	fmt.Printf("API request successful\n")

	return bytes.NewReader(output), nil
}
