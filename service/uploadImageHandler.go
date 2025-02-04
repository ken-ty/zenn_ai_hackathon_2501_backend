package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"log"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	aiplatformpb "cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/structpb"
)

// UploadImageHandler handles image upload requests
func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// マルチパートフォームファイルを取得
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// ファイル形式チェック
	contentType := header.Header.Get("Content-Type")
	if !isAllowedFileType(contentType) {
		http.Error(w, "File type not allowed. Please upload an image file.", http.StatusBadRequest)
		return
	}

	// Cloud Storage にアップロード
	imageURL, err := uploadToCloudStorage(file, header.Filename)
	if err != nil {
		http.Error(w, "Failed to upload to Cloud Storage: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Imagen に画像生成リクエスト
	fakeImages, err := generateFakeImages(imageURL)
	if err != nil {
		http.Error(w, "Failed to generate fake images: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Firestore にクイズデータを保存
	quizID, err := saveQuizData(imageURL, fakeImages)
	if err != nil {
		http.Error(w, "Failed to save quiz data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスを返す
	response := map[string]interface{}{
		"message": "Quiz created successfully",
		"quiz_id": quizID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// uploadToCloudStorage uploads an image to Google Cloud Storage
func uploadToCloudStorage(file multipart.File, filename string) (string, error) {

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		return "", fmt.Errorf("BUCKET_NAME is not set")
	}

	// log
	log.Printf("Uploading to bucket: %s", bucketName)
	log.Printf("Object name: %s", filename)

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(filename)
	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, file); err != nil {
		return "", fmt.Errorf("failed to copy file to storage: %v", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	return fmt.Sprintf("gs://%s/%s", bucketName, filename), nil
}

// generateFakeImages generates fake images using Imagen
func generateFakeImages(imageURL string) ([]string, error) {
	ctx := context.Background()
	client, err := aiplatform.NewPredictionClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create prediction client: %v", err)
	}
	defer client.Close()

	projectID := os.Getenv("PROJECT_ID")
	location := os.Getenv("LOCATION")
	if projectID == "" || location == "" {
		return nil, fmt.Errorf("PROJECT_ID and LOCATION environment variables must be set")
	}

	parameters, err := structpb.NewStruct(map[string]interface{}{
		"num_images": 5,
		"quality":    "high",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create parameters: %v", err)
	}

	req := &aiplatformpb.PredictRequest{
		Endpoint: fmt.Sprintf("projects/%s/locations/%s/endpoints/imagen", projectID, location),
		Instances: []*structpb.Value{
			{
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"image_url": {
								Kind: &structpb.Value_StringValue{
									StringValue: imageURL,
								},
							},
						},
					},
				},
			},
		},
		Parameters: &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: parameters,
			},
		},
	}

	resp, err := client.Predict(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to predict: %v", err)
	}

	var imageUrls []string
	// レスポンスの解析（実際のAPIレスポンス形式に応じて調整が必要）
	for _, prediction := range resp.Predictions {
		url := prediction.GetStructValue().Fields["image_url"].GetStringValue()
		imageUrls = append(imageUrls, url)
	}

	return imageUrls, nil
}

// saveQuizData saves quiz data to Firestore
func saveQuizData(realImage string, fakeImages []string) (string, error) {
	ctx := context.Background()
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		return "", fmt.Errorf("PROJECT_ID environment variable must be set")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to create firestore client: %v", err)
	}
	defer client.Close()

	quizData := map[string]interface{}{
		"real_image":  realImage,
		"fake_images": fakeImages,
		"created_at":  time.Now(),
	}

	docRef, _, err := client.Collection("quizzes").Add(ctx, quizData)
	if err != nil {
		return "", fmt.Errorf("failed to add quiz data: %v", err)
	}

	return docRef.ID, nil
}

// isAllowedFileType checks if the file type is allowed
func isAllowedFileType(contentType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
	return allowedTypes[contentType]
}

// GetQuizzesHandler handles fetching all quizzes
func GetQuizzesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "your-project-id")
	if err != nil {
		http.Error(w, "Failed to connect to Firestore", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	var quizzes []map[string]interface{}
	iter := client.Collection("quizzes").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, "Failed to fetch quizzes", http.StatusInternalServerError)
			return
		}
		quizzes = append(quizzes, doc.Data())
	}

	json.NewEncoder(w).Encode(quizzes)
}
