package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"zenn_ai_hackathon_2501_backend/internal/ai"
	"zenn_ai_hackathon_2501_backend/internal/models"
	"zenn_ai_hackathon_2501_backend/internal/storage"
)

type QuizService struct {
	storageClient *storage.Client
	aiClient      *ai.Client
}

func NewQuizService(storageClient *storage.Client, aiClient *ai.Client) *QuizService {
	return &QuizService{
		storageClient: storageClient,
		aiClient:      aiClient,
	}
}

func (s *QuizService) CreateQuiz(ctx context.Context, file io.Reader, filename string) (*models.UploadResponse, error) {
	// ファイル名の生成
	imageID := fmt.Sprintf("image_%d", time.Now().Unix())
	storagePath := fmt.Sprintf("original/%s%s", imageID, filepath.Ext(filename))

	// オリジナル画像のアップロード
	if err := s.storageClient.UploadFile(ctx, storagePath, file); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// AI生成画像の作成
	fakeImages, err := s.generateFakeImages(ctx, file, imageID, filename)
	if err != nil {
		return nil, err
	}

	// クイズデータの更新
	if err := s.updateQuizData(ctx, imageID, storagePath, fakeImages); err != nil {
		return nil, err
	}

	return &models.UploadResponse{
		ImageID:    imageID,
		StorageURL: fmt.Sprintf("gs://zenn-ai-hackathon-2501/%s", storagePath),
		Status:     "success",
	}, nil
}

func (s *QuizService) GetQuizzes(ctx context.Context) (*models.QuestionsResponse, error) {
	// クイズデータの取得
	reader, err := s.storageClient.GetFile(ctx, "metadata/questions.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read questions: %w", err)
	}

	var response models.QuestionsResponse
	if err := json.NewDecoder(reader).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode questions: %w", err)
	}

	// 署名付きURLの生成
	if err := s.generateSignedURLs(ctx, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// 内部ヘルパー関数
func (s *QuizService) generateFakeImages(ctx context.Context, file io.Reader, imageID, filename string) ([]string, error) {
	// 最初にバッファに読み込む
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var fakeImages []string
	for i := 0; i < 3; i++ {
		// プロンプトを生成
		prompt := fmt.Sprintf("画像に基づいて、似ているが少し異なる画像を生成してください。変更点：%d", i+1)

		// AI画像生成
		generatedImage, err := s.aiClient.GenerateImage(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate image: %w", err)
		}

		fakePath := fmt.Sprintf("generated/%s_fake%d%s", imageID, i, filepath.Ext(filename))
		if err := s.storageClient.UploadFile(ctx, fakePath, bytes.NewReader(generatedImage)); err != nil {
			return nil, fmt.Errorf("failed to upload generated image: %w", err)
		}

		fakeImages = append(fakeImages, fakePath)
	}
	return fakeImages, nil
}

func (s *QuizService) updateQuizData(ctx context.Context, imageID, storagePath string, fakeImages []string) error {
	reader, err := s.storageClient.GetFile(ctx, "metadata/questions.json")
	if err != nil {
		return fmt.Errorf("failed to read questions: %w", err)
	}

	var questions models.QuestionsResponse
	if err := json.NewDecoder(reader).Decode(&questions); err != nil {
		return fmt.Errorf("failed to decode questions: %w", err)
	}

	newQuestion := models.Question{
		ID:            imageID,
		OriginalImage: storagePath,
		FakeImages:    fakeImages,
		CorrectIndex:  0,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	questions.Questions = append(questions.Questions, newQuestion)

	return s.storageClient.UpdateQuestions(ctx, questions)
}

func (s *QuizService) generateSignedURLs(ctx context.Context, response *models.QuestionsResponse) error {
	for i := range response.Questions {
		signedURL, err := s.storageClient.GenerateSignedURL(ctx, response.Questions[i].OriginalImage)
		if err == nil {
			response.Questions[i].OriginalImage = signedURL
		}

		for j, fakePath := range response.Questions[i].FakeImages {
			signedURL, err := s.storageClient.GenerateSignedURL(ctx, fakePath)
			if err == nil {
				response.Questions[i].FakeImages[j] = signedURL
			}
		}
	}
	return nil
}
