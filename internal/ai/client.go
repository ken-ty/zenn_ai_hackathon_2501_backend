package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"

	"golang.org/x/oauth2/google"
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

func (c *Client) GenerateImage(ctx context.Context, prompt string) ([]byte, error) {
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %v", err)
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	// プロンプトをJSONに変換
	jsonStr := fmt.Sprintf(`{
		"instances": [
			{
				"prompt": "%s"
			}
		],
		"parameters": {
			"sampleCount": 1
		}
	}`, prompt)

	// curlコマンドを構築
	endpoint := "https://asia-northeast1-aiplatform.googleapis.com/v1/projects/zenn-ai-hackathon-2501/locations/asia-northeast1/publishers/google/models/imagegeneration:predict"

	cmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: Bearer %s", token.AccessToken),
		"-H", "Content-Type: application/json; charset=utf-8",
		"-d", jsonStr,
		endpoint)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute curl command: %v", err)
	}

	// レスポンスからbase64エンコードされた画像データを抽出
	var response struct {
		Predictions []struct {
			BytesBase64Encoded string `json:"bytesBase64Encoded"`
		} `json:"predictions"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v, response: %s", err, string(output))
	}

	if len(response.Predictions) == 0 {
		return nil, fmt.Errorf("no predictions in response: %s", string(output))
	}

	// base64デコード
	imageBytes, err := base64.StdEncoding.DecodeString(response.Predictions[0].BytesBase64Encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	return imageBytes, nil
}
