package repository

import (
	"context"
	"fmt"
	internalErrors "github.com/Olegsandrik/Exponenta/internal/errors"

	"github.com/Olegsandrik/Exponenta/internal/adapters/minio"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/internal/utils"
	"github.com/Olegsandrik/Exponenta/logger"
)

type ImageRepository struct {
	adapter *minio.Adapter
}

func NewImageRepository(adapter *minio.Adapter) *ImageRepository {
	return &ImageRepository{adapter: adapter}
}

func (ir *ImageRepository) GetImageByID(ctx context.Context,
	filename string, entity string) (models.ImageModel, error) {
	reader, err := ir.adapter.Client.GetObject(
		ctx,
		ir.adapter.BucketName,
		fmt.Sprintf("%s/%s", entity, filename),
		minio.NewEmptyObjectOptions())
	if err != nil {
		logger.Info(ctx, fmt.Sprintf("Error getting image: %v for %s/%s", err, entity, filename))
		return models.ImageModel{}, internalErrors.ErrNoFoundImage
	}

	info, err := reader.Stat()
	if err != nil {
		reader.Close()
		logger.Info(ctx, fmt.Sprintf("Error getting image stat: %v for %s/%s", err, entity, filename))
		return models.ImageModel{}, internalErrors.ErrNoFoundImage
	}

	contentType := utils.GetContentType(filename)

	return models.ImageModel{
		ImageSize:   info.Size,
		Image:       reader,
		ContentType: contentType,
	}, nil
}
