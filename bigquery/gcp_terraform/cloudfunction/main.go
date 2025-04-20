package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
)

func Handler(ctx context.Context, e storage.ObjectAttrs) error {
	log.Printf("New file: %s", e.Name)

	projectID := os.Getenv("GCP_PROJECT")
	datasetID := os.Getenv("BQ_DATASET")

	// ファイル名からテーブルを推測（例: table1.csv）
	parts := strings.Split(e.Name, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid file format")
	}
	tableID := parts[0]

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	gcsRef := bigquery.NewGCSReference(fmt.Sprintf("gs://%s/%s", e.Bucket, e.Name))
	gcsRef.SkipLeadingRows = 1
	gcsRef.SourceFormat = bigquery.CSV

	loader := client.Dataset(datasetID).Table(tableID).LoaderFrom(gcsRef)
	job, err := loader.Run(ctx)
	if err != nil {
		return err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	if err := status.Err(); err != nil {
		return err
	}
	log.Printf("Loaded %s into %s.%s", e.Name, datasetID, tableID)
	return nil
}

// cd cloudfunction
// GOARCH=amd64 GOOS=linux go build -o main
// zip function-source.zip main