# 開発チュートリアル

このチュートリアルでは、作品解釈クイズサービスの開発方法について説明します。

## 前提条件

1. Google Cloud Projectの設定
   ```bash
   export PROJECT_ID=zenn-ai-hackathon-2501
   export LOCATION=us-central1
   ```

2. 必要なAPIの有効化
   - Vertex AI API
   - Cloud Storage API

3. 認証設定
   ```bash
   gcloud auth application-default login
   ```

## 作品と解釈の登録

### 1. 作品のアップロード

```bash
curl -X POST \
  -F "file=@artwork.jpg" \
  -F "interpretation=この作品では、都市の無機質な表面に映る自然光の反射を通じて、
      現代社会における人工と自然の共生を表現しました。色彩の選択は、
      朝もやに包まれた街並みをイメージし、静謐な雰囲気を醸成しています。" \
  http://localhost:8080/upload
```

### 2. レスポンス例

```json
{
  "quiz_id": "quiz_001",
  "image_path": "/images/artwork.jpg",
  "author_interpretation": "この作品では、都市の無機質な表面に映る自然光の反射を通じて...",
  "ai_interpretation": "都市建築の幾何学的なパターンと光の干渉が生み出す視覚的リズムから、
      現代都市における人間の存在と不在の二重性を読み取ることができます。
      建物のファサードが織りなす影の変化は、都市生活の重層的な時間性を象徴しています。"
}
```

## 作品解釈クイズの取得

### 1. クイズデータの取得

```bash
curl http://localhost:8080/quiz/quiz_001
```

### 2. レスポンス例

```json
{
  "quiz_id": "quiz_001",
  "image_path": "/images/artwork.jpg",
  "interpretations": [
    "この作品では、都市の無機質な表面に映る自然光の反射を通じて...",
    "都市建築の幾何学的なパターンと光の干渉が生み出す視覚的リズムから..."
  ]
}
```

## 開発のポイント

1. 作品解釈について
   - 視覚的要素の具体的な説明
   - 制作意図の明確な表現
   - 技法や表現手法への言及

2. AIによる解釈生成
   - 作品の文脈を考慮した分析
   - 多角的な視点からの考察
   - 技術的観察に基づく解釈

3. エラーハンドリング
   - 画像フォーマットの検証
   - 解釈テキストの品質検証
   - AIサービスの応答エラーへの対応

## 参考資料

- [Vertex AI Gemini API リファレンス](https://cloud.google.com/vertex-ai/docs/reference)
- [マルチモーダルAIの基礎](https://cloud.google.com/blog/products/ai-machine-learning)
