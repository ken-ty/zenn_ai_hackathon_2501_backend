package ai

import (
	"context"
	"testing"

	"cloud.google.com/go/vertexai/genai"
)

// MockGenerativeModel はgenai.GenerativeModelのモック
type MockGenerativeModel struct {
	generateContentFunc func(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

func (m *MockGenerativeModel) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return m.generateContentFunc(ctx, parts...)
}

func (m *MockGenerativeModel) SetTemperature(float32) {}

func TestGenerateInterpretation(t *testing.T) {
	// テスト用のダミー画像データ
	imageData := []byte("test image data")

	// テストケース
	tests := []struct {
		name                 string
		imageData            []byte
		authorInterpretation string
		mockResponse         *genai.GenerateContentResponse
		wantError            bool
		wantInterpretation   string
	}{
		{
			name:                 "正常系：有効な応答",
			imageData:            imageData,
			authorInterpretation: "この作品では、都市の無機質な表面に映る自然光の反射を通じて、現代社会における人工と自然の共生を表現しました。",
			mockResponse: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("AIによる解釈"),
							},
						},
					},
				},
			},
			wantError:          false,
			wantInterpretation: "AIによる解釈",
		},
		{
			name:                 "異常系：空の応答",
			imageData:            imageData,
			authorInterpretation: "この作品では、都市の無機質な表面に映る自然光の反射を通じて、現代社会における人工と自然の共生を表現しました。",
			mockResponse: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{},
			},
			wantError:          true,
			wantInterpretation: "",
		},
		{
			name:                 "異常系：画像なし",
			imageData:            nil,
			authorInterpretation: "テスト解釈",
			mockResponse:         nil,
			wantError:            true,
			wantInterpretation:   "",
		},
		{
			name:                 "異常系：解釈なし",
			imageData:            imageData,
			authorInterpretation: "",
			mockResponse:         nil,
			wantError:            true,
			wantInterpretation:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockModel := &MockGenerativeModel{
				generateContentFunc: func(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
					return tt.mockResponse, nil
				},
			}

			client := &Client{
				projectID: "test-project",
				location:  "us-central1",
				model:     mockModel,
			}

			// テストの実行
			got, err := client.GenerateInterpretation(context.Background(), tt.imageData, tt.authorInterpretation)

			// エラーの検証
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// 結果の検証
			if got != tt.wantInterpretation {
				t.Errorf("want %q, got %q", tt.wantInterpretation, got)
			}
		})
	}
}
