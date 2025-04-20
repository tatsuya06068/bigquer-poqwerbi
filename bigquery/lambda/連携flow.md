``
方法	特徴	工数	柔軟性
Lambda + スナップショット	自由に処理可能だがスナップショット復元が必要
Glue + JDBC	コード少なめで定期実行しやすい
DMS（Database Migration Service）	全体レプリケーション向き（テーブル単位はやや工夫必要）
```

```
[RDS]
  ↓（JDBC接続）
[AWS Glueジョブ（ETL）]
  ↓（CSV形式）
[S3バケット]
  ↓（転送）
[GCS バケット]
  ↓
[Cloud Storageイベント（Object Finalize）]
   → [Cloud Functions]
         → インポート対象テーブルを判別
         → BigQuery にロード
```

|ステップ	|サービス	|説明|
|-|-|-|
|データ抽出|	AWS Glue|	RDSから直接ETL（SQLも書ける）|
|ストレージ|	Amazon S3|	Glue出力先としてCSV保存|
|転送|	GCS Transfer（Storage Transfer Service）|	S3→GCSに転送（GUIで設定）|
|インポート|	BigQuery|	GCSのCSVをロードする（自動 or スケジュール）|
