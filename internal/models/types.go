package models

type Question struct {
	ID            string   `json:"id"`
	OriginalImage string   `json:"original_image"`
	FakeImages    []string `json:"fake_images"`
	CorrectIndex  int      `json:"correct_index"`
	CreatedAt     string   `json:"created_at"`
}

type QuestionsResponse struct {
	Questions []Question `json:"questions"`
}

type UploadResponse struct {
	ImageID    string `json:"image_id"`
	StorageURL string `json:"storage_url"`
	Status     string `json:"status"`
}
