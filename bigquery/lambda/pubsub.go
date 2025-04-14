package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/cloud/storage/v1"
)

func handlePubSubMessage(ctx context.Context, m *pubsub.Message) error {
	// Pub/Subメッセージからファイルの詳細情報を取得
	var fileDetails struct {
		Bucket string `json:"bucket"`
		Name   string `json:"name"`
	}

	if err := json.Unmarshal(m.Data, &fileDetails); err != nil {
		log.Printf("Error unmarshalling PubSub message: %v", err)
		return err
	}

	// BigQueryの設定
	projectID := "YOUR_PROJECT_ID"
	datasetID := "YOUR_BIGQUERY_DATASET"
	tableID := "YOUR_BIGQUERY_TABLE"

	// GCSからBigQueryにデータをインポート
	err := loadDataToBigQuery(ctx, projectID, datasetID, tableID, fileDetails.Bucket, fileDetails.Name)
	if err != nil {
		log.Printf("Failed to load data to BigQuery: %v", err)
		return err
	}

	return nil
}

// BigQueryにデータをロードする
func loadDataToBigQuery(ctx context.Context, projectID, datasetID, tableID, bucket, object string) error {
	client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile("path/to/credentials.json"))
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	// BigQueryのテーブル参照を作成
	tableRef := client.Dataset(datasetID).Table(tableID)

	// GCSのファイルをBigQueryにロードする
	loadJob := client.Dataset(datasetID).Table(tableID).LoaderFrom(bigquery.NewGCSReference(fmt.Sprintf("gs://%s/%s", bucket, object)))
	loadJob.SourceFormat = bigquery.CSV
	loadJob.SkipLeadingRows = 1 // CSVのヘッダー行をスキップする場合

	// ジョブを実行
	job, err := loadJob.Run(ctx)
	if err != nil {
		return fmt.Errorf("bigquery.Job.Run: %v", err)
	}

	// ジョブの完了を待つ
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("job.Wait: %v", err)
	}
	if status.Err() != nil {
		return fmt.Errorf("job failed: %v", status.Err())
	}

	log.Printf("Data from GCS %s/%s loaded into BigQuery table %s.%s", bucket, object, datasetID, tableID)
	return nil
}
