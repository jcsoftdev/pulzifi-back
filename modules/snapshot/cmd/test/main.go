package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/application"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/minio"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
)

func main() {
	url := "https://escale.minedu.gob.pe/"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	cfg := config.Load()
	minioClient, err := minio.NewClient(cfg)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := minioClient.EnsureBucket(context.Background()); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	service := application.NewSnapshotService(minioClient)
	req := entities.SnapshotRequest{
		PageID:     "test-page",
		URL:        url,
		SchemaName: os.Getenv("TEST_SCHEMA"),
	}
	res, err := service.CaptureAndUpload(context.Background(), req)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("snapshot uploaded:", res.ImageURL)
}
