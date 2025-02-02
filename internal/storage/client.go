package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/logging"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
)

// StorageClient はストレージ操作のインターフェースを定義します
type StorageClient interface {
	SaveImage(ctx context.Context, imageData []byte) (string, error)
	SaveQuiz(ctx context.Context, quiz *models.Quiz) error
	GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error)
	GenerateSignedURL(ctx context.Context, objectPath string) (string, error)
	GetQuizzes(ctx context.Context) ([]*models.Quiz, error)
	DeleteAllQuizzes(ctx context.Context) error
}

// BucketHandle はCloud Storage Bucketのインターフェース
type BucketHandle interface {
	Object(name string) ObjectHandle
	SignedURL(name string, opts *storage.SignedURLOptions) (string, error)
}

// ObjectHandle はCloud Storage Objectのインターフェース
type ObjectHandle interface {
	NewWriter(ctx context.Context) io.WriteCloser
	NewReader(ctx context.Context) (io.ReadCloser, error)
}

// bucketHandleAdapter はCloud Storage BucketHandleのアダプター
type bucketHandleAdapter struct {
	bucket *storage.BucketHandle
}

// objectHandleAdapter はCloud Storage ObjectHandleのアダプター
type objectHandleAdapter struct {
	obj *storage.ObjectHandle
}

func (b *bucketHandleAdapter) Object(name string) ObjectHandle {
	return &objectHandleAdapter{obj: b.bucket.Object(name)}
}

func (b *bucketHandleAdapter) SignedURL(name string, opts *storage.SignedURLOptions) (string, error) {
	return b.bucket.SignedURL(name, opts)
}

func (o *objectHandleAdapter) NewWriter(ctx context.Context) io.WriteCloser {
	return o.obj.NewWriter(ctx)
}

func (o *objectHandleAdapter) NewReader(ctx context.Context) (io.ReadCloser, error) {
	return o.obj.NewReader(ctx)
}

// Client はCloud Storageとの通信を担当します
type Client struct {
	bucket  BucketHandle
	baseURL string
}

// NewClient は新しいストレージクライアントを作成します
func NewClient(ctx context.Context, bucketName string) (StorageClient, error) {
	logging.Info("ストレージクライアントの初期化を開始: bucket=%s", bucketName)
	client, err := storage.NewClient(ctx)
	if err != nil {
		logging.Error("ストレージクライアントの作成に失敗: %v", err)
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	logging.Debug("ストレージクライアントの作成に成功")

	bucket := client.Bucket(bucketName)
	baseURL := fmt.Sprintf("gs://%s", bucketName)

	return &Client{
		bucket:  &bucketHandleAdapter{bucket: bucket},
		baseURL: baseURL,
	}, nil
}

// SaveImage は画像をCloud Storageに保存します
func (c *Client) SaveImage(ctx context.Context, imageData []byte) (string, error) {
	logging.Info("画像の保存を開始: サイズ=%d bytes", len(imageData))
	if len(imageData) == 0 {
		logging.Error("画像データが空です")
		return "", fmt.Errorf("画像データが必要です")
	}

	imagePath := fmt.Sprintf("images/%s.jpg", generateID())
	logging.Debug("保存先パス: %s", imagePath)

	obj := c.bucket.Object(imagePath)
	writer := obj.NewWriter(ctx)

	if _, err := writer.Write(imageData); err != nil {
		logging.Error("画像データの書き込みに失敗: %v", err)
		return "", fmt.Errorf("画像の書き込みに失敗: %w", err)
	}

	if err := writer.Close(); err != nil {
		logging.Error("画像ファイルのクローズに失敗: %v", err)
		return "", fmt.Errorf("画像の保存に失敗: %w", err)
	}

	logging.Info("画像の保存に成功: path=%s", imagePath)
	return imagePath, nil
}

// SaveQuiz はクイズデータをCloud Storageに保存します
func (c *Client) SaveQuiz(ctx context.Context, quiz *models.Quiz) error {
	logging.Info("クイズの保存を開始: id=%s", quiz.ID)
	if quiz == nil {
		logging.Error("クイズデータがnilです")
		return fmt.Errorf("クイズデータが必要です")
	}

	quizzes, err := c.loadQuizzes(ctx)
	if err != nil {
		logging.Error("既存クイズデータの読み込みに失敗: %v", err)
		return fmt.Errorf("クイズデータの読み込みに失敗: %w", err)
	}
	logging.Debug("既存クイズ数: %d", len(quizzes.Quizzes))

	// 新しいクイズを追加
	quizzes.Quizzes = append(quizzes.Quizzes, quiz)

	// JSONに変換
	data, err := json.Marshal(quizzes)
	if err != nil {
		logging.Error("クイズデータのJSON変換に失敗: %v", err)
		return fmt.Errorf("JSONへの変換に失敗: %w", err)
	}

	// ファイルに書き込み
	obj := c.bucket.Object("metadata/quizzes.json")
	writer := obj.NewWriter(ctx)
	if _, err := writer.Write(data); err != nil {
		logging.Error("クイズデータの書き込みに失敗: %v", err)
		return fmt.Errorf("クイズデータの書き込みに失敗: %w", err)
	}

	if err := writer.Close(); err != nil {
		logging.Error("クイズファイルのクローズに失敗: %v", err)
		return fmt.Errorf("クイズデータの保存に失敗: %w", err)
	}

	logging.Info("クイズの保存に成功: id=%s", quiz.ID)
	return nil
}

// GetQuiz は指定されたIDのクイズを取得します
func (c *Client) GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error) {
	logging.Info("クイズの取得を開始: id=%s", quizID)
	if quizID == "" {
		logging.Error("クイズIDが空です")
		return nil, fmt.Errorf("クイズIDが必要です")
	}

	quizzes, err := c.loadQuizzes(ctx)
	if err != nil {
		logging.Error("クイズデータの読み込みに失敗: %v", err)
		return nil, fmt.Errorf("クイズデータの読み込みに失敗: %w", err)
	}

	for _, quiz := range quizzes.Quizzes {
		if quiz.ID == quizID {
			logging.Info("クイズの取得に成功: id=%s", quizID)
			return quiz, nil
		}
	}

	logging.Warn("クイズが見つかりません: id=%s", quizID)
	return nil, fmt.Errorf("クイズが見つかりません: %s", quizID)
}

// loadQuizzes はすべてのクイズデータを読み込みます
func (c *Client) loadQuizzes(ctx context.Context) (*models.QuizList, error) {
	logging.Debug("クイズデータの読み込みを開始")
	obj := c.bucket.Object("metadata/quizzes.json")
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			logging.Info("クイズデータファイルが存在しないため、新規作成します")
			return &models.QuizList{}, nil
		}
		logging.Error("クイズデータファイルの読み込みに失敗: %v", err)
		return nil, fmt.Errorf("クイズデータの読み込みに失敗: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		logging.Error("クイズデータの読み取りに失敗: %v", err)
		return nil, fmt.Errorf("データの読み込みに失敗: %w", err)
	}

	var quizzes models.QuizList
	if err := json.Unmarshal(data, &quizzes); err != nil {
		logging.Error("クイズデータのJSONパースに失敗: %v", err)
		return nil, fmt.Errorf("JSONのパースに失敗: %w", err)
	}

	logging.Debug("クイズデータの読み込みに成功: クイズ数=%d", len(quizzes.Quizzes))
	return &quizzes, nil
}

// generateID は一意のIDを生成します
func generateID() string {
	return fmt.Sprintf("quiz_%d", time.Now().UnixNano())
}

// GenerateSignedURL は、指定されたオブジェクトの公開URLを生成します。
func (c *Client) GenerateSignedURL(ctx context.Context, objectPath string) (string, error) {
	if objectPath == "" {
		return "", fmt.Errorf("object path is empty")
	}

	// パブリックアクセス用のURLを生成
	publicURL := fmt.Sprintf("https://storage.googleapis.com/zenn-ai-hackathon-2501-bucket/%s", objectPath)
	logging.Info("Generated public URL: %s", publicURL)
	return publicURL, nil
}

// GetQuizzes はすべてのクイズを取得します
func (c *Client) GetQuizzes(ctx context.Context) ([]*models.Quiz, error) {
	logging.Info("クイズ一覧の取得を開始")
	quizzes, err := c.loadQuizzes(ctx)
	if err != nil {
		logging.Error("クイズデータの読み込みに失敗: %v", err)
		return nil, fmt.Errorf("クイズデータの読み込みに失敗: %w", err)
	}
	logging.Info("クイズ一覧の取得に成功: クイズ数=%d", len(quizzes.Quizzes))
	return quizzes.Quizzes, nil
}

// DeleteAllQuizzes は全てのクイズを削除します
func (c *Client) DeleteAllQuizzes(ctx context.Context) error {
	logging.Info("全クイズの削除を開始")

	// 空のクイズリストを作成
	emptyQuizzes := struct {
		Quizzes []*models.Quiz `json:"quizzes"`
	}{
		Quizzes: []*models.Quiz{},
	}

	// JSONに変換
	data, err := json.Marshal(emptyQuizzes)
	if err != nil {
		logging.Error("空のクイズリストのJSON変換に失敗: %v", err)
		return fmt.Errorf("JSONへの変換に失敗: %w", err)
	}

	// ファイルに書き込み
	obj := c.bucket.Object("metadata/quizzes.json")
	writer := obj.NewWriter(ctx)
	if _, err := writer.Write(data); err != nil {
		logging.Error("クイズデータの書き込みに失敗: %v", err)
		return fmt.Errorf("クイズデータの書き込みに失敗: %w", err)
	}

	if err := writer.Close(); err != nil {
		logging.Error("クイズファイルのクローズに失敗: %v", err)
		return fmt.Errorf("クイズデータの保存に失敗: %w", err)
	}

	logging.Info("全クイズの削除に成功")
	return nil
}
