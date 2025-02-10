# AI Art Quiz - マルチモーダルクイズゲーム

## 概要

このプロジェクトは、作品に込められた作者の意図とAIが生成した代替解釈を比較し、芸術作品への理解を深めるクイズゲームのバックエンドAPIです。

## 特徴

- 視覚と言語を組み合わせたマルチモーダルな芸術体験
- Vertex AI (Gemini Pro Vision)を活用した深い作品解釈の生成

## 技術スタック

- Go 1.21
- Google Cloud Platform
  - Vertex AI (Gemini Pro Vision)
  - Cloud Run
  - Cloud Storage

## 主要機能

1. クイズ作成
   - 作品と作者の解釈のアップロード
   - AIによる代替解釈の生成
   - クイズデータの保存

2. クイズ取得
   - 画像の取得
   - 解釈のランダム表示
   - 結果の検証

## 開発環境のセットアップ

1. 必要なツール
   - Go 1.21以上
   - Google Cloud SDK

2. 環境変数の設定

   ```bash
   export PROJECT_ID=zenn-ai-hackathon-2501
   export LOCATION=asia-northeast1
   ```

3. 依存関係のインストール

   ```bash
   go mod tidy
   ```

## APIの使用方法

詳細は以下のドキュメントを参照してください：

- [チュートリアル](docs/TUTORIAL.md)
- [アーキテクチャ設計](docs/ARCHITECTURE.md)
- [ゲームルール](docs/RULES.md)

## 使用例

### 作品の登録
```bash
curl -X POST \
  -F "file=@artwork.jpg" \
  -F "interpretation=この作品では、都市の無機質な表面に映る自然光の反射を通じて..." \
  http://localhost:8080/upload
```

### クイズの取得
```bash
curl http://localhost:8080/quiz/quiz_001
```

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## 貢献

1. このリポジトリをフォーク
2. 機能開発用のブランチを作成
3. 変更をコミット
4. プルリクエストを作成

## 開発ガイドライン

- 作品解釈の品質を重視
- ユーザー体験の一貫性を保持
- セキュリティとプライバシーの確保
- パフォーマンスの最適化

## 開発者

- [開発者名]
- [連絡先]

## 謝辞

- Google Cloud Japan
- Zenn
- ハッカソン運営チーム
