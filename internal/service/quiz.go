package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/zenn-dev/zenn-ai-hackathon/internal/ai"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/logging"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/storage"
)

// QuizService はクイズ関連の操作を提供するインターフェース
type QuizService interface {
	CreateQuiz(ctx context.Context, imageData []byte, authorInterpretation string) (*models.Quiz, error)
	GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error)
	GetRandomizedInterpretations(quiz *models.Quiz) []string
	VerifyAnswer(quiz *models.Quiz, selectedInterpretation string) bool
	GetSignedImageURL(ctx context.Context, imagePath string) (string, error)
	GetQuizList(ctx context.Context) ([]*models.Quiz, error)
	DeleteAllQuizzes(ctx context.Context) error
}

// QuizServiceImpl はクイズ関連の操作を実装します
type QuizServiceImpl struct {
	aiClient      ai.AIClient
	storageClient storage.StorageClient
}

// NewQuizService は新しいQuizServiceインスタンスを作成します
func NewQuizService(aiClient ai.AIClient, storageClient storage.StorageClient) QuizService {
	return &QuizServiceImpl{
		aiClient:      aiClient,
		storageClient: storageClient,
	}
}

// CreateQuiz は新しいクイズを作成します
func (s *QuizServiceImpl) CreateQuiz(ctx context.Context, imageData []byte, authorInterpretation string) (*models.Quiz, error) {
	// 入力値の検証
	if len(imageData) == 0 {
		return nil, fmt.Errorf("画像データが必要です")
	}
	if authorInterpretation == "" {
		return nil, fmt.Errorf("投稿者の解釈が必要です")
	}

	// 画像の保存
	imagePath, err := s.storageClient.SaveImage(ctx, imageData)
	if err != nil {
		return nil, fmt.Errorf("画像の保存に失敗: %w", err)
	}

	// AIによる代替解釈の生成
	aiInterpretation, err := s.aiClient.GenerateInterpretation(ctx, imageData, authorInterpretation)
	if err != nil {
		return nil, fmt.Errorf("AIによる解釈の生成に失敗: %w", err)
	}

	// クイズの作成
	quiz := &models.Quiz{
		ID:                   generateID(),
		ImagePath:            imagePath,
		AuthorInterpretation: authorInterpretation,
		AIInterpretation:     aiInterpretation,
		CreatedAt:            time.Now(),
	}

	// クイズの保存
	if err := s.storageClient.SaveQuiz(ctx, quiz); err != nil {
		return nil, fmt.Errorf("クイズの保存に失敗: %w", err)
	}

	return quiz, nil
}

// GetQuiz は指定されたIDのクイズを取得します
func (s *QuizServiceImpl) GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error) {
	if quizID == "" {
		return nil, fmt.Errorf("クイズIDが必要です")
	}

	quiz, err := s.storageClient.GetQuiz(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("クイズの取得に失敗: %w", err)
	}

	return quiz, nil
}

// GetRandomizedInterpretations は解釈をランダムな順序で返します
func (s *QuizServiceImpl) GetRandomizedInterpretations(quiz *models.Quiz) []string {
	interpretations := []string{
		quiz.AuthorInterpretation,
		quiz.AIInterpretation,
	}
	shuffle(interpretations)
	return interpretations
}

// VerifyAnswer は回答が正しいかを検証します
func (s *QuizServiceImpl) VerifyAnswer(quiz *models.Quiz, selectedInterpretation string) bool {
	return selectedInterpretation == quiz.AuthorInterpretation
}

// generateID は一意のIDを生成します
func generateID() string {
	return fmt.Sprintf("quiz_%d", time.Now().UnixNano())
}

// shuffle はスライスの要素をランダムに並び替えます
func shuffle(slice []string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(slice) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// GetSignedImageURL は画像の署名付きURLを生成します
func (s *QuizServiceImpl) GetSignedImageURL(ctx context.Context, imagePath string) (string, error) {
	logging.Debug("GetSignedImageURL: imagePath=%s の署名付きURL生成を開始", imagePath)
	if imagePath == "" {
		logging.Error("GetSignedImageURL: 画像パスが空です")
		return "", fmt.Errorf("画像パスが必要です")
	}

	signedURL, err := s.storageClient.GenerateSignedURL(ctx, imagePath)
	if err != nil {
		logging.Error("GetSignedImageURL: 署名付きURLの生成に失敗: %v", err)
		return "", fmt.Errorf("署名付きURLの生成に失敗: %w", err)
	}

	logging.Debug("GetSignedImageURL: 署名付きURL生成成功: URL=%s", signedURL)
	return signedURL, nil
}

// GetQuizList はすべてのクイズを取得します
func (s *QuizServiceImpl) GetQuizList(ctx context.Context) ([]*models.Quiz, error) {
	quizzes, err := s.storageClient.GetQuizzes(ctx)
	if err != nil {
		return nil, fmt.Errorf("クイズ一覧の取得に失敗: %w", err)
	}
	return quizzes, nil
}

// DeleteAllQuizzes は全てのクイズを削除します
func (s *QuizServiceImpl) DeleteAllQuizzes(ctx context.Context) error {
	return s.storageClient.DeleteAllQuizzes(ctx)
}
