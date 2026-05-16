package tests

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

type subImgServiceStub struct {
	uploadURL string
	uploadErr error
	called    bool
	folder    string
}

func (s *subImgServiceStub) UploadImg(_ context.Context, _ *multipart.FileHeader, folder string) (string, error) {
	s.called = true
	s.folder = folder
	if s.uploadErr != nil {
		return "", s.uploadErr
	}
	return s.uploadURL, nil
}

func TestSubHandlerChangePhotoRejectsMissingImageWithCommonErrorShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	imgService := &subImgServiceStub{}
	handler := handlers.NewSubHandler(imgService)

	rec := performSubPhotoRequest(t, handler.ChangePhoto, false)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	assertCommonErrorShape(t, rec)
	if imgService.called {
		t.Fatal("UploadImg should not be called when image form file is missing")
	}
}

func TestSubHandlerChangePhotoUploadErrorUsesCommonErrorShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	imgService := &subImgServiceStub{uploadErr: errors.New("upload failed")}
	handler := handlers.NewSubHandler(imgService)

	rec := performSubPhotoRequest(t, handler.ChangePhoto, true)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	payload := assertCommonErrorShape(t, rec)
	if payload["error"] != "upload failed" {
		t.Fatalf("error = %#v, want %q", payload["error"], "upload failed")
	}
	if !imgService.called {
		t.Fatal("UploadImg was not called")
	}
	if imgService.folder != "photos" {
		t.Fatalf("folder = %q, want %q", imgService.folder, "photos")
	}
}

func TestSubHandlerChangePhotoSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	imgService := &subImgServiceStub{uploadURL: "https://cdn.example.com/photo.png"}
	handler := handlers.NewSubHandler(imgService)

	rec := performSubPhotoRequest(t, handler.ChangePhoto, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"image":"https://cdn.example.com/photo.png"`) {
		t.Fatalf("response does not contain uploaded image URL: %s", rec.Body.String())
	}
}

func performSubPhotoRequest(t *testing.T, handler gin.HandlerFunc, withImage bool) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	if withImage {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("image", "photo.png")
		if err != nil {
			t.Fatalf("CreateFormFile returned error: %v", err)
		}
		if _, err := part.Write([]byte("fake-image-bytes")); err != nil {
			t.Fatalf("Write returned error: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("writer.Close returned error: %v", err)
		}

		ctx.Request = httptest.NewRequest(http.MethodPost, "/", &body)
		ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		ctx.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	}

	handler(ctx)
	return rec
}
