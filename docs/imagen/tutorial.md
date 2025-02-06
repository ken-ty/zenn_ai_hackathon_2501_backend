# 画像生成コマンド

このセクションでは、画像生成に使用するコマンドと必要な設定について説明します。

## 必要な設定

- **モデルの有効化**: モデルガーデンでモデルを有効化してください。以下のリンクからアクセスできます。
  - [Imagen for Editing and Customization](https://console.cloud.google.com/vertex-ai/publishers/google/model-garden/imagen-3.0-capability-001)
  - モデルID: `publishers/google/models/imagen-3.0-capability-001`
  - バージョン名: `google/imagen-3.0-capability`

- **価格情報**: 価格については[Vertex AIの価格ページ](https://cloud.google.com/vertex-ai/pricing?hl=ja)をご確認ください。

## 使用したコマンド

画像生成テストコマンド（Goではなく、直接GCPのAI PlatformのAPIを使用しています）

```bash
```bash
curl -X POST -H "Authorization: Bearer $(gcloud auth print-access-token)" -H "Content-Type: application/json; charset=utf-8" -d @request.json "https://us-central1-aiplatform.googleapis.com/v1/projects/zenn-ai-hackathon-2501/locations/us-central1/publishers/google/models/imagegeneration@006:predict" | jq -r '.predictions[0].bytesBase64Encoded' | base64 -d > generated_image.png
```

## 参考資料

- [A Developer's Guide to Imagen 3 on Vertex AI](https://cloud.google.com/blog/products/ai-machine-learning/a-developers-guide-to-imagen-3-on-vertex-ai?e=0?utm_source%3Dlinkedin&hl=en)
