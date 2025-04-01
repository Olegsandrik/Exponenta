package repository

import (
	"context"
	"fmt"

	"github.com/Olegsandrik/Exponenta/internal/adapters/minio"
	"github.com/Olegsandrik/Exponenta/internal/repository/errors"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/logger"
)

type ImageRepository struct {
	adapter *minio.Adapter
}

func NewImageRepository(adapter *minio.Adapter) *ImageRepository {
	return &ImageRepository{adapter: adapter}
}

func (ir *ImageRepository) GetImageByID(ctx context.Context, id int, entity string) (models.ImageModel, error) {
	reader, err := ir.adapter.Client.GetObject(
		ctx,
		ir.adapter.BucketName,
		fmt.Sprintf("%s/%d.jpg", entity, id),
		minio.NewEmptyObjectOptions())
	if err != nil {
		logger.Info(ctx, fmt.Sprintf("Error getting image: %v for %s/%d.jpg", err, entity, id))
		return models.ImageModel{}, errors.ErrNoFoundImage
	}

	info, err := reader.Stat()
	if err != nil {
		reader.Close()
		logger.Info(ctx, fmt.Sprintf("Error getting image stat: %v for %s/%d.jpg", err, entity, id))
		return models.ImageModel{}, errors.ErrNoFoundImage
	}

	return models.ImageModel{
		ImageSize:   info.Size,
		Image:       reader,
		ContentType: "image/jpeg",
	}, nil
}
