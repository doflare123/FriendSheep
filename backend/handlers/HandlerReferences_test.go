package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"friendship/models/dto"
	"friendship/services/references"

	"github.com/gin-gonic/gin"
)

type referencesHandlerContextKey struct{}

type referencesHandlerServiceStub struct {
	referencesResult *dto.ReferencesDto
	referencesErr    error
	searchResult     *dto.GenreSearchResponseDto
	searchErr        error
	cityResult       *dto.CitySearchResponseDto
	cityErr          error
	referencesCalls  int
	searchCalls      int
	cityCalls        int
	referencesCtx    context.Context
	searchCtx        context.Context
	searchInput      references.GenreSearchInput
	cityCtx          context.Context
	cityInput        references.CitySearchInput
}

func (s *referencesHandlerServiceStub) GetReferences(ctx context.Context) (*dto.ReferencesDto, error) {
	s.referencesCalls++
	s.referencesCtx = ctx
	return s.referencesResult, s.referencesErr
}

func (s *referencesHandlerServiceStub) SearchGenres(ctx context.Context, input references.GenreSearchInput) (*dto.GenreSearchResponseDto, error) {
	s.searchCalls++
	s.searchCtx = ctx
	s.searchInput = input
	return s.searchResult, s.searchErr
}

func (s *referencesHandlerServiceStub) SearchCities(ctx context.Context, input references.CitySearchInput) (*dto.CitySearchResponseDto, error) {
	s.cityCalls++
	s.cityCtx = ctx
	s.cityInput = input
	return s.cityResult, s.cityErr
}

func TestReferencesHandlerGetReferencesPassesContextAndOmitsGenres(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &referencesHandlerServiceStub{
		referencesResult: &dto.ReferencesDto{
			EventTypes: []dto.ReferenceItemDto{{ID: 7, Name: "Game"}},
		},
	}
	handler := NewReferencesHandler(stub)
	recorder, ctx := newReferencesHandlerContext(t, "/api/v2/references")
	requestCtx := context.WithValue(ctx.Request.Context(), referencesHandlerContextKey{}, "references")
	ctx.Request = ctx.Request.WithContext(requestCtx)

	handler.GetReferences(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if stub.referencesCalls != 1 {
		t.Fatalf("GetReferences calls = %d, want 1", stub.referencesCalls)
	}
	if stub.referencesCtx.Value(referencesHandlerContextKey{}) != "references" {
		t.Fatal("request context was not passed to ReferenceService")
	}
	var response map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, exists := response["genres"]; exists {
		t.Fatalf("general references response still contains genres: %s", recorder.Body.String())
	}
	if _, exists := response["cities"]; exists {
		t.Fatalf("general references response still contains cities: %s", recorder.Body.String())
	}
	if _, exists := response["eventTypes"]; !exists {
		t.Fatalf("general references response is missing eventTypes: %s", recorder.Body.String())
	}
}

func TestReferencesHandlerGetReferencesMapsServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &referencesHandlerServiceStub{referencesErr: errors.New("reference storage failed")}
	handler := NewReferencesHandler(stub)
	recorder, ctx := newReferencesHandlerContext(t, "/api/v2/references")

	handler.GetReferences(ctx)

	assertReferencesHandlerError(t, recorder, http.StatusInternalServerError, "reference storage failed")
}

func TestReferencesHandlerSearchGenresPassesNormalizedQueryAndPagination(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wantInput references.GenreSearchInput
	}{
		{
			name:   "defaults",
			target: "/api/v2/references/genres",
			wantInput: references.GenreSearchInput{
				Page:  references.DefaultGenrePage,
				Limit: references.DefaultGenreLimit,
			},
		},
		{
			name:   "explicit values",
			target: "/api/v2/references/genres?q=%20Board%20&page=2&limit=10",
			wantInput: references.GenreSearchInput{
				Query: "Board",
				Page:  2,
				Limit: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			stub := &referencesHandlerServiceStub{
				searchResult: &dto.GenreSearchResponseDto{
					Items: []dto.ReferenceItemDto{{ID: 8, Name: "Board Games"}},
					Total: 1,
					Page:  tt.wantInput.Page,
					Limit: tt.wantInput.Limit,
				},
			}
			handler := NewReferencesHandler(stub)
			recorder, ctx := newReferencesHandlerContext(t, tt.target)
			requestCtx := context.WithValue(ctx.Request.Context(), referencesHandlerContextKey{}, "search")
			ctx.Request = ctx.Request.WithContext(requestCtx)

			handler.SearchGenres(ctx)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			if stub.searchCalls != 1 {
				t.Fatalf("SearchGenres calls = %d, want 1", stub.searchCalls)
			}
			if stub.searchCtx.Value(referencesHandlerContextKey{}) != "search" {
				t.Fatal("request context was not passed to ReferenceService")
			}
			if stub.searchInput != tt.wantInput {
				t.Fatalf("input = %#v, want %#v", stub.searchInput, tt.wantInput)
			}
			var response dto.GenreSearchResponseDto
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(response.Items) != 1 ||
				response.Items[0].ID != 8 ||
				response.Items[0].Name != "Board Games" ||
				response.Total != 1 ||
				response.Page != tt.wantInput.Page ||
				response.Limit != tt.wantInput.Limit ||
				response.HasMore {
				t.Fatalf("genre search response = %#v, want complete paginated payload", response)
			}
		})
	}
}

func TestReferencesHandlerSearchGenresRejectsInvalidPaginationWithoutServiceCall(t *testing.T) {
	tests := []string{
		"page=0",
		"page=-1",
		"page=bad",
		"limit=0",
		"limit=101",
		"limit=bad",
	}

	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			stub := &referencesHandlerServiceStub{}
			handler := NewReferencesHandler(stub)
			recorder, ctx := newReferencesHandlerContext(t, "/api/v2/references/genres?"+query)

			handler.SearchGenres(ctx)

			assertReferencesHandlerError(t, recorder, http.StatusBadRequest, "")
			if stub.searchCalls != 0 {
				t.Fatalf("SearchGenres calls = %d, want 0", stub.searchCalls)
			}
		})
	}
}

func TestReferencesHandlerSearchGenresMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{
			name:       "invalid page",
			serviceErr: references.ErrInvalidGenrePage,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid limit",
			serviceErr: references.ErrInvalidGenreLimit,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "storage error",
			serviceErr: errors.New("genre storage failed"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			stub := &referencesHandlerServiceStub{searchErr: tt.serviceErr}
			handler := NewReferencesHandler(stub)
			recorder, ctx := newReferencesHandlerContext(t, "/api/v2/references/genres?page=1&limit=10")

			handler.SearchGenres(ctx)

			assertReferencesHandlerError(t, recorder, tt.wantStatus, tt.serviceErr.Error())
		})
	}
}

func TestReferencesHandlerSearchCitiesPassesInputAndMapsResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &referencesHandlerServiceStub{cityResult: &dto.CitySearchResponseDto{
		Items: []dto.CityReferenceItemDto{{ID: 8, Name: "Советск", Region: "Калининградская обл"}},
		Total: 1, Page: 2, Limit: 10,
	}}
	handler := NewReferencesHandler(stub)
	recorder, ctx := newReferencesHandlerContext(t, "/api/v2/references/cities?q=%20%D0%A1%D0%BE%D0%B2%D0%B5%D1%82%20&page=2&limit=10")
	requestCtx := context.WithValue(ctx.Request.Context(), referencesHandlerContextKey{}, "cities")
	ctx.Request = ctx.Request.WithContext(requestCtx)

	handler.SearchCities(ctx)

	if recorder.Code != http.StatusOK || stub.cityCalls != 1 || stub.cityCtx.Value(referencesHandlerContextKey{}) != "cities" ||
		stub.cityInput != (references.CitySearchInput{Query: "Совет", Page: 2, Limit: 10}) {
		t.Fatalf("status = %d, calls = %d, input = %#v", recorder.Code, stub.cityCalls, stub.cityInput)
	}
	var response map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode city response: %v", err)
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(response["items"], &items); err != nil {
		t.Fatalf("decode city items: %v", err)
	}
	if len(items) != 1 || len(items[0]) != 3 || items[0]["id"] == nil || items[0]["name"] == nil || items[0]["region"] == nil {
		t.Fatalf("public city item = %#v, want only id, name and region", items)
	}
	var payload dto.CitySearchResponseDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode city payload: %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0].ID != 8 || payload.Items[0].Name != "Советск" || payload.Items[0].Region != "Калининградская обл" || payload.Total != 1 || payload.Page != 2 || payload.Limit != 10 || payload.HasMore {
		t.Fatalf("city payload = %#v", payload)
	}
}

func TestReferencesHandlerSearchCitiesRejectsInvalidPaginationAndMapsErrors(t *testing.T) {
	for _, query := range []string{"page=0", "page=bad", "limit=0", "limit=101", "limit=bad"} {
		t.Run(query, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			stub := &referencesHandlerServiceStub{}
			recorder, ctx := newReferencesHandlerContext(t, "/api/v2/references/cities?"+query)
			NewReferencesHandler(stub).SearchCities(ctx)
			assertReferencesHandlerError(t, recorder, http.StatusBadRequest, "")
			if stub.cityCalls != 0 {
				t.Fatalf("SearchCities calls = %d, want 0", stub.cityCalls)
			}
		})
	}

	gin.SetMode(gin.TestMode)
	stub := &referencesHandlerServiceStub{cityErr: errors.New("city storage failed")}
	recorder, ctx := newReferencesHandlerContext(t, "/api/v2/references/cities?page=1&limit=10")
	NewReferencesHandler(stub).SearchCities(ctx)
	assertReferencesHandlerError(t, recorder, http.StatusInternalServerError, "city storage failed")
}

func newReferencesHandlerContext(t *testing.T, target string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, target, nil)
	ctx.Request = request
	return recorder, ctx
}

func assertReferencesHandlerError(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
	t.Helper()

	if recorder.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, wantStatus, recorder.Body.String())
	}
	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error == "" || response.Message == "" || response.Error != response.Message {
		t.Fatalf("error response = %#v, want mirrored non-empty error and message", response)
	}
	if wantMessage != "" && response.Error != wantMessage {
		t.Fatalf("error = %q, want %q", response.Error, wantMessage)
	}
}

var _ references.ReferenceService = (*referencesHandlerServiceStub)(nil)
