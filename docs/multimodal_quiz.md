# Gemini Pro Vision による作品解釈生成

## 概要

このドキュメントでは、Vertex AI (Gemini Pro Vision)を使用して、作品に対する代替解釈を生成する実装について説明します。

## モデルの設定

- **モデル**: Gemini Pro Vision
- **エンドポイント**: `gemini-pro-vision`
- **温度パラメータ**: 0.4（創造性と一貫性のバランス）

## プロンプト設計

### 基本構造
```go
prompt := genai.Text(`
作品: [画像]
作者のコメント: "%s"

以下の点を考慮して、異なる視点からの解釈を生成してください：
1. 作品の視覚的要素（構図、色彩、光と影など）
2. 可能な象徴的な意味
3. 社会的または文化的なコンテキスト

生成する解釈は：
- 作品の本質を尊重すること
- 技術的な観察に基づくこと
- 説得力のある代替視点を提供すること
`)
```

### プロンプトの例
```go
input := fmt.Sprintf(prompt, 
    "この作品では、都市の喧騒の中に潜む静寂を表現しました。" +
    "光と影の対比を通じて、現代社会における孤独と共生の両義性を描き出そうと試みています。")
```

## 実装例

```go
func generateInterpretation(ctx context.Context, imageData []byte, authorComment string) (string, error) {
    client, err := genai.NewClient(ctx, projectID, location)
    if err != nil {
        return "", fmt.Errorf("クライアントの作成に失敗: %w", err)
    }
    defer client.Close()

    model := client.GenerativeModel("gemini-pro-vision")
    model.SetTemperature(0.4)

    // 画像の準備
    img := genai.ImageData("jpeg", imageData)
    
    // プロンプトの構築
    prompt := buildPrompt(authorComment)

    // 解釈の生成
    resp, err := model.GenerateContent(ctx, img, prompt)
    if err != nil {
        return "", fmt.Errorf("解釈の生成に失敗: %w", err)
    }

    return extractInterpretation(resp)
}
```

## 応答の処理

```go
func extractInterpretation(resp *genai.GenerateContentResponse) (string, error) {
    if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
        return "", errors.New("応答が空です")
    }

    text, ok := resp.Candidates[0].Content.Parts[0].(string)
    if !ok {
        return "", fmt.Errorf("予期しない応答形式です: %T", resp.Candidates[0].Content.Parts[0])
    }

    return text, nil
}
```

## エラーハンドリング

1. クライアントエラー
   - 認証エラー
   - ネットワークタイムアウト
   - 無効なリクエスト

2. モデルエラー
   - 生成失敗
   - コンテンツフィルタリング
   - クォータ超過

3. 応答処理エラー
   - 空の応答
   - 不正な形式
   - 品質基準未達

## 参考資料

- [Vertex AI Go SDK](https://pkg.go.dev/cloud.google.com/go/vertexai/genai)
- [Gemini Pro Vision ガイド](https://cloud.google.com/vertex-ai/docs/generative-ai/multimodal/overview)
