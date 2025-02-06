# 1. Google Cloudの設定

## Step 1: GCPプロジェクトの設定

```bash
# GCPプロジェクトの設定
gcloud config set project zenn-ai-hackathon-2501
gcloud config set run/region us-east1
### 3. 必要なGoogle Cloudサービスの有効化
# 必要なAPIの有効化
gcloud services enable \
  run.googleapis.com \
  aiplatform.googleapis.com \
  storage.googleapis.com
```

## Step 2: Cloud Storage のセットアップ

### バケットの作成

```bash
gcloud storage buckets create gs://zenn-ai-hackathon-2501 \
  --location=us-east1 \
  --uniform-bucket-level-access

# サブディレクトリの作成（メタデータ用）
echo '{"questions":[]}' | gcloud storage cp - gs://zenn-ai-hackathon-2501/metadata/questions.json

# 画像用ディレクトリの作成を確認
gcloud storage ls gs://zenn-ai-hackathon-2501/
```

### 画像用のディレクトリを作成

```bash
# original と generated ディレクトリを作成
echo "" | gcloud storage cp - gs://zenn-ai-hackathon-2501/original/.keep
echo "" | gcloud storage cp - gs://zenn-ai-hackathon-2501/generated/.keep

# 構造を確認
gcloud storage ls -r gs://zenn-ai-hackathon-2501/
```

## Step 3: 認証設定

```bash
# サービスアカウントの作成
gcloud iam service-accounts create zenn-ai-backend \
  --display-name="Zenn AI Backend Service Account"

# 必要な権限の付与
gcloud projects add-iam-policy-binding zenn-ai-hackathon-2501 \
  --member="serviceAccount:zenn-ai-backend@zenn-ai-hackathon-2501.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding zenn-ai-hackathon-2501 \
  --member="serviceAccount:zenn-ai-backend@zenn-ai-hackathon-2501.iam.gserviceaccount.com" \
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding zenn-ai-hackathon-2501 \
  --member="serviceAccount:zenn-ai-backend@zenn-ai-hackathon-2501.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# キーファイルの作成とダウンロード
gcloud iam service-accounts keys create config/credentials/keyfile.json \
  --iam-account=zenn-ai-backend@zenn-ai-hackathon-2501.iam.gserviceaccount.com
```


## Step 4: ローカル開発環境のセットアップ

```bash
# 環境変数の設定
export GOOGLE_APPLICATION_CREDENTIALS="$(pwd)/config/credentials/keyfile.json"

# 依存関係のインストール
go mod download

# ローカルサーバーの起動準備
mkdir -p cmd/server
```

## Step 5: ローカルサーバーの実装

- `/cmd/server/main.go` を実装。

```bash
# ローカルサーバーの起動
go run cmd/server/main.go

# 期待する出力
2025/02/06 13:07:23 Server starting on port 8080

# 別ターミナルで確認
curl http://localhost:8080/health
OK                     
```

## Step 6: Cloud Storageとの連携

- `/internal/storage/client.go` を実装。
- `/cmd/server/main.go:main()` にCloud Storageの初期化処理を追加。

```bash
# ローカルサーバーの起動
go run cmd/server/main.go

# 期待する出力
2025/02/06 13:07:23 Server starting on port 8080

# 別ターミナルで確認
curl http://localhost:8080/health
OK
```

## Step 7: データモデルの実装

- `/internal/models/types.go` を実装。
- `/cmd/server/main.go:uploadHandler()` にCloud Storageにアップロードする処理を追加。

```bash
# ローカルサーバーの起動
go run cmd/server/main.go

# 期待する出力
2025/02/06 13:07:23 Server starting on port 8080

# 別ターミナルで確認
screencapture -t jpg test.jpg 
curl -X POST -F "file=@test.jpg" http://localhost:8080/upload
{"image_id":"image_1738815217","storage_url":"gs://zenn-ai-hackathon-2501/original/image_1738815217.jpg","status":"success"}
```

## Step 8: クイズ情報を取得する機能の実装

- `/cmd/server/main.go:questionsHandler()` にクイズ情報取得処理を実装。

```bash
# ローカルサーバーの起動
go run cmd/server/main.go

# 期待する出力
2025/02/06 13:07:23 Server starting on port 8080

# 別ターミナルで確認
curl http://localhost:8080/questions
{"questions":[]}
```

## Step 9: アップロードされた画像をクイズデータとして登録する機能の実装

- `/internal/storage/client.go:UpdateQuestions()` を追加。
- `/cmd/server/main.go:uploadHandler()` にクイズデータを登録する処理を追加。

```bash
# ローカルサーバーの起動
go run cmd/server/main.go

# 期待する出力
2025/02/06 16:23:05 Server starting on port 8080

# 別ターミナルで確認 1
curl -X POST -F "file=@test.jpg" http://localhost:8080/upload

{"image_id":"image_1738826600","storage_url":"gs://zenn-ai-hackathon-2501/original/image_1738826600.jpg","status":"success"}

# 別ターミナルで確認 2
curl http://localhost:8080/questions
{"questions":[{"id":"image_1738826600","original_image":"original/image_1738826600.jpg","fake_images":[],"correct_index":0,"created_at":"2025-02-06T07:23:23Z"}]}
```

## Step 10: Vertex AIを使用して偽画像を生成する機能の実装

- `/internal/ai/client.go` を実装。
- `/cmd/server/main.go:aiClient` を初期化。

これはまだ実行できません。

