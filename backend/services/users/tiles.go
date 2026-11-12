package users

import (
	"context"
	"errors"

	"friendship/services/references"
	"friendship/services/statistics"
)

var (
	ErrInvalidTileSettings  = errors.New("некорректные настройки плиток статистики")
	ErrTileReferenceMissing = errors.New("не найден обязательный тип мероприятия для плитки статистики")
)

type TileKey string

const (
	TileFilms                  TileKey = "films"
	TileGames                  TileKey = "games"
	TileTableGames             TileKey = "tableGames"
	TileOther                  TileKey = "other"
	TileCompletedEvents        TileKey = "completedEvents"
	TilePlannedMinutes         TileKey = "plannedMinutes"
	TileCreatedCompletedEvents TileKey = "createdCompletedEvents"
	TileUniquePeopleCount      TileKey = "uniquePeopleCount"
	TileMostFrequentTypeID     TileKey = "mostFrequentTypeId"
	TileMostFrequentGenreID    TileKey = "mostFrequentGenreId"
	TileMostFrequentWeekday    TileKey = "mostFrequentWeekday"
)

const MaxStatisticTiles = 4

var allTileKeys = map[TileKey]struct{}{
	TileFilms: {}, TileGames: {}, TileTableGames: {}, TileOther: {},
	TileCompletedEvents: {}, TilePlannedMinutes: {}, TileCreatedCompletedEvents: {},
	TileUniquePeopleCount: {}, TileMostFrequentTypeID: {}, TileMostFrequentGenreID: {},
	TileMostFrequentWeekday: {},
}

var eventTypeNameByTile = map[TileKey]string{
	TileFilms:      "Фильмы",
	TileGames:      "Игры",
	TileTableGames: "Настольные игры",
	TileOther:      "Другое",
}

type TileSettings struct {
	Tiles []TileKey `json:"tiles" binding:"required,max=4,unique" maxItems:"4" uniqueItems:"true" enums:"films,games,tableGames,other,completedEvents,plannedMinutes,createdCompletedEvents,uniquePeopleCount,mostFrequentTypeId,mostFrequentGenreId,mostFrequentWeekday"`
}

func DefaultTileSettings() TileSettings {
	return TileSettings{Tiles: []TileKey{TileFilms, TileGames, TileTableGames, TileCompletedEvents}}
}

type StatisticTile struct {
	Key          TileKey       `json:"key" enums:"films,games,tableGames,other,completedEvents,plannedMinutes,createdCompletedEvents,uniquePeopleCount,mostFrequentTypeId,mostFrequentGenreId,mostFrequentWeekday"`
	Value        *int64        `json:"value" extensions:"x-nullable"`
	Unit         string        `json:"unit,omitempty" enums:"minutes"`
	Status       *string       `json:"status,omitempty" enums:"not_calculated,pending,fresh"`
	CalculatedAt *NullableTime `json:"calculatedAt,omitempty" swaggertype:"string" format:"date-time" extensions:"x-nullable"`
}

type TileSettingsStore interface {
	GetTileSettings(context.Context, uint) (TileSettings, error)
	ReplaceTileSettings(context.Context, uint, TileSettings) (TileSettings, error)
	FindEventTypesByNames(context.Context, []string) ([]references.ReferenceItem, error)
}

func (s *ProfileService) GetTileSettings(ctx context.Context, userID uint) (TileSettings, error) {
	if userID == 0 {
		return TileSettings{}, ErrUnauthenticated
	}
	return s.store.GetTileSettings(ctx, userID)
}

func (s *ProfileService) ReplaceTileSettings(ctx context.Context, userID uint, settings TileSettings) (TileSettings, error) {
	if userID == 0 {
		return TileSettings{}, ErrUnauthenticated
	}
	normalized, err := normalizeTileSettings(settings)
	if err != nil {
		return TileSettings{}, err
	}
	return s.store.ReplaceTileSettings(ctx, userID, normalized)
}

func normalizeTileSettings(settings TileSettings) (TileSettings, error) {
	if settings.Tiles == nil || len(settings.Tiles) > MaxStatisticTiles {
		return TileSettings{}, ErrInvalidTileSettings
	}
	seen := make(map[TileKey]struct{}, len(settings.Tiles))
	for _, key := range settings.Tiles {
		if _, ok := allTileKeys[key]; !ok {
			return TileSettings{}, ErrInvalidTileSettings
		}
		if _, duplicate := seen[key]; duplicate {
			return TileSettings{}, ErrInvalidTileSettings
		}
		seen[key] = struct{}{}
	}
	result := make([]TileKey, len(settings.Tiles))
	copy(result, settings.Tiles)
	return TileSettings{Tiles: result}, nil
}

func visibleSelectedTiles(settings TileSettings, visible map[StatisticField]struct{}) []TileKey {
	result := make([]TileKey, 0, len(settings.Tiles))
	for _, key := range settings.Tiles {
		if _, ok := visible[StatisticField(key)]; ok {
			result = append(result, key)
		}
	}
	return result
}

func compactStatisticsRequested(visible map[StatisticField]struct{}) bool {
	for _, field := range []StatisticField{
		StatisticCompletedEvents,
		StatisticPlannedMinutes,
		StatisticCreatedCompletedEvents,
		StatisticUniquePeopleCount,
		StatisticMostFrequentTypeID,
		StatisticMostFrequentGenreID,
		StatisticMostFrequentWeekday,
	} {
		if _, ok := visible[field]; ok {
			return true
		}
	}
	return false
}

func (s *ProfileService) statisticTiles(ctx context.Context, view statistics.View, keys []TileKey) ([]StatisticTile, error) {
	if len(keys) == 0 {
		return []StatisticTile{}, nil
	}
	typeNames := make([]string, 0, len(eventTypeNameByTile))
	for _, key := range keys {
		if name, ok := eventTypeNameByTile[key]; ok {
			typeNames = append(typeNames, name)
		}
	}
	typeIDByTile := make(map[TileKey]uint, len(typeNames))
	if len(typeNames) > 0 {
		// Типы плиток разрешаются одним узким запросом только для реально видимых ключей.
		items, err := s.store.FindEventTypesByNames(ctx, typeNames)
		if err != nil {
			return nil, err
		}
		itemByName := make(map[string]references.ReferenceItem, len(items))
		for _, item := range items {
			itemByName[item.Name] = item
		}
		for _, key := range keys {
			name, ok := eventTypeNameByTile[key]
			if !ok {
				continue
			}
			item, exists := itemByName[name]
			if !exists || item.ID == 0 {
				return nil, ErrTileReferenceMissing
			}
			typeIDByTile[key] = item.ID
		}
	}
	typeCounts := make(map[uint]int64, len(view.Types))
	for _, item := range view.Types {
		typeCounts[item.ID] = item.Count
	}
	result := make([]StatisticTile, 0, len(keys))
	for _, key := range keys {
		tile := StatisticTile{Key: key}
		switch key {
		case TileFilms, TileGames, TileTableGames, TileOther:
			value := typeCounts[typeIDByTile[key]]
			tile.Value = &value
		case TileCompletedEvents:
			value := view.CompletedEvents
			tile.Value = &value
		case TilePlannedMinutes:
			value := view.PlannedMinutes
			tile.Value = &value
			tile.Unit = "minutes"
		case TileCreatedCompletedEvents:
			value := view.CreatedCompletedEvents
			tile.Value = &value
		case TileUniquePeopleCount:
			tile.Value = view.UniquePeopleCount
			status := view.UniquePeopleStatus
			tile.Status = &status
			tile.CalculatedAt = &NullableTime{Value: view.UniquePeopleCalculatedAt}
		case TileMostFrequentTypeID:
			tile.Value = numericPointerToInt64(view.MostFrequentTypeID)
		case TileMostFrequentGenreID:
			tile.Value = numericPointerToInt64(view.MostFrequentGenreID)
		case TileMostFrequentWeekday:
			tile.Value = numericPointerToInt64(view.MostFrequentWeekday)
		}
		result = append(result, tile)
	}
	return result, nil
}

func numericPointerToInt64[T ~uint | ~int](value *T) *int64 {
	if value == nil {
		return nil
	}
	converted := int64(*value)
	return &converted
}
