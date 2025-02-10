package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
	"golang.org/x/oauth2/google"
)

// StorageClient はストレージ操作のインターフェースを定義します
type StorageClient interface {
	SaveImage(ctx context.Context, imageData []byte) (string, error)
	SaveQuiz(ctx context.Context, quiz *models.Quiz) error
	GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error)
	GenerateSignedURL(ctx context.Context, objectPath string) (string, error)
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
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}

	bucket := client.Bucket(bucketName)
	baseURL := fmt.Sprintf("gs://%s", bucketName)

	return &Client{
		bucket:  &bucketHandleAdapter{bucket: bucket},
		baseURL: baseURL,
	}, nil
}

// SaveImage は画像をCloud Storageに保存します
func (c *Client) SaveImage(ctx context.Context, imageData []byte) (string, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("画像データが必要です")
	}

	imagePath := fmt.Sprintf("images/%s.jpg", generateID())
	obj := c.bucket.Object(imagePath)
	writer := obj.NewWriter(ctx)

	if _, err := writer.Write(imageData); err != nil {
		return "", fmt.Errorf("画像の書き込みに失敗: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("画像の保存に失敗: %w", err)
	}

	return imagePath, nil
}

// SaveQuiz はクイズデータをCloud Storageに保存します
func (c *Client) SaveQuiz(ctx context.Context, quiz *models.Quiz) error {
	if quiz == nil {
		return fmt.Errorf("クイズデータが必要です")
	}

	quizzes, err := c.loadQuizzes(ctx)
	if err != nil {
		return fmt.Errorf("クイズデータの読み込みに失敗: %w", err)
	}

	// 新しいクイズを追加
	quizzes.Quizzes = append(quizzes.Quizzes, quiz)

	// JSONに変換
	data, err := json.Marshal(quizzes)
	if err != nil {
		return fmt.Errorf("JSONへの変換に失敗: %w", err)
	}

	// ファイルに書き込み
	obj := c.bucket.Object("metadata/quizzes.json")
	writer := obj.NewWriter(ctx)
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("クイズデータの書き込みに失敗: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("クイズデータの保存に失敗: %w", err)
	}

	return nil
}

// GetQuiz は指定されたIDのクイズを取得します
func (c *Client) GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error) {
	if quizID == "" {
		return nil, fmt.Errorf("クイズIDが必要です")
	}

	quizzes, err := c.loadQuizzes(ctx)
	if err != nil {
		return nil, fmt.Errorf("クイズデータの読み込みに失敗: %w", err)
	}

	for _, quiz := range quizzes.Quizzes {
		if quiz.ID == quizID {
			return quiz, nil
		}
	}

	return nil, fmt.Errorf("クイズが見つかりません: %s", quizID)
}

// loadQuizzes はすべてのクイズデータを読み込みます
func (c *Client) loadQuizzes(ctx context.Context) (*models.QuizList, error) {
	obj := c.bucket.Object("metadata/quizzes.json")
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return &models.QuizList{}, nil
		}
		return nil, fmt.Errorf("クイズデータの読み込みに失敗: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("データの読み込みに失敗: %w", err)
	}

	var quizzes models.QuizList
	if err := json.Unmarshal(data, &quizzes); err != nil {
		return nil, fmt.Errorf("JSONのパースに失敗: %w", err)
	}

	return &quizzes, nil
}

// generateID は一意のIDを生成します
func generateID() string {
	return fmt.Sprintf("quiz_%d", time.Now().UnixNano())
}

// GenerateSignedURL は、指定されたオブジェクトの署名付きURLを生成します。
// このURLは、オブジェクトの読み取り専用アクセスを許可し、有効期限が15分間です。
func (c *Client) GenerateSignedURL(ctx context.Context, objectPath string) (string, error) {
	log.Printf("Starting to generate signed URL for: %s", objectPath)

	// キーファイルを直接読み込む
	credBytes, err := os.ReadFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		log.Printf("Error reading credentials file: %v", err)
		return "", fmt.Errorf("failed to read credentials: %v", err)
	}
	log.Printf("Successfully read credentials file")

	conf, err := google.JWTConfigFromJSON(credBytes)
	if err != nil {
		log.Printf("Error parsing JWT config: %v", err)
		return "", fmt.Errorf("failed to parse JWT config: %v", err)
	}
	log.Printf("Successfully parsed JWT config, using email: %s", conf.Email)

	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(15 * time.Minute),
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
	}

	url, err := c.bucket.SignedURL(objectPath, opts)
	if err != nil {
		log.Printf("Error generating signed URL: %v", err)
		return "", err
	}
	log.Printf("Successfully generated signed URL for: %s", objectPath)

	return url, nil
}
