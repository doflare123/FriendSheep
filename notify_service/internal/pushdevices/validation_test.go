package pushdevices

import (
	"math"
	"strings"
	"testing"
)

func TestRegisterValidatesV1FieldBoundariesAndOwnership(t *testing.T) {
	t.Parallel()
	registry := Registry{allowedScopes: map[ProviderScope]struct{}{{Project: "project", Environment: "test"}: {}}}
	valid := validRegisterInput(
		"ni5BJjBEB8flsp3sh_95VIcYA52LdZHWeupCksV0y-I",
		"AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8",
		math.MaxInt64, testSessionID(1), testOperationID(1), nil, strings.Repeat("z", 4096), "project", "test",
	)
	if err := registry.validateRegister(valid); err != nil {
		t.Fatalf("запрос на верхней границе полей: %v", err)
	}
	tests := []struct {
		name   string
		change func(*RegisterInput)
		code   string
	}{
		{"нет согласия", func(input *RegisterInput) { input.NotificationsEnabled = false }, ErrorCodePushOptInRequired},
		{"неподдерживаемая версия", func(input *RegisterInput) { input.SchemaVersion++ }, ErrorCodeUnsupportedSchemaVersion},
		{"неразрешённое окружение", func(input *RegisterInput) { input.Environment = "prod" }, ErrorCodeInvalidProviderScope},
		{"адрес больше лимита", func(input *RegisterInput) { input.ProviderToken += "x" }, ErrorCodeInvalidRequest},
		{"адрес с пробелом", func(input *RegisterInput) { input.ProviderToken = "a b" }, ErrorCodeInvalidRequest},
		{"userId выше BIGINT", func(input *RegisterInput) { input.UserID++ }, ErrorCodeInvalidRequest},
		{"неканонический ID", func(input *RegisterInput) { input.InstallationID = input.InstallationID[:42] + "B" }, ErrorCodeInvalidRequest},
		{"известный ID без секрета", func(input *RegisterInput) { input.InstallationSecret = "" }, ErrorCodePushDeviceNotFound},
		{"секрет другой установки", func(input *RegisterInput) { input.InstallationSecret = strings.Repeat("A", 43) }, ErrorCodePushDeviceNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := valid
			test.change(&input)
			assertPushDomainError(t, registry.validateRegister(input), test.code)
		})
	}
}
