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
	storageClient  *storage.Client
	aiClient       *ai.Client
	imageValidator ImageValidatorInterface
}

func NewQuizService(storageClient *storage.Client, aiClient *ai.Client) *QuizService {
	return &QuizService{
		storageClient:  storageClient,
		aiClient:       aiClient,
		imageValidator: NewImageValidator(10 * 1024 * 1024), // 10MB
	}
}

func (s *QuizService) CreateQuiz(ctx context.Context, file io.Reader, filename string) (*models.UploadResponse, error) {
	buf, err := s.imageValidator.ValidateAndCopy(file, filename)
	if err != nil {
		return nil, err
	}

	imageID := fmt.Sprintf("image_%d", time.Now().Unix())
	storagePath := s.generateStoragePath(imageID, nil, filepath.Ext(filename))

	// オリジナル画像を保存
	if err := s.saveImageToStorage(ctx, buf.Bytes(), storagePath); err != nil {
		return nil, fmt.Errorf("failed to upload original file: %w", err)
	}

	// Fake画像を生成して保存
	fakeImages, err := s.generateAndStoreFakeImages(ctx, bytes.NewReader(buf.Bytes()), imageID, filename)
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

func (s *QuizService) generateAndStoreFakeImages(ctx context.Context, file io.Reader, imageID, filename string) ([]string, error) {
	var fakeImages []string
	for i := 0; i < 3; i++ {
		prompt := fmt.Sprintf("画像に基づいて、似ているが少し異なる画像を生成してください。変更点：%d", i+1)

		generatedImageBytes, err := s.aiClient.GenerateImage(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate image: %w", err)
		}

		fakePath := s.generateStoragePath(imageID, &i, filepath.Ext(filename))

		// 生成画像を保存
		if err := s.saveImageToStorage(ctx, generatedImageBytes, fakePath); err != nil {
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

// saveImageToStorage は画像をCloud Storageに保存するヘルパー関数です
func (s *QuizService) saveImageToStorage(ctx context.Context, imageData []byte, path string) error {
	return s.storageClient.UploadFile(ctx, path, bytes.NewReader(imageData))
}

// generateStoragePath はストレージパスを生成するヘルパー関数です
// index == nil の場合はオリジナル画像のパスを生成
// index != nil の場合は生成画像のパスを生成
func (s *QuizService) generateStoragePath(imageID string, index *int, ext string) string {
	if index == nil {
		return fmt.Sprintf("original/%s%s", imageID, ext)
	}
	return fmt.Sprintf("generated/%s_fake%d%s", imageID, *index, ext)
}
