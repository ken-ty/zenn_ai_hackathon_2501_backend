package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		wantError bool
	}{
		{
			name: "正常系：必須環境変数あり",
			envVars: map[string]string{
				"PROJECT_ID":  "test-project",
				"BUCKET_NAME": "test-bucket",
				"LOCATION":    "us-central1",
				"PORT":        "8080",
				"DEBUG":       "true",
			},
			wantError: false,
		},
		{
			name: "正常系：オプション環境変数なし",
			envVars: map[string]string{
				"PROJECT_ID":  "test-project",
				"BUCKET_NAME": "test-bucket",
			},
			wantError: false,
		},
		{
			name: "異常系：PROJECT_IDなし",
			envVars: map[string]string{
				"BUCKET_NAME": "test-bucket",
			},
			wantError: true,
		},
		{
			name: "異常系：BUCKET_NAMEなし",
			envVars: map[string]string{
				"PROJECT_ID": "test-project",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア
			os.Clearenv()

			// テスト用の環境変数を設定
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// テストの実行
			cfg, err := Load()

			// エラーの検証
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// 設定値の検証
			if cfg.ProjectID != tt.envVars["PROJECT_ID"] {
				t.Errorf("expected ProjectID %q, got %q", tt.envVars["PROJECT_ID"], cfg.ProjectID)
			}
			if cfg.BucketName != tt.envVars["BUCKET_NAME"] {
				t.Errorf("expected BucketName %q, got %q", tt.envVars["BUCKET_NAME"], cfg.BucketName)
			}

			// デフォルト値の検証
			if tt.envVars["LOCATION"] == "" && cfg.Location != "us-central1" {
				t.Errorf("expected default Location %q, got %q", "us-central1", cfg.Location)
			}
			if tt.envVars["PORT"] == "" && cfg.Port != "8080" {
				t.Errorf("expected default Port %q, got %q", "8080", cfg.Port)
			}
		})
	}
}

func TestGetPort(t *testing.T) {
	cfg := &Config{Port: "8080"}
	expected := ":8080"
	if got := cfg.GetPort(); got != expected {
		t.Errorf("GetPort() = %q, want %q", got, expected)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "正常系：すべての項目が設定されている",
			config: &Config{
				ProjectID:  "test-project",
				BucketName: "test-bucket",
				Location:   "test-location",
				Port:       "8080",
			},
			wantError: false,
		},
		{
			name: "異常系：ProjectIDが未設定",
			config: &Config{
				BucketName: "test-bucket",
				Location:   "test-location",
				Port:       "8080",
			},
			wantError: true,
		},
		{
			name: "異常系：BucketNameが未設定",
			config: &Config{
				ProjectID: "test-project",
				Location:  "test-location",
				Port:      "8080",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
