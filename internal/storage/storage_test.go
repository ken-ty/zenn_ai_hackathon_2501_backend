package storage

import (
	"context"
	"io"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
)

// MockBucket はCloud Storageのモック
type MockBucket struct {
	objects map[string][]byte
}

// NewMockBucket はテスト用のモックバケットを作成します
func NewMockBucket() *MockBucket {
	return &MockBucket{
		objects: make(map[string][]byte),
	}
}

// Object はオブジェクトへの参照を返します
func (b *MockBucket) Object(name string) ObjectHandle {
	return &MockObject{
		name:   name,
		bucket: b,
	}
}

// SignedURL は署名付きURLを生成します
func (b *MockBucket) SignedURL(name string, opts *storage.SignedURLOptions) (string, error) {
	return "https://example.com/" + name, nil
}

// MockObject はCloud Storage Objectのモック
type MockObject struct {
	name   string
	bucket *MockBucket
}

// NewWriter は新しいWriterを返します
func (o *MockObject) NewWriter(ctx context.Context) io.WriteCloser {
	return &MockWriter{
		name:   o.name,
		bucket: o.bucket,
	}
}

// NewReader は新しいReaderを返します
func (o *MockObject) NewReader(ctx context.Context) (io.ReadCloser, error) {
	data, ok := o.bucket.objects[o.name]
	if !ok {
		return nil, storage.ErrObjectNotExist
	}
	return &MockReader{data: data}, nil
}

// MockWriter はCloud Storage Object Writerのモック
type MockWriter struct {
	name   string
	bucket *MockBucket
	data   []byte
}

func (w *MockWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *MockWriter) Close() error {
	w.bucket.objects[w.name] = w.data
	return nil
}

// MockReader はCloud Storage Object Readerのモック
type MockReader struct {
	data   []byte
	offset int64
}

func (r *MockReader) Read(p []byte) (n int, err error) {
	if r.offset >= int64(len(r.data)) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.offset:])
	r.offset += int64(n)
	return n, nil
}

func (r *MockReader) Close() error {
	return nil
}

func TestSaveImage(t *testing.T) {
	tests := []struct {
		name      string
		imageData []byte
		wantError bool
	}{
		{
			name:      "正常系：有効な画像データ",
			imageData: []byte("test image data"),
			wantError: false,
		},
		{
			name:      "異常系：空の画像データ",
			imageData: nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBucket := NewMockBucket()
			client := &Client{
				bucket:  mockBucket,
				baseURL: "gs://test-bucket",
			}

			imagePath, err := client.SaveImage(context.Background(), tt.imageData)

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

			if !hasPrefix(imagePath, "images/") {
				t.Errorf("invalid image path format: %s", imagePath)
			}
		})
	}
}

func TestSaveAndGetQuiz(t *testing.T) {
	mockBucket := NewMockBucket()
	client := &Client{
		bucket:  mockBucket,
		baseURL: "gs://test-bucket",
	}

	// テストデータ
	quiz := &models.Quiz{
		ID:                   "test-quiz",
		ImagePath:            "/images/test.jpg",
		AuthorInterpretation: "投稿者の解釈",
		AIInterpretation:     "AIの解釈",
		CreatedAt:            time.Now(),
	}

	// クイズの保存
	err := client.SaveQuiz(context.Background(), quiz)
	if err != nil {
		t.Fatalf("SaveQuiz failed: %v", err)
	}

	// クイズの取得
	savedQuiz, err := client.GetQuiz(context.Background(), quiz.ID)
	if err != nil {
		t.Fatalf("GetQuiz failed: %v", err)
	}

	// 結果の検証
	if savedQuiz.ID != quiz.ID {
		t.Errorf("expected quiz ID %q, got %q", quiz.ID, savedQuiz.ID)
	}
	if savedQuiz.ImagePath != quiz.ImagePath {
		t.Errorf("expected image path %q, got %q", quiz.ImagePath, savedQuiz.ImagePath)
	}
	if savedQuiz.AuthorInterpretation != quiz.AuthorInterpretation {
		t.Errorf("expected author interpretation %q, got %q", quiz.AuthorInterpretation, savedQuiz.AuthorInterpretation)
	}
	if savedQuiz.AIInterpretation != quiz.AIInterpretation {
		t.Errorf("expected AI interpretation %q, got %q", quiz.AIInterpretation, savedQuiz.AIInterpretation)
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
