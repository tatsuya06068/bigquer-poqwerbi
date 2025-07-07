variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  default     = "us-central1"
}

variable "bq_table_id" {
  type        = string
  description = "Target BigQuery table ID (e.g., project.dataset.table)"
}
