package usecase

import (
	"context"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type ImageRepository interface {
	GetImageByID(ctx context.Context, filename string, entity string) (models.ImageModel, error)
}

type ImageUsecase struct {
	imageRepository ImageRepository
}

func NewImageUsecase(imageRepository ImageRepository) *ImageUsecase {
	return &ImageUsecase{imageRepository: imageRepository}
}

func (iu *ImageUsecase) GetImageByID(ctx context.Context, filename string, entity string) (dto.Image, error) {
	imageModel, err := iu.imageRepository.GetImageByID(ctx, filename, entity)
	if err != nil {
		return dto.Image{}, err
	}
	image := models.ConvertImageModelToDto(imageModel)
	return image, nil
}
