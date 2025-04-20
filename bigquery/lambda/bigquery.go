package function

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
)

type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

// テーブル名のマッピング
var tableMap = map[string]string{
	"users":  "your-project.your_dataset.users",
	"orders": "your-project.your_dataset.orders",
	"logs":   "your-project.your_dataset.logs",
}

func GCSHandler(ctx context.Context, e GCSEvent) error {
	log.Printf("Processing file: gs://%s/%s", e.Bucket, e.Name)

	// テーブル名のプレフィックス判定
	var tableID string
	for prefix, id := range tableMap {
		if strings.HasPrefix(e.Name, prefix) {
			tableID = id
			break
		}
	}

	if tableID == "" {
		log.Printf("No matching table found for file: %s", e.Name)
		return nil
	}

	uri := fmt.Sprintf("gs://%s/%s", e.Bucket, e.Name)

	// BigQuery クライアント作成
	client, err := bigquery.NewClient(ctx, "your-project")
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	gcsRef := bigquery.NewGCSReference(uri)
	gcsRef.SourceFormat = bigquery.CSV
	gcsRef.SkipLeadingRows = 1 // ヘッダー行がある場合

	loader := client.DatasetInProject("your-project", "your_dataset").Table(tableID[strings.LastIndex(tableID, ".")+1:]).LoaderFrom(gcsRef)
	loader.WriteDisposition = bigquery.WriteAppend

	job, err := loader.Run(ctx)
	if err != nil {
		return fmt.Errorf("job.Run: %v", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("job.Wait: %v", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("load job failed: %v", err)
	}

	log.Printf("Loaded data into %s successfully.", tableID)
	return nil
}
