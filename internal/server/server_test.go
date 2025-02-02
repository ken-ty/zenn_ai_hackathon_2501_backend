package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/logging"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
)

func init() {
	// テスト時はERRORレベルのみ出力
	logging.SetLevel(logging.ERROR)
}

// MockQuizService はQuizServiceのモック
type MockQuizService struct {
	mock.Mock
}

func (m *MockQuizService) CreateQuiz(ctx context.Context, imageData []byte, authorInterpretation string) (*models.Quiz, error) {
	args := m.Called(ctx, imageData, authorInterpretation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Quiz), args.Error(1)
}

func (m *MockQuizService) GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error) {
	args := m.Called(ctx, quizID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Quiz), args.Error(1)
}

func (m *MockQuizService) GetRandomizedInterpretations(quiz *models.Quiz) []string {
	args := m.Called(quiz)
	return args.Get(0).([]string)
}

func (m *MockQuizService) VerifyAnswer(quiz *models.Quiz, selectedInterpretation string) bool {
	args := m.Called(quiz, selectedInterpretation)
	return args.Bool(0)
}

func (m *MockQuizService) GetSignedImageURL(ctx context.Context, imagePath string) (string, error) {
	args := m.Called(ctx, imagePath)
	return args.String(0), args.Error(1)
}

func (m *MockQuizService) GetQuizList(ctx context.Context) ([]*models.Quiz, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Quiz), args.Error(1)
}

func (m *MockQuizService) DeleteAllQuizzes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHandleUpload(t *testing.T) {
	// モックのレスポンスを作成
	mockQuiz := &models.Quiz{
		ID:                   "test-quiz",
		ImagePath:            "test-image.jpg",
		CreatedAt:            time.Now().UTC(),
		AuthorInterpretation: "投稿者の解釈",
		AIInterpretation:     "AIの解釈",
	}

	// モックのサービスを設定
	mockService := &MockQuizService{}
	mockService.On("CreateQuiz", mock.Anything, mock.Anything, mock.Anything).Return(mockQuiz, nil)
	mockService.On("GetSignedImageURL", mock.Anything, mockQuiz.ImagePath).Return("https://storage.example.com/test-image.jpg", nil)

	// ハンドラーを作成
	handler := NewServer(mockService)

	// テストリクエストを作成
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// 画像ファイルを追加
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("フォームファイルの作成に失敗: %v", err)
	}

	// 最小限の有効なJPEGファイル（1x1ピクセル、グレースケール）
	jpegData := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, 0x00, 0x10, // APP0 segment
		0x4A, 0x46, 0x49, 0x46, 0x00, // JFIF identifier
		0x01, 0x01, // version
		0x00,       // units
		0x00, 0x01, // X density
		0x00, 0x01, // Y density
		0x00, 0x00, // thumbnail
		0xFF, 0xDB, 0x00, 0x43, // DQT
		0x00, // table 0, precision 0
		0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07,
		0x07, 0x07, 0x09, 0x09, 0x08, 0x0A, 0x0C, 0x14,
		0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12, 0x13,
		0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A,
		0x1C, 0x1C, 0x20, 0x24, 0x2E, 0x27, 0x20, 0x22,
		0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29, 0x2C,
		0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39,
		0x3D, 0x38, 0x32, 0x3C, 0x2E, 0x33, 0x34, 0x32,
		0xFF, 0xC0, 0x00, 0x0B, // SOF0
		0x08,       // precision
		0x00, 0x01, // height
		0x00, 0x01, // width
		0x01,             // number of components
		0x01, 0x11, 0x00, // parameters
		0xFF, 0xC4, 0x00, 0x14, // DHT
		0x00, // table 0
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00,
		0xFF, 0xDA, 0x00, 0x08, // SOS
		0x01, 0x01, 0x00, 0x00, 0x3F, 0x00,
		0xFF, 0xD9, // EOI
	}
	part.Write(jpegData)

	// 解釈を追加
	writer.WriteField("interpretation", "投稿者の解釈")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	handler.handleUpload(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期待するステータスコード %d に対して、%d が返されました", http.StatusOK, rec.Code)
	}

	var response models.QuizResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("レスポンスのデコードに失敗: %v", err)
	}

	if response.ID != mockQuiz.ID {
		t.Errorf("期待するクイズID %q に対して、%q が返されました", mockQuiz.ID, response.ID)
	}
	if response.AuthorInterpretation != mockQuiz.AuthorInterpretation {
		t.Errorf("期待する投稿者の解釈 %q に対して、%q が返されました", mockQuiz.AuthorInterpretation, response.AuthorInterpretation)
	}
	if response.AIInterpretation != mockQuiz.AIInterpretation {
		t.Errorf("期待するAIの解釈 %q に対して、%q が返されました", mockQuiz.AIInterpretation, response.AIInterpretation)
	}
}

func TestHandleGetQuiz(t *testing.T) {
	// モックの設定
	mockService := &MockQuizService{}
	mockService.On("GetQuiz", mock.Anything, "test-quiz").Return(&models.Quiz{
		ID:                   "test-quiz",
		ImagePath:            "/images/test.jpg",
		AuthorInterpretation: "投稿者の解釈",
		AIInterpretation:     "AIの解釈",
		CreatedAt:            time.Now(),
	}, nil)
	mockService.On("GetRandomizedInterpretations", mock.Anything).Return([]string{"投稿者の解釈", "AIの解釈"})
	mockService.On("GetSignedImageURL", mock.Anything, "/images/test.jpg").Return("https://example.com/test.jpg", nil)

	srv := NewServer(mockService)

	// リクエストの作成
	req := httptest.NewRequest("GET", "/quizzes/test-quiz", nil)
	rec := httptest.NewRecorder()

	// ハンドラーの実行
	srv.handleGetQuiz(rec, req)

	// レスポンスの検証
	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		ID                   string `json:"id"`
		ImageURL             string `json:"image_url"`
		CreatedAt            string `json:"created_at"`
		AuthorInterpretation string `json:"author_interpretation"`
		AIInterpretation     string `json:"ai_interpretation"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("レスポンスのデコードに失敗: %v", err)
	}

	if response.ID != "test-quiz" {
		t.Errorf("期待するクイズID %q に対して、%q が返されました", "test-quiz", response.ID)
	}
	if response.AuthorInterpretation != "投稿者の解釈" {
		t.Errorf("期待する投稿者の解釈 %q に対して、%q が返されました", "投稿者の解釈", response.AuthorInterpretation)
	}
	if response.AIInterpretation != "AIの解釈" {
		t.Errorf("期待するAIの解釈 %q に対して、%q が返されました", "AIの解釈", response.AIInterpretation)
	}
}

func TestHandleDeleteAllQuizzes(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		setup        func(*MockQuizService)
		expectedCode int
		expectedBody string
	}{
		{
			name:   "正常系：全クイズの削除に成功",
			method: http.MethodDelete,
			setup: func(m *MockQuizService) {
				m.On("DeleteAllQuizzes", mock.Anything).Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"全てのクイズを削除しました"}`,
		},
		{
			name:         "異常系：不正なHTTPメソッド",
			method:       http.MethodGet,
			setup:        func(m *MockQuizService) {},
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "異常系：サービス層でエラー発生",
			method: http.MethodDelete,
			setup: func(m *MockQuizService) {
				m.On("DeleteAllQuizzes", mock.Anything).Return(fmt.Errorf("service error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockQuizService{}
			tt.setup(mockService)

			server := NewServer(mockService)
			req := httptest.NewRequest(tt.method, "/delete-all-quizzes", nil)
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
