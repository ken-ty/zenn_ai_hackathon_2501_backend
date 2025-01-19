# zenn_ai_hackathon_2501_backend

AI Agent Hackathon with Google Cloud のバックエンド

## FE開発者向け

mock server の起動方法

```bash
# image 作成
docker compose build --no-cache

# 起動
docker compose up

# 停止
docker compose down
```

例:

```bash
curl -X 'GET' 'http://localhost:8080/api/v1/health' -H 'accept: application/json'
curl -X 'GET' 'http://localhost:8080/api/v1/questions?page=1&limit=25&order=latest' -H 'accept: application/json'
curl -X 'GET' 'http://localhost:8080/api/v1/questions?page=1&limit=25&order=latest' -H 'accept: application/json' -H '__example: Example'
```

BE の更新を反映するには、 最新 main を pull して、 上記コマンドを 初めからやり直す

## API開発フロー

このプロジェクトでは、OpenAPI (Swagger) 仕様を使用してAPI定義を管理し、型安全な開発を実現します。

### ディレクトリ構造

```text
backend/
├── api/
│   ├── openapi/           # OpenAPI定義ファイル
│   │   ├── spec/         # API仕様ファイル
│   │   └── generated/    # 生成されたコード
│   └── proto/            # gRPC用のprotoファイル（必要な場合）
├── cmd/
├── internal/
│   ├── handler/          # HTTPハンドラー
│   ├── service/          # ビジネスロジック
│   ├── repository/       # データアクセス層
│   └── mock/            # 生成されたモック
└── pkg/
```

### API定義の作成

1. OpenAPI仕様でAPIを定義

```yaml
# api/openapi/spec/api.yaml
openapi: 3.0.0
info:
  title: AI Hackathon API
  version: 1.0.0
paths:
  /api/v1/example:
    get:
      summary: Example endpoint
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ExampleResponse'
components:
  schemas:
    ExampleResponse:
      type: object
      properties:
        message:
          type: string
```

### 開発ツール

必要なツールをインストール：

```bash
# OpenAPI generator のインストール
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

# モックジェネレーターのインストール
go install github.com/golang/mock/mockgen@latest

# Swagger UI（オプション）
docker pull swaggerapi/swagger-ui
```

### コード生成コマンド

```bash
# OpenAPI定義から型とサーバーコードを生成
oapi-codegen -package api -generate types,server,spec api/openapi/spec/api.yaml > api/openapi/generated/api.gen.go

# インターフェースからモックを生成
mockgen -source=internal/service/service.go -destination=internal/mock/service_mock.go

# Swagger UIの起動（開発時）
docker run -p 8081:8080 -e SWAGGER_JSON=/api.yaml -v $(pwd)/api/openapi/spec/api.yaml:/api.yaml swaggerapi/swagger-ui
```

### Makefileによる自動化

```makefile
.PHONY: generate
generate: generate-api generate-mock

.PHONY: generate-api
generate-api:
    oapi-codegen -package api -generate types,server,spec api/openapi/spec/api.yaml > api/openapi/generated/api.gen.go

.PHONY: generate-mock
generate-mock:
    mockgen -source=internal/service/service.go -destination=internal/mock/service_mock.go

.PHONY: swagger-ui
swagger-ui:
    docker run -d -p 8081:8080 -e SWAGGER_JSON=/api.yaml -v $(pwd)/api/openapi/spec/api.yaml:/api.yaml swaggerapi/swagger-ui

.PHONY: test
test:
    go test -v -race -cover ./...
```

### 使用例

#### 1. APIハンドラーの実装

```go
// internal/handler/example.go
package handler

import (
    "net/http"
    "github.com/your-org/your-project/api/openapi/generated"
)

type ExampleHandler struct {
    service Service
}

func (h *ExampleHandler) GetExample(w http.ResponseWriter, r *http.Request) {
    response := api.ExampleResponse{
        Message: "Hello, World!",
    }
    // レスポンスを返す処理
}
```

#### 2. テストの実装

```go
// internal/handler/example_test.go
package handler_test

import (
    "testing"
    "github.com/golang/mock/gomock"
    "github.com/your-org/your-project/internal/mock"
)

func TestExampleHandler_GetExample(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockService := mock.NewMockService(ctrl)
    // テストケースの実装
}
```

### API開発のベストプラクティス

1. **API設計**
   - APIの変更は必ずOpenAPI仕様から開始
   - バージョニングは `/api/v1/` のようにパスで管理
   - レスポンスは一貫した形式を維持

2. **コード生成**
   - コード生成は `make generate` で一括実行
   - 生成されたコードは直接編集しない
   - CIでコード生成の整合性をチェック

3. **テスト**
   - 生成されたモックを活用
   - エンドポイントごとに統合テストを作成
   - エッジケースのテストを忘れずに実装

4. **ドキュメント**
   - Swagger UIで最新のAPI仕様を確認可能
   - 変更履歴はGitで管理
   - 重要な設計判断は必ずコメントとして残す
