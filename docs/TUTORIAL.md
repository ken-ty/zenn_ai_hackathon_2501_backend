# チュートリアルガイド

## 🎯 概要

Zenn AI Agent Hackathon 2024-2025向けのAI画像ゲームバックエンドAPIを構築します。

### プロジェクト情報

- 言語：Go 1.22.5
- プラットフォーム：Google Cloud Platform
- 開発範囲：バックエンドAPI

## 🔧 前提条件

### 必要なツール

| ツール | バージョン | 用途 |
|--------|------------|------|
| Go | 1.22.5以上 | バックエンド開発 |
| gcloud CLI | 最新 | GCPリソース管理 |
| Docker | 最新 | ローカル開発・デプロイ |

## 🚀 開発環境のセットアップ

### 1. Google Cloudの設定

```bash
# GCPプロジェクトの設定
gcloud config set project zenn-ai-hackathon-2501
gcloud config set run/region us-east1

# 必要なAPIの有効化
gcloud services enable \
  run.googleapis.com \
  aiplatform.googleapis.com \
  storage.googleapis.com
```

### 2. 必要なGoogle Cloudサービスの有効化

- Vertex AI API
- Cloud Storage API
- Cloud Run API

### 3. 認証設定

- Google Cloud Console でサービスアカウントを作成
- キーファイル（keyfile.json）をダウンロード
- `config/credentials/` に配置

## 💾 主要機能

### 画像処理API

1. オリジナル画像のアップロード（/upload）
2. Vertex AIを使用した画像生成
3. クイズ情報の取得（/questions）

## 🔧 デプロイ手順

### Cloud Runへのデプロイ

```bash
gcloud run deploy zenn-ai --source .
```

### デプロイ後の確認

### Cloud Storageの設定

- バケット名：zenn-ai-hackathon-2501_original_images
- リージョン：us-east1

## 🏆 ハッカソン提出要件

### 必須要件

1. Google Cloud AIプロダクトの使用（最低1つ）
   - 本プロジェクトではVertex AIを使用

2. Google Cloudコンピュートプロダクトの使用（最低1つ）
   - 本プロジェクトではCloud Runを使用

## 🐛 トラブルシューティング

### よくあるエラー

1. 認証エラー
   - keyfile.jsonの配置確認
   - 環境変数の設定確認
2. ストレージエラー
   - バケットの権限設定
   - リージョンの確認

## 📚 参考リンク

### ハッカソン関連

- [Zenn AI Agent Hackathon 公式ページ](https://zenn.dev/hackathons/2024-google-cloud-japan-ai-hackathon)

### Vertex AI 関連

- [Vertex AI ドキュメント](https://cloud.google.com/vertex-ai/docs)
  - [サンプルコード](https://cloud.google.com/vertex-ai/docs/samples?language=golang)
  - [Vertex AI Model Garden（公式）](https://console.cloud.google.com/vertex-ai/model-garden)
- [Vertex AI API for Gemini](https://cloud.google.com/vertex-ai/generative-ai/docs/start/quickstarts/quickstart-multimodal?hl=ja)
- [Vertex AIのサンプルコード](https://github.com/GoogleCloudPlatform/golang-samples/tree/main/vertexai)

### その他のGCPサービス

- [Google Cloud 公式ドキュメント](https://cloud.google.com/docs)
- [Cloud Runのサンプルコード](https://github.com/GoogleCloudPlatform/golang-samples/tree/main/run)
- [Cloud Storageのサンプルコード](https://github.com/GoogleCloudPlatform/golang-samples/tree/main/storage)
- [Firebase ドキュメント](https://firebase.google.com/docs)
- [Cloud Billing API](https://cloud.google.com/billing/docs/how-to/notify?hl=ja#cap_disable_billing_to_stop_usage)

### 開発言語

- [Go言語ドキュメント](https://golang.org/doc/)
