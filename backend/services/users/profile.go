package users

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"friendship/services/statistics"
)

var (
	ErrUnauthenticated = errors.New("пользователь не авторизован")
	ErrInvalidTarget   = errors.New("некорректный идентификатор профиля")
	ErrUserNotFound    = errors.New("пользователь не найден")
	ErrProfileClosed   = errors.New("профиль закрыт")
	ErrInvalidPrivacy  = errors.New("некорректные настройки приватности")
)

type TargetKind uint8

const (
	TargetSelf TargetKind = iota + 1
	TargetByID
	TargetByUs
)

type Target struct {
	Kind TargetKind
	ID   uint
	Us   string
}

func OwnTarget() Target       { return Target{Kind: TargetSelf} }
func IDTarget(id uint) Target { return Target{Kind: TargetByID, ID: id} }
func UsTarget(us string) Target {
	return Target{Kind: TargetByUs, Us: us}
}

// ParseIdentifier однозначно разделяет числовой ID и адрес с префиксом @.
func ParseIdentifier(raw string) (Target, error) {
	if strings.HasPrefix(raw, "@") {
		if len(raw) == 1 {
			return Target{}, ErrInvalidTarget
		}
		return UsTarget(raw[1:]), nil
	}
	if raw == "" {
		return Target{}, ErrInvalidTarget
	}
	for i := range raw {
		if raw[i] < '0' || raw[i] > '9' {
			return Target{}, ErrInvalidTarget
		}
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 || uint64(uint(id)) != id {
		return Target{}, ErrInvalidTarget
	}
	return IDTarget(uint(id)), nil
}

type PrivacyMode string

const (
	PrivacyModeOpen   PrivacyMode = "open"
	PrivacyModeClosed PrivacyMode = "closed"
	PrivacyModeCustom PrivacyMode = "custom"
)

type StatisticField string

const (
	StatisticCompletedEvents        StatisticField = "completedEvents"
	StatisticPlannedMinutes         StatisticField = "plannedMinutes"
	StatisticCreatedCompletedEvents StatisticField = "createdCompletedEvents"
	StatisticUniquePeopleCount      StatisticField = "uniquePeopleCount"
	StatisticMostFrequentTypeID     StatisticField = "mostFrequentTypeId"
	StatisticMostFrequentGenreID    StatisticField = "mostFrequentGenreId"
	StatisticMostFrequentWeekday    StatisticField = "mostFrequentWeekday"
	StatisticFilms                  StatisticField = "films"
	StatisticGames                  StatisticField = "games"
	StatisticTableGames             StatisticField = "tableGames"
	StatisticOther                  StatisticField = "other"
)

var allStatisticFields = []StatisticField{
	StatisticFilms,
	StatisticGames,
	StatisticTableGames,
	StatisticOther,
	StatisticCompletedEvents,
	StatisticPlannedMinutes,
	StatisticCreatedCompletedEvents,
	StatisticUniquePeopleCount,
	StatisticMostFrequentTypeID,
	StatisticMostFrequentGenreID,
	StatisticMostFrequentWeekday,
}

func AllStatisticFields() []StatisticField {
	return append([]StatisticField(nil), allStatisticFields...)
}

type Privacy struct {
	Mode              PrivacyMode      `json:"mode" enums:"open,closed,custom"`
	VisibleStatistics []StatisticField `json:"visibleStatistics"`
}

func DefaultPrivacy() Privacy {
	return Privacy{Mode: PrivacyModeOpen, VisibleStatistics: AllStatisticFields()}
}

type PrivacyPatch struct {
	Mode              *PrivacyMode      `json:"mode,omitempty" enums:"open,closed,custom"`
	VisibleStatistics *[]StatisticField `json:"visibleStatistics,omitempty"`
}

type UserRecord struct {
	ID           uint
	Name         string
	Us           string
	Image        string
	Status       string
	CreatedAt    time.Time
	VerifiedUser bool
	Privacy      Privacy
}

type ProfileStore interface {
	TileSettingsStore
	FindUser(context.Context, Target) (UserRecord, error)
	GetPrivacy(context.Context, uint) (Privacy, error)
	PatchPrivacy(context.Context, uint, PrivacyPatch) (Privacy, error)
}

type StatisticsReader interface {
	Read(context.Context, uint) (statistics.View, error)
}

type NullableInt64 struct{ Value *int64 }
type NullableUint struct{ Value *uint }
type NullableInt struct{ Value *int }
type NullableTime struct{ Value *time.Time }

func (v NullableInt64) MarshalJSON() ([]byte, error) { return json.Marshal(v.Value) }
func (v NullableUint) MarshalJSON() ([]byte, error)  { return json.Marshal(v.Value) }
func (v NullableInt) MarshalJSON() ([]byte, error)   { return json.Marshal(v.Value) }
func (v NullableTime) MarshalJSON() ([]byte, error)  { return json.Marshal(v.Value) }

type CompactStatistics struct {
	CompletedEvents          *int64         `json:"completedEvents,omitempty"`
	PlannedMinutes           *int64         `json:"plannedMinutes,omitempty"`
	CreatedCompletedEvents   *int64         `json:"createdCompletedEvents,omitempty"`
	UniquePeopleCount        *NullableInt64 `json:"uniquePeopleCount,omitempty" swaggertype:"integer" extensions:"x-nullable"`
	UniquePeopleStatus       *string        `json:"uniquePeopleStatus,omitempty"`
	UniquePeopleCalculatedAt *NullableTime  `json:"uniquePeopleCalculatedAt,omitempty" swaggertype:"string" format:"date-time" extensions:"x-nullable"`
	MostFrequentTypeID       *NullableUint  `json:"mostFrequentTypeId,omitempty" swaggertype:"integer" extensions:"x-nullable"`
	MostFrequentGenreID      *NullableUint  `json:"mostFrequentGenreId,omitempty" swaggertype:"integer" extensions:"x-nullable"`
	MostFrequentWeekday      *NullableInt   `json:"mostFrequentWeekday,omitempty" swaggertype:"integer" extensions:"x-nullable"`
}

type Profile struct {
	ID           uint               `json:"id"`
	Name         string             `json:"name"`
	Us           string             `json:"us"`
	Image        string             `json:"image"`
	Status       string             `json:"status"`
	DataRegister time.Time          `json:"dataRegister"`
	Verified     bool               `json:"verified"`
	IsOwner      bool               `json:"isOwner"`
	Statistics   *CompactStatistics `json:"statistics,omitempty"`
	Tiles        []StatisticTile    `json:"tiles"`
}

type ProfileService struct {
	store      ProfileStore
	statistics StatisticsReader
}

func NewProfileService(store ProfileStore, statisticsReader StatisticsReader) *ProfileService {
	return &ProfileService{store: store, statistics: statisticsReader}
}

func (s *ProfileService) GetProfile(ctx context.Context, viewerID uint, target Target) (Profile, error) {
	if viewerID == 0 {
		return Profile{}, ErrUnauthenticated
	}
	resolvedTarget, err := resolveTarget(viewerID, target)
	if err != nil {
		return Profile{}, err
	}
	user, err := s.store.FindUser(ctx, resolvedTarget)
	if err != nil {
		return Profile{}, err
	}
	isOwner := user.ID == viewerID
	if !isOwner && user.Privacy.Mode == PrivacyModeClosed {
		return Profile{}, ErrProfileClosed
	}
	visible := AllStatisticFields()
	if !isOwner && user.Privacy.Mode == PrivacyModeCustom {
		visible = user.Privacy.VisibleStatistics
	}
	profile := Profile{
		ID:           user.ID,
		Name:         user.Name,
		Us:           user.Us,
		Image:        user.Image,
		Status:       user.Status,
		DataRegister: user.CreatedAt,
		Verified:     user.VerifiedUser,
		IsOwner:      isOwner,
		Statistics:   &CompactStatistics{},
		Tiles:        []StatisticTile{},
	}
	if len(visible) == 0 {
		return profile, nil
	}
	// Один набор видимости управляет и компактной статистикой, и плитками профиля.
	settings, err := s.store.GetTileSettings(ctx, user.ID)
	if err != nil {
		return Profile{}, err
	}
	visibleSet := make(map[StatisticField]struct{}, len(visible))
	for _, field := range visible {
		visibleSet[field] = struct{}{}
	}
	selectedTiles := visibleSelectedTiles(settings, visibleSet)
	if !compactStatisticsRequested(visibleSet) && len(selectedTiles) == 0 {
		return profile, nil
	}
	if s.statistics == nil {
		return Profile{}, errors.New("сервис статистики профиля не настроен")
	}
	view, err := s.statistics.Read(ctx, user.ID)
	if err != nil {
		return Profile{}, err
	}
	// Компактные поля и плитки строятся из одного сохранённого среза статистики.
	profile.Statistics = compactStatistics(view, visible)
	profile.Tiles, err = s.statisticTiles(ctx, view, selectedTiles)
	if err != nil {
		return Profile{}, err
	}
	return profile, nil
}

func (s *ProfileService) GetPrivacy(ctx context.Context, userID uint) (Privacy, error) {
	if userID == 0 {
		return Privacy{}, ErrUnauthenticated
	}
	return s.store.GetPrivacy(ctx, userID)
}

func (s *ProfileService) PatchPrivacy(ctx context.Context, userID uint, patch PrivacyPatch) (Privacy, error) {
	if userID == 0 {
		return Privacy{}, ErrUnauthenticated
	}
	normalized, err := normalizePatch(patch)
	if err != nil {
		return Privacy{}, err
	}
	return s.store.PatchPrivacy(ctx, userID, normalized)
}

func resolveTarget(viewerID uint, target Target) (Target, error) {
	switch target.Kind {
	case TargetSelf:
		if target.ID != 0 || target.Us != "" {
			return Target{}, ErrInvalidTarget
		}
		return IDTarget(viewerID), nil
	case TargetByID:
		if target.ID == 0 || target.Us != "" {
			return Target{}, ErrInvalidTarget
		}
		return target, nil
	case TargetByUs:
		if target.ID != 0 || target.Us == "" {
			return Target{}, ErrInvalidTarget
		}
		return target, nil
	default:
		return Target{}, ErrInvalidTarget
	}
}

func normalizePatch(patch PrivacyPatch) (PrivacyPatch, error) {
	if patch.Mode == nil && patch.VisibleStatistics == nil {
		return PrivacyPatch{}, ErrInvalidPrivacy
	}
	if patch.Mode != nil && !validMode(*patch.Mode) {
		return PrivacyPatch{}, ErrInvalidPrivacy
	}
	if patch.VisibleStatistics == nil {
		return patch, nil
	}
	requested := make(map[StatisticField]struct{}, len(*patch.VisibleStatistics))
	for _, field := range *patch.VisibleStatistics {
		if !validStatisticField(field) {
			return PrivacyPatch{}, ErrInvalidPrivacy
		}
		if _, duplicate := requested[field]; duplicate {
			return PrivacyPatch{}, ErrInvalidPrivacy
		}
		requested[field] = struct{}{}
	}
	ordered := make([]StatisticField, 0, len(requested))
	for _, field := range allStatisticFields {
		if _, ok := requested[field]; ok {
			ordered = append(ordered, field)
		}
	}
	patch.VisibleStatistics = &ordered
	return patch, nil
}

func validMode(mode PrivacyMode) bool {
	return mode == PrivacyModeOpen || mode == PrivacyModeClosed || mode == PrivacyModeCustom
}

func validStatisticField(field StatisticField) bool {
	for _, candidate := range allStatisticFields {
		if candidate == field {
			return true
		}
	}
	return false
}

func compactStatistics(view statistics.View, visible []StatisticField) *CompactStatistics {
	result := &CompactStatistics{}
	for _, field := range visible {
		switch field {
		case StatisticCompletedEvents:
			value := view.CompletedEvents
			result.CompletedEvents = &value
		case StatisticPlannedMinutes:
			value := view.PlannedMinutes
			result.PlannedMinutes = &value
		case StatisticCreatedCompletedEvents:
			value := view.CreatedCompletedEvents
			result.CreatedCompletedEvents = &value
		case StatisticUniquePeopleCount:
			status := view.UniquePeopleStatus
			result.UniquePeopleCount = &NullableInt64{Value: view.UniquePeopleCount}
			result.UniquePeopleStatus = &status
			result.UniquePeopleCalculatedAt = &NullableTime{Value: view.UniquePeopleCalculatedAt}
		case StatisticMostFrequentTypeID:
			result.MostFrequentTypeID = &NullableUint{Value: view.MostFrequentTypeID}
		case StatisticMostFrequentGenreID:
			result.MostFrequentGenreID = &NullableUint{Value: view.MostFrequentGenreID}
		case StatisticMostFrequentWeekday:
			result.MostFrequentWeekday = &NullableInt{Value: view.MostFrequentWeekday}
		}
	}
	return result
}
