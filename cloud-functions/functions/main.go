package function

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

var (
	bucketName   = os.Getenv("GCS_BUCKET")   // 例: "your-bucket"
	projectID    = os.Getenv("GCP_PROJECT")  // 環境変数か Cloud Function が自動補完
	bqDataset    = os.Getenv("BQ_DATASET")   // 例: "rawdata"
	dateFormat   = "2006/01/02"              // GCS のパス日付形式
	parquetRegex = regexp.MustCompile(`export/auroradb/([^.]+)\.([^/]+)/parquet/(.+\.parquet)$`)
)

func ProcessTodayFiles(ctx context.Context, _ PubSubMessage) error {
	today := time.Now().Format("2006/01/02") // 例: 2025/07/05 → "2025/07/05"
	prefix := today + "/"                    // GCS 上のプレフィックスを作成

	log.Printf("Scanning bucket '%s' with prefix '%s'", bucketName, prefix)

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer storageClient.Close()

	bqClient, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create bigquery client: %w", err)
	}
	defer bqClient.Close()

	it := storageClient.Bucket(bucketName).Objects(ctx, &storage.Query{Prefix: prefix})
	processed := 0

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error listing objects: %w", err)
		}

		objectName := attrs.Name
		if !parquetRegex.MatchString(objectName) {
			continue
		}

		matches := parquetRegex.FindStringSubmatch(objectName)
		schema := matches[1]
		table := matches[2]
		fileName := matches[3]

		fullURI := fmt.Sprintf("gs://%s/%s", bucketName, objectName)
		bqTableID := fmt.Sprintf("%s_%s", schema, table)

		log.Printf("Loading %s into BigQuery table %s.%s", fullURI, bqDataset, bqTableID)

		gcsRef := bigquery.NewGCSReference(fullURI)
		gcsRef.SourceFormat = bigquery.Parquet
		gcsRef.AutoDetect = true

		loader := bqClient.Dataset(bqDataset).Table(bqTableID).LoaderFrom(gcsRef)
		loader.WriteDisposition = bigquery.WriteAppend

		job, err := loader.Run(ctx)
		if err != nil {
			log.Printf("Load failed for %s: %v", fullURI, err)
			continue
		}

		status, err := job.Wait(ctx)
		if err != nil {
			log.Printf("Job error: %v", err)
			continue
		}
		if err := status.Err(); err != nil {
			log.Printf("BQ load error: %v", err)
			continue
		}

		log.Printf("Successfully loaded: %s", fileName)
		processed++
	}

	log.Printf("Done. Processed %d files.", processed)
	return nil
}

// Pub/Sub trigger stub
type PubSubMessage struct {
	Data []byte `json:"data"`
}
