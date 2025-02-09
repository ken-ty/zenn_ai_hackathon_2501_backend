package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
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
	// 認証情報の確認を追加
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		fmt.Printf("Failed to get credentials: %+v\n", err)
		return nil, fmt.Errorf("failed to get credentials: %v", err)
	}
	fmt.Printf("Credentials found with quota project: %s\n", creds.ProjectID)

	client, err := aiplatform.NewPredictionClient(ctx,
		option.WithEndpoint(c.endpoint),
		option.WithCredentials(creds)) // 明示的に認証情報を指定
	if err != nil {
		fmt.Printf("Failed to create client: %+v\n", err)
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

	// エンドポイントパスを定義
	endpointPath := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/imagegeneration@latest",
		c.project, c.location)

	fmt.Printf("=== Connection Info ===\n")
	fmt.Printf("Client endpoint: %s\n", c.endpoint)
	fmt.Printf("API endpoint: %s\n", endpointPath)
	fmt.Printf("Project: %s\n", c.project)
	fmt.Printf("Location: %s\n", c.location)

	// インスタンスの構造を構築する前にデバッグ出力
	fmt.Printf("\n=== Input Info ===\n")
	fmt.Printf("Image size (base64): %d bytes\n", len(base64Input))
	fmt.Printf("Mode: variation\n")

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
		Endpoint:  endpointPath,
		Instances: []*structpb.Value{imageValue},
	}

	// リクエストの詳細を表示
	fmt.Printf("\n=== Request Details ===\n")
	fmt.Printf("Request endpoint: %s\n", req.Endpoint)
	fmt.Printf("Request instances: %d\n", len(req.Instances))

	fmt.Printf("\n=== Making API Call ===\n")
	fmt.Printf("Request instances count: %d\n", len(req.Instances))
	resp, err := client.Predict(ctx, req)
	if err != nil {
		fmt.Printf("=== Error Details ===\n")
		fmt.Printf("Error type: %T\n", err)
		if st, ok := status.FromError(err); ok {
			fmt.Printf("Status code: %v\n", st.Code())
			fmt.Printf("Status message: %v\n", st.Message())
		}
		fmt.Printf("Full error: %+v\n", err)
		return nil, fmt.Errorf("client.Predict: %v", err)
	}

	fmt.Printf("=== Response Info ===\n")
	predictions := resp.GetPredictions()
	if len(predictions) == 0 {
		return nil, fmt.Errorf("no predictions returned")
	}

	base64Image := predictions[0].GetStructValue().GetFields()["image"].GetStringValue()
	fmt.Printf("Response image length: %d\n", len(base64Image))

	generatedImage, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %v", err)
	}
	fmt.Printf("Decoded image length: %d\n", len(generatedImage))

	return bytes.NewReader(generatedImage), nil
}
