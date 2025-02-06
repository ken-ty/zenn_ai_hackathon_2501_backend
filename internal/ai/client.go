package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
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
	client, err := aiplatform.NewPredictionClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("aiplatform.NewPredictionClient: %v", err)
	}
	defer client.Close()

	// 画像をバイト配列に変換
	imageBytes, err := io.ReadAll(originalImage)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}

	instance := &structpb.Value{
		Kind: &structpb.Value_StringValue{
			Fields: map[string]*structpb.Value{
				"image": {
					Kind: &structpb.Value_BytesValue{
						BytesValue: imageBytes,
					},
				},
				"mode": {
					Kind: &structpb.Value_StringValue{
						StringValue: "variation",
					},
				},
			},
		},
	}

	req := &aiplatformpb.PredictRequest{
		Endpoint: fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/imagegeneration",
			c.project, c.location, "YOUR_ENDPOINT_ID"),
		Instances: []*structpb.Value{instance},
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

	generatedImage := predictions[0].GetStructValue().GetFields()["image"].GetBytesValue()
	return bytes.NewReader(generatedImage), nil
}
