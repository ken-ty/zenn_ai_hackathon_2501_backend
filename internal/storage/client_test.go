package storage

import (
	"context"
	"fmt"
	"io"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWriteCloser はio.WriteCloserのモック
type MockWriteCloser struct {
	mock.Mock
}

func (m *MockWriteCloser) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockWriteCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockBucketHandle struct {
	mock.Mock
}

func (m *MockBucketHandle) Object(name string) ObjectHandle {
	args := m.Called(name)
	return args.Get(0).(ObjectHandle)
}

func (m *MockBucketHandle) SignedURL(name string, opts *storage.SignedURLOptions) (string, error) {
	args := m.Called(name, opts)
	return args.String(0), args.Error(1)
}

type MockObjectHandle struct {
	mock.Mock
}

func (m *MockObjectHandle) NewWriter(ctx context.Context) io.WriteCloser {
	args := m.Called(ctx)
	return args.Get(0).(io.WriteCloser)
}

func (m *MockObjectHandle) NewReader(ctx context.Context) (io.ReadCloser, error) {
	args := m.Called(ctx)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func TestDeleteAllQuizzes(t *testing.T) {
	// テストケースの定義
	tests := []struct {
		name    string
		setup   func(*MockBucketHandle, *MockObjectHandle)
		wantErr bool
	}{
		{
			name: "正常系：全クイズの削除に成功",
			setup: func(mb *MockBucketHandle, mo *MockObjectHandle) {
				writer := &MockWriteCloser{}
				writer.On("Write", mock.Anything).Return(10, nil)
				writer.On("Close").Return(nil)

				mo.On("NewWriter", mock.Anything).Return(writer)
				mb.On("Object", "metadata/quizzes.json").Return(mo)
			},
			wantErr: false,
		},
		{
			name: "異常系：書き込みエラー",
			setup: func(mb *MockBucketHandle, mo *MockObjectHandle) {
				writer := &MockWriteCloser{}
				writer.On("Write", mock.Anything).Return(0, fmt.Errorf("write error"))
				writer.On("Close").Return(nil)

				mo.On("NewWriter", mock.Anything).Return(writer)
				mb.On("Object", "metadata/quizzes.json").Return(mo)
			},
			wantErr: true,
		},
	}

	// テストの実行
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBucket := &MockBucketHandle{}
			mockObject := &MockObjectHandle{}
			tt.setup(mockBucket, mockObject)

			client := &Client{
				bucket: mockBucket,
			}

			err := client.DeleteAllQuizzes(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
