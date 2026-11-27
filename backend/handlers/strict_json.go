package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"unicode/utf8"
)

var (
	ErrStrictJSONInvalid         = errors.New("некорректный JSON-объект")
	ErrStrictJSONRequestTooLarge = errors.New("тело запроса превышает лимит")
)

// DecodeStrictJSONObject читает ограниченный JSON-объект без неизвестных, повторных и пропущенных полей.
func DecodeStrictJSONObject(w http.ResponseWriter, r *http.Request, maxBytes int64, destination any, requiredFields ...string) error {
	if r == nil || r.Body == nil || maxBytes <= 0 || destination == nil {
		return ErrStrictJSONInvalid
	}
	body := http.MaxBytesReader(w, r.Body, maxBytes)
	defer body.Close()
	raw, err := io.ReadAll(body)
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return ErrStrictJSONRequestTooLarge
	}
	if err != nil {
		return ErrStrictJSONInvalid
	}
	if !utf8.Valid(raw) {
		return ErrStrictJSONInvalid
	}

	nullable := nullableJSONFields(destination)
	seen, err := inspectStrictJSONObject(raw, nullable)
	if err != nil {
		return ErrStrictJSONInvalid
	}
	for _, field := range requiredFields {
		if _, exists := seen[field]; !exists {
			return ErrStrictJSONInvalid
		}
	}
	if len(seen) != len(requiredFields) {
		return ErrStrictJSONInvalid
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return ErrStrictJSONInvalid
	}
	if _, err := decoder.Token(); err != io.EOF {
		return ErrStrictJSONInvalid
	}
	return nil
}

func inspectStrictJSONObject(raw []byte, nullable map[string]struct{}) (map[string]struct{}, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, ErrStrictJSONInvalid
	}
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return nil, ErrStrictJSONInvalid
		}
		if _, duplicate := seen[key]; duplicate {
			return nil, ErrStrictJSONInvalid
		}
		seen[key] = struct{}{}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, ErrStrictJSONInvalid
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			if _, allowed := nullable[key]; !allowed {
				return nil, ErrStrictJSONInvalid
			}
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return nil, ErrStrictJSONInvalid
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, ErrStrictJSONInvalid
	}
	return seen, nil
}

func nullableJSONFields(destination any) map[string]struct{} {
	result := make(map[string]struct{})
	typeOf := reflect.TypeOf(destination)
	if typeOf.Kind() == reflect.Pointer {
		typeOf = typeOf.Elem()
	}
	if typeOf.Kind() != reflect.Struct {
		return result
	}
	for index := 0; index < typeOf.NumField(); index++ {
		field := typeOf.Field(index)
		if field.Type.Kind() != reflect.Pointer {
			continue
		}
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name != "" && name != "-" {
			result[name] = struct{}{}
		}
	}
	return result
}
