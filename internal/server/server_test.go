package server

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
)

// MockQuizService はQuizServiceのモック
type MockQuizService struct {
	createQuizFunc                   func(ctx context.Context, imageData []byte, authorInterpretation string) (*models.Quiz, error)
	getQuizFunc                      func(ctx context.Context, quizID string) (*models.Quiz, error)
	getRandomizedInterpretationsFunc func(quiz *models.Quiz) []string
	verifyAnswerFunc                 func(quiz *models.Quiz, selectedInterpretation string) bool
}

func (m *MockQuizService) CreateQuiz(ctx context.Context, imageData []byte, authorInterpretation string) (*models.Quiz, error) {
	return m.createQuizFunc(ctx, imageData, authorInterpretation)
}

func (m *MockQuizService) GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error) {
	return m.getQuizFunc(ctx, quizID)
}

func (m *MockQuizService) GetRandomizedInterpretations(quiz *models.Quiz) []string {
	return m.getRandomizedInterpretationsFunc(quiz)
}

func (m *MockQuizService) VerifyAnswer(quiz *models.Quiz, selectedInterpretation string) bool {
	if m.verifyAnswerFunc != nil {
		return m.verifyAnswerFunc(quiz, selectedInterpretation)
	}
	return selectedInterpretation == quiz.AuthorInterpretation
}

func TestHandleUpload(t *testing.T) {
	// モックの設定
	mockService := &MockQuizService{
		createQuizFunc: func(ctx context.Context, imageData []byte, authorInterpretation string) (*models.Quiz, error) {
			return &models.Quiz{
				ID:                   "test-quiz",
				ImagePath:            "/images/test.jpg",
				AuthorInterpretation: authorInterpretation,
				AIInterpretation:     "AIの解釈",
				CreatedAt:            time.Now(),
			}, nil
		},
	}

	srv := NewServer(mockService)

	// マルチパートフォームの作成
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fw.Write([]byte("test image data"))
	if err != nil {
		t.Fatal(err)
	}
	err = w.WriteField("interpretation", "作者の解釈")
	if err != nil {
		t.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}

	// リクエストの作成
	req := httptest.NewRequest("POST", "/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()

	// ハンドラーの実行
	srv.handleUpload(rec, req)

	// レスポンスの検証
	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var quiz models.Quiz
	if err := json.NewDecoder(rec.Body).Decode(&quiz); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if quiz.ID != "test-quiz" {
		t.Errorf("expected quiz ID %q, got %q", "test-quiz", quiz.ID)
	}
}

func TestHandleGetQuiz(t *testing.T) {
	// モックの設定
	mockService := &MockQuizService{
		getQuizFunc: func(ctx context.Context, quizID string) (*models.Quiz, error) {
			return &models.Quiz{
				ID:                   quizID,
				ImagePath:            "/images/test.jpg",
				AuthorInterpretation: "作者の解釈",
				AIInterpretation:     "AIの解釈",
				CreatedAt:            time.Now(),
			}, nil
		},
		getRandomizedInterpretationsFunc: func(quiz *models.Quiz) []string {
			return []string{"作者の解釈", "AIの解釈"}
		},
	}

	srv := NewServer(mockService)

	// リクエストの作成
	req := httptest.NewRequest("GET", "/quiz/test-quiz", nil)
	rec := httptest.NewRecorder()

	// ハンドラーの実行
	srv.handleGetQuiz(rec, req)

	// レスポンスの検証
	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		ID              string   `json:"id"`
		ImagePath       string   `json:"image_path"`
		Interpretations []string `json:"interpretations"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if response.ID != "test-quiz" {
		t.Errorf("expected quiz ID %q, got %q", "test-quiz", response.ID)
	}
	if len(response.Interpretations) != 2 {
		t.Errorf("expected 2 interpretations, got %d", len(response.Interpretations))
	}
}
