package usecase

import (
	"context"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type SearchRepo interface {
	Search(ctx context.Context, query string) (models.SearchResponseModel, error)
}

type SearchUsecase struct {
	searchRepo SearchRepo
}

func NewSearchUsecase(searchRepo SearchRepo) *SearchUsecase {
	return &SearchUsecase{searchRepo: searchRepo}
}

func (s *SearchUsecase) Search(ctx context.Context, query string) (dto.SearchResponseDto, error) {
	searchResultModel, err := s.searchRepo.Search(ctx, query)

	if err != nil {
		return dto.SearchResponseDto{}, err
	}

	searchResult := models.ConvertSearchResponseToDto(searchResultModel)

	return searchResult, nil
}
