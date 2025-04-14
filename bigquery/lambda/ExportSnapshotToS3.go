package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"log"
)

type Event struct {
	SnapshotArn string `json:"snapshot_arn"`
	S3Bucket    string `json:"s3_bucket"`
	RoleArn     string `json:"role_arn"`
	KmsKeyId    string `json:"kms_key_id"`
}

func HandleRequest(event Event) (string, error) {
	sess := session.Must(session.NewSession())
	rdsSvc := rds.New(sess)

	exportTaskInput := &rds.StartExportTaskInput{
		ExportTaskIdentifier: aws.String("rds-snapshot-to-s3"),
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

	return fmt.Sprintf("Export started successfully for snapshot %s to S3 bucket %s", event.SnapshotArn, event.S3Bucket), nil
}

func main() {
	lambda.Start(HandleRequest)
}
