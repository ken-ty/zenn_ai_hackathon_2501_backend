package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	endpoint string
	location string
	project  string
}

func NewClient(project, location string) *Client {
	return &Client{
		endpoint: "us-east1-aiplatform.googleapis.com:443",
		location: location,
		project:  project,
	}
}

func (c *Client) GenerateImage(ctx context.Context, originalImage io.Reader) (io.Reader, error) {
	client, err := aiplatform.NewPredictionClient(ctx,
		option.WithEndpoint(c.endpoint))
	if err != nil {
		return nil, fmt.Errorf("aiplatform.NewPredictionClient: %v", err)
	}
	defer client.Close()

	// 画像をバイト配列に変換
	imageBytes, err := io.ReadAll(originalImage)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}

	// バイナリデータをbase64エンコード
	base64Input := base64.StdEncoding.EncodeToString(imageBytes)

	// インスタンスの構造を正しく構築
	imageValue, err := structpb.NewValue(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"image": {
				Kind: &structpb.Value_StringValue{
					StringValue: base64Input,
				},
			},
			"mode": {
				Kind: &structpb.Value_StringValue{
					StringValue: "variation",
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("structpb.NewValue: %v", err)
	}

	req := &aiplatformpb.PredictRequest{
		// エンドポイントのフォーマット文字列を修正
		Endpoint: fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/imagegeneration",
			c.project, c.location),
		Instances: []*structpb.Value{imageValue},
	}

	resp, err := client.Predict(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("client.Predict: %v", err)
	}

	// レスポンスから画像データを取得
	predictions := resp.GetPredictions()
	if len(predictions) == 0 {
		return nil, fmt.Errorf("no predictions returned")
	}

	// base64エンコードされた文字列を取得し、デコード
	base64Image := predictions[0].GetStructValue().GetFields()["image"].GetStringValue()
	generatedImage, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %v", err)
	}

	return bytes.NewReader(generatedImage), nil
}
