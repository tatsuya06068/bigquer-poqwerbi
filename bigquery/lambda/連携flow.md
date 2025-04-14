```
[RDS]
  ↓（JDBC接続）
[AWS Glueジョブ（ETL）]
  ↓（CSV形式）
[S3バケット]
  ↓（転送）
[GCSバケット]
  ↓（取り込み）
[BigQueryテーブル]
```

|ステップ	|サービス	|説明|
|-|-|-|
|データ抽出|	AWS Glue|	RDSから直接ETL（SQLも書ける）|
|ストレージ|	Amazon S3|	Glue出力先としてCSV保存|
|転送|	GCS Transfer（Storage Transfer Service）|	S3→GCSに転送（GUIで設定）|
|インポート|	BigQuery|	GCSのCSVをロードする（自動 or スケジュール）|
