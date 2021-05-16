# umaevent-server
イベント名と下3つ分の選択肢の合計4つの画像を送りつけると可能性が高そうなイベントのJSONを返してくれるサーバ

## Usage
- `$EVENT_DATA_ENDPOINT` にイベントデータを返してくれるエンドポイントを指定
- `$OCR_SERVER_ENDPOINT` にOCRサーバのエンドポイントを指定(https://github.com/otiai10/ocrserver の /file)
- `$PORT` にポート番号指定(optional, default: 8080)
- make run で実行
