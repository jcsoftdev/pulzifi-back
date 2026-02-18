package storage

import (
	"fmt"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/cloudinary"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/minio"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
)

func NewObjectStorage(cfg *config.Config) (repositories.ObjectStorage, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.ObjectStorageProvider))

	switch provider {
	case "", "minio", "s3":
		return minio.NewClient(cfg)
	case "cloudinary":
		return cloudinary.NewClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported object storage provider: %s", cfg.ObjectStorageProvider)
	}
}
