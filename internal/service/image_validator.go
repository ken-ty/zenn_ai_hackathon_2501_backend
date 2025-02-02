package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// ImageValidatorInterface は画像検証の機能を定義するインターフェース
type ImageValidatorInterface interface {
	ValidateAndCopy(file io.Reader, filename string) (*bytes.Buffer, error)
}

// ImageValidator は画像の検証を行う構造体
type ImageValidator struct {
	maxFileSize      int64
	allowedMimeTypes map[string]bool
	allowedExts      map[string]bool
}

// NewImageValidator は新しいImageValidatorを作成します
func NewImageValidator(maxFileSize int64) *ImageValidator {
	return &ImageValidator{
		maxFileSize: maxFileSize,
		allowedMimeTypes: map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
		},
		allowedExts: map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		},
	}
}

// ValidateAndCopy は画像を検証し、バッファにコピーします
func (v *ImageValidator) ValidateAndCopy(file io.Reader, filename string) (*bytes.Buffer, error) {
	// 拡張子の検証
	ext := strings.ToLower(filepath.Ext(filename))
	if !v.allowedExts[ext] {
		return nil, fmt.Errorf("unsupported file format: %s. Allowed formats: jpg, jpeg, png", ext)
	}

	// ファイルサイズの制限付きで読み込み
	limitedReader := io.LimitReader(file, v.maxFileSize)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, limitedReader); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// ファイルサイズのチェック
	if int64(buf.Len()) >= v.maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", v.maxFileSize)
	}

	// MIMEタイプの検証
	mimeType := http.DetectContentType(buf.Bytes())
	if !v.allowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("invalid file type: %s. Only jpeg and png are allowed", mimeType)
	}

	return buf, nil
}
