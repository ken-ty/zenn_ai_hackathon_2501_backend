package ai

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/vertexai/genai"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/logging"
	"google.golang.org/api/option"
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
	logging.Info("AIクライアントの初期化を開始: projectID=%s, location=%s", projectID, location)
	ctx := context.Background()

	var opts []option.ClientOption
	// ローカル環境の場合はkeyfile.jsonを使用
	if keyPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); keyPath != "" {
		logging.Info("認証情報ファイルを使用: %s", keyPath)
		opts = append(opts, option.WithCredentialsFile(keyPath))
	} else {
		logging.Info("デフォルトの認証情報を使用（Workload Identity）")
	}

	// Vertex AIクライアントの作成
	client, err := genai.NewClient(ctx, projectID, location, opts...)
	if err != nil {
		logging.Error("Vertex AIクライアントの作成に失敗: %v", err)
		return nil, fmt.Errorf("Vertex AIクライアントの作成に失敗: %w", err)
	}
	logging.Info("Vertex AIクライアントの作成に成功")

	model := client.GenerativeModel("gemini-pro-vision")
	model.SetTemperature(0.7)

	return &Client{
		projectID: projectID,
		location:  location,
		model:     model,
	}, nil
}

// generatePrompt はプロンプトを生成します
func generatePrompt(authorInterpretation string) string {
	return fmt.Sprintf(`
この画像に対して、投稿者は以下のような解釈をしています：
%s

この作品に対して、投稿者とは異なる視点から、新しい解釈を生成してください。
以下の点に注意してください：
1. 投稿者の解釈と同じような親しみやすい文体で書く
2. 投稿者の解釈の0.8倍から1.2倍以内の文字数に収める（長すぎないように注意）
4. 投稿者の解釈が自然な会話調なら、同じように自然な会話調で書く
5. 投稿者の解釈がもっともらしいものなら、それを踏襲する
6. 1つの段落にまとめる（改行を入れない）
7. 重複する表現は避ける

投稿者の解釈の文字数は%d文字です。これを参考に、簡潔な解釈を生成してください。
`, authorInterpretation, len(authorInterpretation))
}

// GenerateInterpretation は画像の解釈を生成します
func (c *Client) GenerateInterpretation(ctx context.Context, imageData []byte, authorInterpretation string) (string, error) {
	logging.Info("解釈生成を開始: 画像サイズ=%d bytes", len(imageData))
	if len(imageData) == 0 {
		logging.Error("画像データが空です")
		return "", fmt.Errorf("画像データが必要です")
	}
	if authorInterpretation == "" {
		logging.Error("投稿者の解釈が空です")
		return "", fmt.Errorf("投稿者の解釈が必要です")
	}

	prompt := generatePrompt(authorInterpretation)
	logging.Debug("プロンプトを生成: 長さ=%d文字", len(prompt))

	response, err := c.model.GenerateContent(ctx,
		genai.ImageData("image/jpeg", imageData),
		genai.Text(prompt),
	)
	if err != nil {
		logging.Error("AIからの応答の取得に失敗: %v", err)
		return "", fmt.Errorf("AIからの応答の取得に失敗: %w", err)
	}

	if len(response.Candidates) == 0 {
		logging.Error("AIからの応答が空です")
		return "", fmt.Errorf("AIからの応答が空です")
	}

	text, ok := response.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		logging.Error("応答のテキスト変換に失敗")
		return "", fmt.Errorf("テキスト応答の解析に失敗")
	}

	interpretation := string(text)
	logging.Info("解釈の生成に成功: 長さ=%d文字", len(interpretation))
	return interpretation, nil
}
