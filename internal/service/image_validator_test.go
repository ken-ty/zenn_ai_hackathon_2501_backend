package service

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestImageValidator_ValidateAndCopy(t *testing.T) {
	validator := NewImageValidator(5 * 1024 * 1024) // 5MB

	tests := []struct {
		name     string
		file     io.Reader
		filename string
		wantErr  bool
	}{
		{
			name:     "invalid extension",
			file:     strings.NewReader("dummy data"),
			filename: "test.txt",
			wantErr:  true,
		},
		{
			name:     "file too large",
			file:     bytes.NewReader(make([]byte, 6*1024*1024)), // 6MB
			filename: "test.jpg",
			wantErr:  true,
		},
		// 実際の画像ファイルを使用したテストケースは別途追加することをお勧めします
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.ValidateAndCopy(tt.file, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndCopy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
