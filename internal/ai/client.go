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
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("認証情報の取得に失敗: %w", err)
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("トークンの取得に失敗: %w", err)
	}

	// リクエストの構築
	req := generateImageRequest{
		Instances: []struct {
			Prompt string `json:"prompt"`
		}{{Prompt: prompt}},
		Parameters: struct {
			SampleCount int    `json:"sampleCount"`
			AspectRatio string `json:"aspectRatio"`
		}{
			SampleCount: 1,
			AspectRatio: "1:1",
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("JSONの生成に失敗: %w", err)
	}

	// エンドポイントの構築
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/imagegeneration@006:predict",
		c.location, c.project, c.location)

	// curlコマンドの実行
	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization:Bearer %s", token.AccessToken),
		"-H", "Content-Type:application/json;charset=utf-8",
		"-d", string(jsonData),
		endpoint)

	// デバッグ用にコマンドとJSONデータを出力
	fmt.Printf("実行するコマンド: %s\n", cmd.String())
	fmt.Printf("JSONデータ: %s\n", string(jsonData))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("curlコマンドの実行に失敗: %w\n出力: %s", err, string(output))
	}

	// レスポンスの解析
	var response struct {
		Predictions []struct {
			BytesBase64Encoded string `json:"bytesBase64Encoded"`
		} `json:"predictions"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("レスポンスの解析に失敗: %w\nレスポンス: %s", err, string(output))
	}

	if len(response.Predictions) == 0 {
		return nil, fmt.Errorf("予測結果が空です: %s", string(output))
	}

	// base64デコード
	imageBytes, err := base64.StdEncoding.DecodeString(response.Predictions[0].BytesBase64Encoded)
	if err != nil {
		return nil, fmt.Errorf("base64デコードに失敗: %w", err)
	}

	return imageBytes, nil
}
