// Start of Selection
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/zenn-dev/zenn-ai-hackathon/internal/ai"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/config"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/logging"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/server"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/service"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/storage"
)

func main() {
	// ログレベルの設定
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "DEBUG" // デフォルトでDEBUGレベルに設定
	}
	logging.Info("ログレベルを %s に設定しました。", logLevel)

	// 認証情報の確認
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		if _, err := os.Stat(credPath); err != nil {
			logging.Debug("認証情報ファイル %s が見つかりません。Workload Identity Federation を使用します。", credPath)
		} else {
			logging.Info("ローカル開発用の認証情報ファイル %s を使用します。", credPath)
		}
	} else {
		logging.Info("GOOGLE_APPLICATION_CREDENTIALS が未設定です。Workload Identity Federation を使用します。")
	}

	// 作業ディレクトリの確認
	if wd, err := os.Getwd(); err != nil {
		logging.Error("作業ディレクトリの取得に失敗しました。")
	} else {
		logging.Info("現在の作業ディレクトリを確認しました。")
		// ディレクトリ構造の表示
		logging.Info("ディレクトリ構造:")
		filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logging.Error("パスの走査中にエラーが発生しました。")
				return nil
			}
			rel, err := filepath.Rel(wd, path)
			if err != nil {
				rel = path
			}
			logging.Debug("- %s", rel)
			return nil
		})
	}

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		logging.Error("設定の読み込みに失敗しました。")
		os.Exit(1)
	}
	logging.Info("設定を読み込みました。")

	// ポート番号の取得と検証
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		logging.Info("PORT環境変数が設定されていません。デフォルトポート %s を使用します。", port)
	} else {
		logging.Info("環境変数から指定されたポートを使用します。")
	}

	// コンテキストの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// AIクライアントの初期化
	aiClient, err := ai.NewClient(cfg.ProjectID, cfg.Location)
	if err != nil {
		logging.Error("AIクライアントの初期化に失敗しました。")
		dumpError(err)
		os.Exit(1)
	}
	logging.Info("AIクライアントを初期化しました。")

	// ストレージクライアントの初期化
	storageClient, err := storage.NewClient(ctx, cfg.BucketName)
	if err != nil {
		logging.Error("ストレージクライアントの初期化に失敗しました。")
		dumpError(err)
		os.Exit(1)
	}
	logging.Info("ストレージクライアントを初期化しました。")

	// サービスの初期化
	quizService := service.NewQuizService(aiClient, storageClient)
	logging.Info("クイズサービスを初期化しました。")

	// サーバーの初期化
	srv := server.NewServer(quizService)
	logging.Info("HTTPサーバーを初期化しました。")

	// HTTPサーバーの設定
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv,
	}

	// シグナル処理の設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// サーバーの起動
	logging.Info("サーバーを起動します - アドレス: %s", addr)
	go func() {
		logging.Info("HTTPサーバーを起動します。")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("サーバーの起動に失敗しました。")
			dumpError(err)
			cancel() // コンテキストをキャンセルしてシャットダウンを開始
		}
	}()

	// 準備完了のログ
	logging.Info("サーバーの準備が完了しました - %s", addr)

	// シグナルを待機
	sig := <-sigChan
	logging.Info("シグナル %v を受信。シャットダウンを開始します...", sig)

	// シャットダウン処理
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logging.Error("シャットダウン中にエラーが発生しました。")
		dumpError(err)
	}

	logging.Info("サーバーを停止しました。")
}

func dumpError(err error) {
	for err != nil {
		logging.Error("エラーが発生しました。")
		logging.Error("エラーの種類: %T", err)
		logging.Error("エラーメッセージ: %v", err)
		if unwrappable, ok := err.(interface{ Unwrap() error }); ok {
			wrapped := unwrappable.Unwrap()
			if wrapped != nil {
				logging.Error("内包されたエラーがあります。")
				err = wrapped
			} else {
				break
			}
		} else {
			break
		}
	}
}
