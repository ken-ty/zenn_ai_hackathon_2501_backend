package ai

import (
	"context"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
)

// GenerativeModel はAIモデルのインターフェース
type GenerativeModel interface {
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// AIClient はAIサービスとの通信を抽象化するインターフェース
type AIClient interface {
	GenerateInterpretation(ctx context.Context, imageData []byte, authorInterpretation string) (string, error)
}

// Client はVertex AIとの通信を担当します
type Client struct {
	projectID string
	location  string
	model     GenerativeModel
}

// NewClient は新しいAIクライアントを作成します
func NewClient(projectID, location string) (*Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("Vertex AIクライアントの作成に失敗: %w", err)
	}

	model := client.GenerativeModel("gemini-pro-vision")
	return &Client{
		projectID: projectID,
		location:  location,
		model:     model,
	}, nil
}

// GenerateInterpretation は画像と作者の解釈から新しい解釈を生成します
func (c *Client) GenerateInterpretation(ctx context.Context, imageData []byte, authorInterpretation string) (string, error) {
	prompt := fmt.Sprintf(`
この画像に対して、作者は以下のような解釈をしています：
%s

この作品に対して、作者とは異なる視点から、説得力のある新しい解釈を生成してください。
以下の点に注意してください：
1. 作者の解釈とは明確に異なる視点を提供する
2. 画像の具体的な要素に基づいて解釈する
3. 芸術的・文化的な文脈を考慮する
4. 300文字程度で簡潔に説明する
`, authorInterpretation)

	resp, err := c.model.GenerateContent(ctx,
		genai.ImageData("jpeg", imageData),
		genai.Text(prompt),
	)
	if err != nil {
		return "", fmt.Errorf("解釈の生成に失敗: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("有効な応答が得られませんでした")
	}

	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("テキスト応答の解析に失敗")
	}

	return string(text), nil
}
