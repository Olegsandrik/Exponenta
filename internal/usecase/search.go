package usecase

import (
	"context"
	"github.com/Olegsandrik/Exponenta/utils"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type SearchRepo interface {
	Search(ctx context.Context, query string) (models.SearchResponseModel, error)
	Suggest(ctx context.Context, query string) (models.SuggestResponseModel, error)
}

type SearchUsecase struct {
	searchRepo SearchRepo
}

func NewSearchUsecase(searchRepo SearchRepo) *SearchUsecase {
	return &SearchUsecase{searchRepo: searchRepo}
}

func (s *SearchUsecase) Search(ctx context.Context, query string) (dto.SearchResponseDto, error) {
	searchResultModel, err := s.searchRepo.Search(ctx, query)

	if searchResultModel.Recipes != nil {
		utils.SanitizeRecipeDescription(searchResultModel.Recipes)
	}

	if err != nil {
		return dto.SearchResponseDto{}, err
	}

	searchResult := models.ConvertSearchResponseToDto(searchResultModel)

	return searchResult, nil
}

func (s *SearchUsecase) Suggest(ctx context.Context, query string) (dto.SuggestResponseDto, error) {
	suggestResultModel, err := s.searchRepo.Suggest(ctx, query)

	if err != nil {
		return dto.SuggestResponseDto{}, err
	}

	suggestResult := models.ConvertSuggestResponseToDto(suggestResultModel)

	return suggestResult, nil
}
