package usecase

import (
	"context"

	"github.com/Olegsandrik/Exponenta/internal/delivery/dto"
	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
	"github.com/Olegsandrik/Exponenta/utils"
)

type SearchRepo interface {
	Search(ctx context.Context, query string, diet string, dishType string,
		maxTime int) (models.SearchResponseModel, error)
	Suggest(ctx context.Context, query string) (models.SuggestResponseModel, error)
	GetDishTypes(ctx context.Context) ([]string, error)
	GetDiets(ctx context.Context) ([]string, error)
	GetMaxMinCookingTime(ctx context.Context) (models.TimeModel, error)
}

type SearchUsecase struct {
	searchRepo SearchRepo
}

func NewSearchUsecase(searchRepo SearchRepo) *SearchUsecase {
	return &SearchUsecase{searchRepo: searchRepo}
}

func (s *SearchUsecase) Search(
	ctx context.Context, query string, diet string, dishType string, maxTime int) (dto.SearchResponseDto, error) {
	searchResultModel, err := s.searchRepo.Search(ctx, query, diet, dishType, maxTime)

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

func (s *SearchUsecase) GetFilter(ctx context.Context) (dto.FiltersDto, error) {
	dishTypesModel, err := s.searchRepo.GetDishTypes(ctx)

	if err != nil {
		return dto.FiltersDto{}, err
	}

	dietsModel, err := s.searchRepo.GetDiets(ctx)

	if err != nil {
		return dto.FiltersDto{}, err
	}

	timeModel, err := s.searchRepo.GetMaxMinCookingTime(ctx)

	if err != nil {
		return dto.FiltersDto{}, err
	}

	filterDto := models.ConvertFilterModelToDto(dishTypesModel, dietsModel, timeModel)

	return filterDto, nil
}
