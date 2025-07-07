provider "google" {
  project = var.project_id
  region  = var.region
}

# GCS バケット for Function ソースコード
resource "google_storage_bucket" "function_source" {
  name     = "${var.project_id}-function-code"
  location = var.region
}

# アップロードするアーカイブ（zip 化は事前にローカルで行う or 外部デプロイ）
resource "google_storage_bucket_object" "function_zip" {
  name   = "function-source.zip"
  bucket = google_storage_bucket.function_source.name
  source = "function/function-source.zip" # zip ファイルのパス
}

# Pub/Sub トピック（Scheduler → Function）
resource "google_pubsub_topic" "scheduler_topic" {
  name = "function-scheduler-topic"
}

# Cloud Function
resource "google_cloudfunctions2_function" "my_function" {
  name        = "parquet-loader"
  location    = var.region
  description = "Load parquet to BigQuery"

  build_config {
    runtime     = "go120"
    entry_point = "LoadParquetToBigQuery"
    source {
      storage_source {
        bucket = google_storage_bucket.function_source.name
        object = google_storage_bucket_object.function_zip.name
      }
    }
  }

  service_config {
    available_memory   = "256M"
    timeout_seconds    = 60
    ingress_settings   = "ALLOW_ALL"
    max_instance_count = 1
    environment_variables = {
      BQ_TABLE_ID = var.bq_table_id
    }
    pubsub_topic = google_pubsub_topic.scheduler_topic.id
  }

  depends_on = [google_storage_bucket_object.function_zip]
}

# Scheduler job（毎日午前5時に実行）
resource "google_cloud_scheduler_job" "daily_trigger" {
  name             = "trigger-function-daily"
  description      = "Triggers the Cloud Function daily"
  schedule         = "0 5 * * *"
  time_zone        = "Asia/Tokyo"

  pubsub_target {
    topic_name = google_pubsub_topic.scheduler_topic.id
    data       = base64encode("run") # Payload (base64)
  }
}

# Cloud Function 実行者に Pub/Sub → Invoker 許可
resource "google_cloudfunctions2_function_iam_member" "invoker" {
  project        = var.project_id
  region         = var.region
  cloud_function = google_cloudfunctions2_function.my_function.name

  role   = "roles/cloudfunctions.invoker"
  member = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-cloudscheduler.iam.gserviceaccount.com"
}

# 認証情報（Cloud Scheduler → Pub/Sub）関連
data "google_project" "project" {}
