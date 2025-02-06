package ai

import (
	"context"
	"fmt"
	"io"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
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

func (c *Client) GenerateImage(ctx context.Context, prompt string) (io.Reader, error) {
	client, err := aiplatform.NewPredictionClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("aiplatform.NewPredictionClient: %v", err)
	}
	defer client.Close()

	req := &aiplatformpb.PredictRequest{
		Endpoint: fmt.Sprintf("projects/%s/locations/%s/endpoints/%s",
			c.project, c.location, "YOUR_ENDPOINT_ID"),
		Instances: []*aiplatformpb.Value{
			{
				Value: &aiplatformpb.Value_StringValue{
					StringValue: prompt,
				},
			},
		},
	}

	resp, err := client.Predict(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("client.Predict: %v", err)
	}

	// TODO: 画像データの取得処理を実装
	return nil, fmt.Errorf("not implemented")
}
