package dto

import "io"

type Image struct {
	ImageSize   int64
	Image       io.ReadSeeker
	ContentType string
}
