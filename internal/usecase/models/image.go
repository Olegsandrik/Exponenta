package models

import (
	"io"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
)

type ImageModel struct {
	ImageSize   int64
	Image       io.ReadSeeker
	ContentType string
}

func ConvertImageModelToDto(img ImageModel) dto.Image {
	return dto.Image{
		ImageSize:   img.ImageSize,
		Image:       img.Image,
		ContentType: img.ContentType,
	}
}
