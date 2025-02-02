package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
)

// MockStorageClient はStorageClientのモック
type MockStorageClient struct {
	mock.Mock
}

func (m *MockStorageClient) SaveImage(ctx context.Context, imageData []byte) (string, error) {
	args := m.Called(ctx, imageData)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) SaveQuiz(ctx context.Context, quiz *models.Quiz) error {
	args := m.Called(ctx, quiz)
	return args.Error(0)
}

func (m *MockStorageClient) GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error) {
	args := m.Called(ctx, quizID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Quiz), args.Error(1)
}

func (m *MockStorageClient) GenerateSignedURL(ctx context.Context, objectPath string) (string, error) {
	args := m.Called(ctx, objectPath)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) GetQuizzes(ctx context.Context) ([]*models.Quiz, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Quiz), args.Error(1)
}

func (m *MockStorageClient) DeleteAllQuizzes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockAIClient はAIClientのモック
type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) GenerateInterpretation(ctx context.Context, imageData []byte, authorInterpretation string) (string, error) {
	args := m.Called(ctx, imageData, authorInterpretation)
	return args.String(0), args.Error(1)
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
			authorInterpretation: "投稿者の解釈",
			mockAIResponse:       "AIの解釈",
			mockImagePath:        "/images/test.jpg",
			wantError:            false,
		},
		{
			name:                 "異常系：画像データなし",
			imageData:            nil,
			authorInterpretation: "投稿者の解釈",
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
			authorInterpretation: "投稿者の解釈",
			mockImageError:       fmt.Errorf("storage error"),
			wantError:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockAI := &MockAIClient{}
			mockAI.On("GenerateInterpretation", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockAIResponse, tt.mockAIError)

			mockStorage := &MockStorageClient{}
			mockStorage.On("SaveImage", mock.Anything, mock.Anything).Return(tt.mockImagePath, tt.mockImageError)
			mockStorage.On("SaveQuiz", mock.Anything, mock.Anything).Return(tt.mockSaveQuizError)
			mockStorage.On("GetQuizzes", mock.Anything).Return([]*models.Quiz{}, nil)

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
		AuthorInterpretation: "投稿者の解釈",
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
		AuthorInterpretation: "投稿者の解釈",
		AIInterpretation:     "AIの解釈",
	}

	tests := []struct {
		name                   string
		selectedInterpretation string
		want                   bool
	}{
		{
			name:                   "正解の場合",
			selectedInterpretation: "投稿者の解釈",
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

func TestDeleteAllQuizzes(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockStorageClient)
		wantErr bool
	}{
		{
			name: "正常系：全クイズの削除に成功",
			setup: func(m *MockStorageClient) {
				m.On("DeleteAllQuizzes", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "異常系：ストレージクライアントでエラー発生",
			setup: func(m *MockStorageClient) {
				m.On("DeleteAllQuizzes", mock.Anything).Return(fmt.Errorf("storage error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorageClient{}
			tt.setup(mockStorage)

			service := &QuizServiceImpl{
				storageClient: mockStorage,
			}

			err := service.DeleteAllQuizzes(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
