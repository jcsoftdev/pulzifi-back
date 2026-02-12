package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/application"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/minio"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

func main() {
	url := "https://escale.minedu.gob.pe/"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		fmt.Println("error connecting to db:", err)
		os.Exit(1)
	}

	minioClient, err := minio.NewClient(cfg)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := minioClient.EnsureBucket(context.Background()); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	extractorClient := extractor.NewHTTPClient(cfg.ExtractorURL)

	service := application.NewSnapshotService(minioClient, extractorClient, db)
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
