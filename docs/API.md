# API仕様書

## 概要

このドキュメントでは、AI Art Quizアプリケーションが提供するAPIの詳細な仕様を説明します。

## エンドポイント

### 1. クイズ作成 API

作品画像と作者の解釈をアップロードし、新しいクイズを作成します。

```yaml
POST /upload
Content-Type: multipart/form-data

リクエストパラメータ:
  - file: バイナリ
    必須: true
    説明: アップロードする作品画像（JPEG/PNG形式）
    最大サイズ: 32MB

  - interpretation: string
    必須: true
    説明: 作者による作品の解釈
    最大長: 1000文字

レスポンス (200 OK):
{
    "id": "quiz_1234567890",
    "image_path": "/images/artwork.jpg",
    "author_interpretation": "作者による解釈のテキスト",
    "ai_interpretation": "AIによる代替解釈のテキスト",
    "created_at": "2024-03-20T10:00:00Z"
}

エラーレスポンス:
- 400 Bad Request:
  - 画像データが不正
  - 解釈テキストが空
  - ファイルサイズが上限を超過

- 500 Internal Server Error:
  - AIサービスとの通信エラー
  - ストレージへの保存エラー
```

### 2. クイズ取得 API

指定されたIDのクイズを取得し、ランダム化された解釈を返します。

```yaml
GET /quiz/{quiz_id}
Content-Type: application/json

パスパラメータ:
  - quiz_id: string
    必須: true
    説明: 取得するクイズのID
    形式: "quiz_" + 数値

レスポンス (200 OK):
{
    "id": "quiz_1234567890",
    "image_path": "/images/artwork.jpg",
    "interpretations": [
        "解釈1",  // 作者かAIの解釈がランダムに配置
        "解釈2"
    ]
}

エラーレスポンス:
- 400 Bad Request:
  - 無効なクイズID形式

- 404 Not Found:
  - 指定されたクイズが存在しない

- 500 Internal Server Error:
  - ストレージからの読み込みエラー
```

## 共通仕様

### リクエストヘッダー

```yaml
必須ヘッダー:
  - Content-Type: 
    - multipart/form-data (POST /upload)
    - application/json (GET /quiz/{quiz_id})
```

### エラーレスポンス形式

```json
{
    "error": {
        "code": "ERROR_CODE",
        "message": "エラーの詳細メッセージ"
    }
}
```

### 制限事項

1. レート制限
   - 1分あたり60リクエスト
   - 超過した場合は429 Too Many Requestsを返却

2. ファイルサイズ
   - 画像ファイル: 最大32MB
   - 解釈テキスト: 最大1000文字

3. 対応画像フォーマット
   - JPEG
   - PNG

## セキュリティ

1. 入力検証
   - ファイルタイプの検証
   - テキストの長さ制限
   - XSS対策

2. エラーハンドリング
   - スタックトレースは非公開
   - ユーザーフレンドリーなエラーメッセージ

## 開発者向け情報

### テスト用エンドポイント

開発環境では、以下のテスト用エンドポイントが利用可能です：

```bash
# ヘルスチェック
GET /health

レスポンス:
{
    "status": "ok",
    "version": "1.0.0"
}
```

### デバッグモード

環境変数`DEBUG=true`を設定すると、詳細なエラー情報が返却されます。

### 推奨クライアント実装

```go
// クイズ作成
func CreateQuiz(imageData []byte, interpretation string) (*Quiz, error) {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    // ファイルの追加
    part, err := writer.CreateFormFile("file", "artwork.jpg")
    if err != nil {
        return nil, err
    }
    part.Write(imageData)
    
    // 解釈の追加
    writer.WriteField("interpretation", interpretation)
    writer.Close()
    
    // リクエストの送信
    resp, err := http.Post("/upload", writer.FormDataContentType(), body)
    // エラー処理とレスポンスのパース
}
``` 
