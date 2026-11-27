package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"unicode/utf8"

	"notify_service/internal/pushdevices"
)

const pushBodyLimit = 16_384

type PushDevices interface {
	Register(context.Context, pushdevices.RegisterInput) (pushdevices.MutationResult, error)
	Detach(context.Context, pushdevices.DetachInput) (pushdevices.MutationResult, error)
}

func registerPushDeviceRoutes(mux *http.ServeMux, internalToken string, registry PushDevices) {
	secure := func(handler http.HandlerFunc) http.Handler {
		protected := requireInternalToken(internalToken, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if registry == nil {
				writeError(w, http.StatusServiceUnavailable, "push_registry_unavailable")
				return
			}
			handler(w, r)
		}))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			protected.ServeHTTP(w, r)
		})
	}

	mux.Handle("PUT /internal/v1/push-devices/{installationId}", secure(func(w http.ResponseWriter, r *http.Request) {
		input := pushdevices.RegisterInput{InstallationID: r.PathValue("installationId")}
		secret, ok := readInstallationSecret(w, r)
		if !ok {
			return
		}
		input.InstallationSecret = secret
		if !readPushBody(w, r, map[string]any{
			"schemaVersion": &input.SchemaVersion, "userId": &input.UserID, "sessionId": &input.SessionID,
			"operationId": &input.OperationID, "expectedVersion": &input.ExpectedVersion,
			"platform": &input.Platform, "provider": &input.Provider, "providerProject": &input.ProviderProject,
			"environment": &input.Environment, "providerToken": &input.ProviderToken, "notificationsEnabled": &input.NotificationsEnabled,
		}) {
			return
		}
		result, err := registry.Register(r.Context(), input)
		writePushResult(w, result, err)
	}))

	mux.Handle("POST /internal/v1/push-devices/{installationId}/detach", secure(func(w http.ResponseWriter, r *http.Request) {
		input := pushdevices.DetachInput{InstallationID: r.PathValue("installationId")}
		secret, ok := readInstallationSecret(w, r)
		if !ok {
			return
		}
		input.InstallationSecret = secret
		if !readPushBody(w, r, map[string]any{
			"schemaVersion": &input.SchemaVersion, "userId": &input.UserID, "sessionId": &input.SessionID,
			"operationId": &input.OperationID, "expectedVersion": &input.ExpectedVersion,
		}) {
			return
		}
		result, err := registry.Detach(r.Context(), input)
		writePushResult(w, result, err)
	}))
}

func readInstallationSecret(w http.ResponseWriter, r *http.Request) (string, bool) {
	values := r.Header.Values("X-Installation-Secret")
	if len(values) == 0 || len(values) == 1 && values[0] == "" {
		writeError(w, http.StatusNotFound, "push_device_not_found")
		return "", false
	}
	if len(values) != 1 {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return "", false
	}
	return values[0], true
}

func readPushBody(w http.ResponseWriter, r *http.Request, fields map[string]any) bool {
	if r.URL.RawQuery != "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return false
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" || len(r.Header.Values("Content-Type")) != 1 {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type")
		return false
	}

	body := http.MaxBytesReader(w, r.Body, pushBodyLimit)
	defer body.Close()
	raw, err := io.ReadAll(body)
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "request_too_large")
		return false
	}
	if err != nil || !decodePushFields(raw, fields) {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return false
	}
	return true
}

func decodePushFields(data []byte, fields map[string]any) bool {
	if !utf8.Valid(data) {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return false
	}
	for decoder.More() {
		token, err := decoder.Token()
		key, _ := token.(string)
		destination, exists := fields[key]
		if err != nil || !exists {
			return false
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil || key != "expectedVersion" && bytes.Equal(value, []byte("null")) {
			return false
		}
		if err := json.Unmarshal(value, destination); err != nil {
			return false
		}
		delete(fields, key)
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || len(fields) != 0 {
		return false
	}
	_, err = decoder.Token()
	return err == io.EOF
}

func writePushResult(w http.ResponseWriter, result pushdevices.MutationResult, err error) {
	if err == nil {
		writeAnyJSON(w, http.StatusOK, struct {
			SchemaVersion  int    `json:"schemaVersion"`
			InstallationID string `json:"installationId"`
			Version        string `json:"version"`
			State          string `json:"state"`
		}{
			SchemaVersion:  result.SchemaVersion,
			InstallationID: result.InstallationID,
			Version:        result.Version,
			State:          result.State,
		})
		return
	}

	var domainErr *pushdevices.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case pushdevices.ErrorCodeInvalidRequest, pushdevices.ErrorCodeUnsupportedSchemaVersion, pushdevices.ErrorCodePushOptInRequired, pushdevices.ErrorCodeInvalidProviderScope:
			writeError(w, http.StatusBadRequest, domainErr.Code)
			return
		case pushdevices.ErrorCodePushDeviceNotFound:
			writeError(w, http.StatusNotFound, domainErr.Code)
			return
		case pushdevices.ErrorCodePushVersionConflict:
			writeAnyJSON(w, http.StatusConflict, map[string]any{
				"error": domainErr.Code, "details": map[string]*string{"currentVersion": domainErr.CurrentVersion},
			})
			return
		case pushdevices.ErrorCodePushOperationConflict, pushdevices.ErrorCodePushInstallationConflict,
			pushdevices.ErrorCodePushTokenConflict, pushdevices.ErrorCodePushTokenInvalidated:
			writeError(w, http.StatusConflict, domainErr.Code)
			return
		}
	}
	writeError(w, http.StatusServiceUnavailable, "push_registry_unavailable")
}
