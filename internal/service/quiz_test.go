package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
)

// MockAIClient はAIClientのモック
type MockAIClient struct {
	generateInterpretationFunc func(ctx context.Context, imageData []byte, authorInterpretation string) (string, error)
}

func (m *MockAIClient) GenerateInterpretation(ctx context.Context, imageData []byte, authorInterpretation string) (string, error) {
	return m.generateInterpretationFunc(ctx, imageData, authorInterpretation)
}

// MockStorageClient はStorageClientのモック
type MockStorageClient struct {
	saveImageFunc         func(ctx context.Context, imageData []byte) (string, error)
	saveQuizFunc          func(ctx context.Context, quiz *models.Quiz) error
	getQuizFunc           func(ctx context.Context, quizID string) (*models.Quiz, error)
	generateSignedURLFunc func(ctx context.Context, objectPath string) (string, error)
}

func (m *MockStorageClient) SaveImage(ctx context.Context, imageData []byte) (string, error) {
	return m.saveImageFunc(ctx, imageData)
}

func (m *MockStorageClient) SaveQuiz(ctx context.Context, quiz *models.Quiz) error {
	return m.saveQuizFunc(ctx, quiz)
}

func (m *MockStorageClient) GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error) {
	return m.getQuizFunc(ctx, quizID)
}

func (m *MockStorageClient) GenerateSignedURL(ctx context.Context, objectPath string) (string, error) {
	if m.generateSignedURLFunc != nil {
		return m.generateSignedURLFunc(ctx, objectPath)
	}
	return fmt.Sprintf("https://example.com/%s", objectPath), nil
}

func TestCreateQuiz(t *testing.T) {
	tests := []struct {
		name                 string
		imageData            []byte
		authorInterpretation string
		mockAIResponse       string
		mockAIError          error
		mockImagePath        string
		mockImageError       error
		mockSaveQuizError    error
		wantError            bool
	}{
		{
			name:                 "正常系：すべての処理が成功",
			imageData:            []byte("test image"),
			authorInterpretation: "作者の解釈",
			mockAIResponse:       "AIの解釈",
			mockImagePath:        "/images/test.jpg",
			wantError:            false,
		},
		{
			name:                 "異常系：画像データなし",
			imageData:            nil,
			authorInterpretation: "作者の解釈",
			wantError:            true,
		},
		{
			name:                 "異常系：解釈なし",
			imageData:            []byte("test image"),
			authorInterpretation: "",
			wantError:            true,
		},
		{
			name:                 "異常系：画像保存エラー",
			imageData:            []byte("test image"),
			authorInterpretation: "作者の解釈",
			mockImageError:       fmt.Errorf("storage error"),
			wantError:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockAI := &MockAIClient{
				generateInterpretationFunc: func(ctx context.Context, imageData []byte, authorInterpretation string) (string, error) {
					if tt.mockAIError != nil {
						return "", tt.mockAIError
					}
					return tt.mockAIResponse, nil
				},
			}

			mockStorage := &MockStorageClient{
				saveImageFunc: func(ctx context.Context, imageData []byte) (string, error) {
					if tt.mockImageError != nil {
						return "", tt.mockImageError
					}
					return tt.mockImagePath, nil
				},
				saveQuizFunc: func(ctx context.Context, quiz *models.Quiz) error {
					return tt.mockSaveQuizError
				},
			}

			service := NewQuizService(mockAI, mockStorage)

			// テストの実行
			quiz, err := service.CreateQuiz(context.Background(), tt.imageData, tt.authorInterpretation)

			// 結果の検証
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

			if quiz.ImagePath != tt.mockImagePath {
				t.Errorf("expected image path %q, got %q", tt.mockImagePath, quiz.ImagePath)
			}
			if quiz.AuthorInterpretation != tt.authorInterpretation {
				t.Errorf("expected author interpretation %q, got %q", tt.authorInterpretation, quiz.AuthorInterpretation)
			}
			if quiz.AIInterpretation != tt.mockAIResponse {
				t.Errorf("expected AI interpretation %q, got %q", tt.mockAIResponse, quiz.AIInterpretation)
			}
		})
	}
}

func TestGetRandomizedInterpretations(t *testing.T) {
	service := NewQuizService(nil, nil) // 依存関係不要

	quiz := &models.Quiz{
		AuthorInterpretation: "作者の解釈",
		AIInterpretation:     "AIの解釈",
	}

	// 複数回実行して、順序がランダムになることを確認
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		interpretations := service.GetRandomizedInterpretations(quiz)
		if len(interpretations) != 2 {
			t.Errorf("expected 2 interpretations, got %d", len(interpretations))
		}
		order := interpretations[0] + "|" + interpretations[1]
		seen[order] = true
	}

	// 少なくとも2つの異なる順序が観察されることを期待
	if len(seen) < 2 {
		t.Error("randomization doesn't seem to be working")
	}
}

func TestVerifyAnswer(t *testing.T) {
	service := NewQuizService(nil, nil) // 依存関係不要

	quiz := &models.Quiz{
		AuthorInterpretation: "作者の解釈",
		AIInterpretation:     "AIの解釈",
	}

	tests := []struct {
		name                   string
		selectedInterpretation string
		want                   bool
	}{
		{
			name:                   "正解の場合",
			selectedInterpretation: "作者の解釈",
			want:                   true,
		},
		{
			name:                   "不正解の場合",
			selectedInterpretation: "AIの解釈",
			want:                   false,
		},
		{
			name:                   "存在しない解釈の場合",
			selectedInterpretation: "別の解釈",
			want:                   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service.VerifyAnswer(quiz, tt.selectedInterpretation); got != tt.want {
				t.Errorf("VerifyAnswer() = %v, want %v", got, tt.want)
			}
		})
	}
}
