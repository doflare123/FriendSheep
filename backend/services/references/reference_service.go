package references

import (
	"context"
	"errors"
	"math"
	"strings"

	"friendship/models/dto"
)

const (
	DefaultGenrePage  = 1
	DefaultGenreLimit = 50
	MaxGenreLimit     = 100
	DefaultCityPage   = 1
	DefaultCityLimit  = 50
	MaxCityLimit      = 100
)

var (
	ErrInvalidGenrePage  = errors.New("параметр page должен быть больше 0")
	ErrInvalidGenreLimit = errors.New("параметр limit должен быть от 1 до 100")
	ErrInvalidCityPage   = errors.New("параметр page должен быть больше 0")
	ErrInvalidCityLimit  = errors.New("параметр limit должен быть от 1 до 100")
)

type ReferenceItem struct {
	ID   uint
	Name string
}

type ActionReferenceItem struct {
	ID   uint
	Code string
	Name string
}

type ReferenceSnapshot struct {
	EventTypes       []ReferenceItem
	Locations        []ReferenceItem
	AgeLimits        []ReferenceItem
	Statuses         []ReferenceItem
	GroupCategories  []ReferenceItem
	GroupActionTypes []ActionReferenceItem
}

type GenreSearchQuery struct {
	Query string
	Page  int
	Limit int
}

type GenreSearchPage struct {
	Items []ReferenceItem
	Total int64
}

type GenreSearchInput struct {
	Query string
	Page  int
	Limit int
}

type CityReferenceItem struct {
	ID     uint
	Name   string
	Region string
}

type CitySearchQuery struct {
	Query string
	Page  int
	Limit int
}

type CitySearchPage struct {
	Items []CityReferenceItem
	Total int64
}

type CitySearchInput struct {
	Query string
	Page  int
	Limit int
}

type ReferenceStore interface {
	LoadReferences(ctx context.Context) (ReferenceSnapshot, error)
	SearchGenres(ctx context.Context, query GenreSearchQuery) (GenreSearchPage, error)
	SearchCities(ctx context.Context, query CitySearchQuery) (CitySearchPage, error)
}

type ReferenceService interface {
	GetReferences(ctx context.Context) (*dto.ReferencesDto, error)
	SearchGenres(ctx context.Context, input GenreSearchInput) (*dto.GenreSearchResponseDto, error)
	SearchCities(ctx context.Context, input CitySearchInput) (*dto.CitySearchResponseDto, error)
}

type referenceService struct {
	store ReferenceStore
}

func NewReferenceService(store ReferenceStore) ReferenceService {
	return &referenceService{store: store}
}

func (s *referenceService) GetReferences(ctx context.Context) (*dto.ReferencesDto, error) {
	snapshot, err := s.store.LoadReferences(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.ReferencesDto{
		EventTypes:       toReferenceDTOs(snapshot.EventTypes),
		Locations:        toReferenceDTOs(snapshot.Locations),
		AgeLimits:        toReferenceDTOs(snapshot.AgeLimits),
		Statuses:         toReferenceDTOs(snapshot.Statuses),
		GroupCategories:  toReferenceDTOs(snapshot.GroupCategories),
		GroupActionTypes: toActionReferenceDTOs(snapshot.GroupActionTypes),
	}, nil
}

func (s *referenceService) SearchGenres(ctx context.Context, input GenreSearchInput) (*dto.GenreSearchResponseDto, error) {
	page := input.Page
	if page == 0 {
		page = DefaultGenrePage
	}
	if page < 1 {
		return nil, ErrInvalidGenrePage
	}

	limit := input.Limit
	if limit == 0 {
		limit = DefaultGenreLimit
	}
	if limit < 1 || limit > MaxGenreLimit {
		return nil, ErrInvalidGenreLimit
	}
	if page > math.MaxInt/limit {
		return nil, ErrInvalidGenrePage
	}

	result, err := s.store.SearchGenres(ctx, GenreSearchQuery{
		Query: strings.TrimSpace(input.Query),
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

	return &dto.GenreSearchResponseDto{
		Items:   toReferenceDTOs(result.Items),
		Total:   result.Total,
		Page:    page,
		Limit:   limit,
		HasMore: int64(page)*int64(limit) < result.Total,
	}, nil
}

func (s *referenceService) SearchCities(ctx context.Context, input CitySearchInput) (*dto.CitySearchResponseDto, error) {
	page := input.Page
	if page == 0 {
		page = DefaultCityPage
	}
	if page < 1 {
		return nil, ErrInvalidCityPage
	}

	limit := input.Limit
	if limit == 0 {
		limit = DefaultCityLimit
	}
	if limit < 1 || limit > MaxCityLimit {
		return nil, ErrInvalidCityLimit
	}
	if page > math.MaxInt/limit {
		return nil, ErrInvalidCityPage
	}

	result, err := s.store.SearchCities(ctx, CitySearchQuery{
		Query: strings.TrimSpace(input.Query),
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

	return &dto.CitySearchResponseDto{
		Items:   toCityReferenceDTOs(result.Items),
		Total:   result.Total,
		Page:    page,
		Limit:   limit,
		HasMore: int64(page)*int64(limit) < result.Total,
	}, nil
}

func toReferenceDTOs(items []ReferenceItem) []dto.ReferenceItemDto {
	result := make([]dto.ReferenceItemDto, 0, len(items))
	for _, item := range items {
		result = append(result, dto.ReferenceItemDto{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return result
}

func toCityReferenceDTOs(items []CityReferenceItem) []dto.CityReferenceItemDto {
	result := make([]dto.CityReferenceItemDto, 0, len(items))
	for _, item := range items {
		result = append(result, dto.CityReferenceItemDto{
			ID:     item.ID,
			Name:   item.Name,
			Region: item.Region,
		})
	}
	return result
}

func toActionReferenceDTOs(items []ActionReferenceItem) []dto.ActionReferenceItemDto {
	result := make([]dto.ActionReferenceItemDto, 0, len(items))
	for _, item := range items {
		result = append(result, dto.ActionReferenceItemDto{
			ID:   item.ID,
			Code: item.Code,
			Name: item.Name,
		})
	}
	return result
}
