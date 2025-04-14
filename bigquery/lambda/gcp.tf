# GCSバケットの作成
resource "google_storage_bucket" "my_gcs_bucket" {
  name     = "my-gcs-bucket"
  location = "US"
}

# S3からGCSへのデータ転送
resource "google_storage_transfer_job" "s3_to_gcs_transfer" {
  name     = "s3-to-gcs-transfer"
  project  = "my-gcp-project-id"
  status   = "ENABLED"
  transfer_spec {
    aws_s3_data_source {
      bucket_name = "my-glue-output-bucket"
    }
    gcs_data_sink {
      bucket_name = google_storage_bucket.my_gcs_bucket.name
    }
  }
}

# BigQueryテーブルの作成
resource "google_bigquery_table" "my_bigquery_table" {
  dataset_id = "my_dataset"
  table_id   = "my_table"
  schema     = jsonencode([
    {
      name = "column1"
      type = "STRING"
    },
    {
      name = "column2"
      type = "INTEGER"
    }
  ])
}

# GCSからBigQueryにデータをインポートするジョブ
resource "google_bigquery_job" "load_data_from_gcs_to_bq" {
  project = "my-gcp-project-id"
  configuration {
    load {
      source_uris = [
        "gs://my-gcs-bucket/*.csv"
      ]
      destination_table {
        dataset_id = google_bigquery_table.my_bigquery_table.dataset_id
        table_id   = google_bigquery_table.my_bigquery_table.table_id
      }
      source_format = "CSV"
    }
  }
}
