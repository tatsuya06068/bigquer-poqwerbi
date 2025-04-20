package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "cloud.google.com/go/storage"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/rds"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Event struct {
    SnapshotArn string `json:"snapshot_arn"`
    S3Bucket    string `json:"s3_bucket"`
    RoleArn     string `json:"role_arn"`
    KmsKeyId    string `json:"kms_key_id"`
    GCSBucket   string `json:"gcs_bucket"`
}

func HandleRequest(event Event) (string, error) {
    // Step 1: Export RDS Snapshot to S3
	log.Println("Starting export of RDS snapshot to S3...")
    sess := session.Must(session.NewSession())
    rdsSvc := rds.New(sess)

    exportTaskInput := &rds.StartExportTaskInput{
        ExportTaskIdentifier: aws.String(fmt.Sprintf("rds-snapshot-to-s3-%d", time.Now().Unix())),
        SourceArn:            aws.String(event.SnapshotArn),
        S3BucketName:         aws.String(event.S3Bucket),
        IamRoleArn:           aws.String(event.RoleArn),
        KmsKeyId:             aws.String(event.KmsKeyId),
    }

    _, err := rdsSvc.StartExportTask(exportTaskInput)
    if err != nil {
        log.Fatalf("Failed to start export task: %v", err)
        return "", fmt.Errorf("failed to start export task: %v", err)
    }

    log.Println("Export task started successfully. Waiting for completion...")
    // Note: You may need to implement a polling mechanism here to wait for the export task to complete.

    // Step 2: Transfer data from S3 to GCS
	log.Println("Transferring data from S3 to GCS...")
    s3Downloader := s3manager.NewDownloader(sess)
    tmpFile, err := os.CreateTemp("", "snapshot-*.parquet")
    if err != nil {
        log.Fatalf("Failed to create temporary file: %v", err)
        return "", fmt.Errorf("failed to create temporary file: %v", err)
    }
    defer os.Remove(tmpFile.Name())

    _, err = s3Downloader.Download(tmpFile, &s3manager.GetObjectInput{
        Bucket: aws.String(event.S3Bucket),
        Key:    aws.String("exported-snapshot.parquet"), // Replace with the actual key
    })
    if err != nil {
        log.Fatalf("Failed to download file from S3: %v", err)
        return "", fmt.Errorf("failed to download file from S3: %v", err)
    }

    log.Println("File downloaded from S3 successfully.")

    // Upload to GCS
	log.Println("Uploading file to GCS...")
    ctx := context.Background()
    gcsClient, err := storage.NewClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create GCS client: %v", err)
        return "", fmt.Errorf("failed to create GCS client: %v", err)
    }
    defer gcsClient.Close()

    gcsObject := gcsClient.Bucket(event.GCSBucket).Object("exported-snapshot.parquet")
    writer := gcsObject.NewWriter(ctx)
    defer writer.Close()

    tmpFile.Seek(0, 0) // Reset file pointer
    if _, err := tmpFile.WriteTo(writer); err != nil {
        log.Fatalf("Failed to upload file to GCS: %v", err)
        return "", fmt.Errorf("failed to upload file to GCS: %v", err)
    }

    log.Println("File uploaded to GCS successfully.")

    return fmt.Sprintf("Exported snapshot %s to GCS bucket %s successfully", event.SnapshotArn, event.GCSBucket), nil
}

func main() {
    lambda.Start(HandleRequest)
}