provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_storage_bucket" "upload_bucket" {
  name     = var.gcs_bucket_name
  location = var.region
}

resource "google_bigquery_dataset" "dataset" {
  dataset_id = var.bq_dataset
  location   = var.region
}

# 複数のテーブルを定義したい場合は locals で管理
locals {
  bq_tables = ["table1", "table2"]
}

# Cloud Function 用サービスアカウント
resource "google_service_account" "function_sa" {
  account_id   = "gcs-to-bq-sa"
  display_name = "Cloud Function GCS to BigQuery"
}

# Cloud Function ソースコード zip をアップロード
resource "google_storage_bucket_object" "function_source" {
  name   = "function-source.zip"
  bucket = google_storage_bucket.upload_bucket.name
  source = "${path.module}/cloudfunction/function-source.zip"
}

resource "google_cloudfunctions2_function" "gcs_trigger" {
  name        = "gcs-to-bq"
  location    = var.region
  build_config {
    runtime     = "go121"
    entry_point = "Handler"
    source {
      storage_source {
        bucket = google_storage_bucket.upload_bucket.name
        object = google_storage_bucket_object.function_source.name
      }
    }
  }

  service_config {
    available_memory   = "256M"
    timeout_seconds    = 60
    service_account_email = google_service_account.function_sa.email
  }

  event_trigger {
    event_type = "google.cloud.storage.object.v1.finalized"
    resource   = google_storage_bucket.upload_bucket.name
  }
}

# IAM 権限付与（BigQuery, Storage）
resource "google_project_iam_member" "function_bigquery" {
  role   = "roles/bigquery.dataEditor"
  member = "serviceAccount:${google_service_account.function_sa.email}"
}

resource "google_project_iam_member" "function_storage" {
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.function_sa.email}"
}
