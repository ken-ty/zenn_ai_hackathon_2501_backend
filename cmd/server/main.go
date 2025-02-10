package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zenn-dev/zenn-ai-hackathon/internal/ai"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/config"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/server"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/service"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/storage"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// コンテキストの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// AIクライアントの初期化
	aiClient, err := ai.NewClient(cfg.ProjectID, cfg.Location)
	if err != nil {
		log.Fatalf("AIクライアントの初期化に失敗: %v", err)
	}

	// ストレージクライアントの初期化
	storageClient, err := storage.NewClient(ctx, cfg.BucketName)
	if err != nil {
		log.Fatalf("ストレージクライアントの初期化に失敗: %v", err)
	}

	// クイズサービスの初期化
	quizService := service.NewQuizService(aiClient, storageClient)

	// HTTPサーバーの初期化
	srv := server.NewServer(quizService)
	httpServer := &http.Server{
		Addr:    cfg.GetPort(),
		Handler: srv,
	}

	// シグナルハンドリングの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// サーバーの起動
	go func() {
		log.Printf("サーバーを起動: %s", cfg.GetPort())
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("サーバーの起動に失敗: %v", err)
		}
	}()

	// シグナルの待機
	<-sigChan
	log.Println("シャットダウンを開始...")

	// グレースフルシャットダウン
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("シャットダウンに失敗: %v", err)
	}

	log.Println("サーバーを停止しました")
}
