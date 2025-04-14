package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	S3Bucket string `json:"s3_bucket"`
	S3Key    string `json:"s3_key"`
	GcsBucket string `json:"gcs_bucket"`
}

func HandleRequest(event Event) (string, error) {
	// Google Cloud SDKがインストールされていることを前提として、gsutilを使ってS3からGCSに転送
	// gsutilがLambda環境にインストールされている必要があります

	// gsutilコマンドを構築
	sourceS3Path := fmt.Sprintf("s3://%s/%s", event.S3Bucket, event.S3Key)
	destinationGCSPath := fmt.Sprintf("gs://%s/%s", event.GcsBucket, event.S3Key)

	cmd := exec.Command("gsutil", "cp", sourceS3Path, destinationGCSPath)
	cmd.Env = append(os.Environ(), "GOOGLE_APPLICATION_CREDENTIALS=/tmp/gcp-credentials.json")

	// コマンドを実行
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to transfer from S3 to GCS: %v", err)
		return "", fmt.Errorf("failed to transfer from S3 to GCS: %s", output)
	}

	return fmt.Sprintf("Successfully transferred %s from S3 to GCS", event.S3Key), nil
}

func main() {
	lambda.Start(HandleRequest)
}
