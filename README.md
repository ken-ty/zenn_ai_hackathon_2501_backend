# AI Art Quiz

[![codecov](https://codecov.io/gh/ken-ty/zenn_ai_hackathon_2501_backend/graph/badge.svg)](https://codecov.io/gh/ken-ty/zenn_ai_hackathon_2501_backend)

AIを活用したアート作品の解釈クイズアプリケーション

## 概要

このアプリケーションは、作品とその解釈を組み合わせたマルチモーダルクイズゲームのバックエンドシステムです。
Vertex AI (Gemini Pro Vision)を活用して、作品に対する説得力のある代替解釈を生成します。

## 機能

- 作品画像のアップロード
- 作者による解釈の登録
- AIによる代替解釈の生成
- クイズの作成と保存
- ランダム化された解釈の提供
- 回答の検証

## 必要条件

- Go 1.21以上
- Google Cloud Platform アカウント
- 以下のAPIの有効化:
  - Cloud Storage API
  - Vertex AI API

## 環境変数

```bash
# 必須
PROJECT_ID=your-project-id
BUCKET_NAME=your-bucket-name

# オプション（デフォルト値あり）
LOCATION=us-central1  # Vertex AIのリージョン
PORT=8080            # サーバーのポート番号
```

## インストール

```bash
# リポジトリのクローン
git clone https://github.com/zenn-dev/zenn-ai-hackathon.git
cd zenn-ai-hackathon

# 依存関係のインストール
go mod download
```

## 使用方法

### サーバーの起動

```bash
# 環境変数の設定
export PROJECT_ID=your-project-id
export BUCKET_NAME=your-bucket-name

# サーバーの起動
go run cmd/server/main.go
```

### APIの利用

#### 1. クイズの作成

```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@artwork.jpg" \
  -F "interpretation=作者による解釈のテキスト"
```

レスポンス:
```json
{
  "id": "quiz_1234567890",
  "image_path": "/images/artwork.jpg",
  "author_interpretation": "作者による解釈のテキスト",
  "ai_interpretation": "AIによる代替解釈のテキスト",
  "created_at": "2024-03-20T10:00:00Z"
}
```

#### 2. クイズの取得

```bash
curl http://localhost:8080/quiz/quiz_1234567890
```

レスポンス:
```json
{
  "id": "quiz_1234567890",
  "image_path": "/images/artwork.jpg",
  "interpretations": [
    "作者による解釈のテキスト",
    "AIによる代替解釈のテキスト"
  ]
}
```

## テスト

```bash
# すべてのテストを実行
go test ./...

# カバレッジレポートの生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## デプロイ

```bash
# Cloud Runへのデプロイ
gcloud run deploy ai-art-quiz \
  --source . \
  --platform managed \
  --region us-central1 \
  --set-env-vars PROJECT_ID=your-project-id,BUCKET_NAME=your-bucket-name
```

## ライセンス

MIT License

## 貢献

1. このリポジトリをフォーク
2. 新しいブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチをプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## 関連ドキュメント

- [アーキテクチャ設計](docs/ARCHITECTURE.md)
- [API仕様](docs/API.md)
