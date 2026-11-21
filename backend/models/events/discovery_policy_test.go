package events

import (
	"reflect"
	"testing"
	"time"
)

func TestAllowedAgeLimitNamesCalendarBoundaries(t *testing.T) {
	now := time.Date(2033, 6, 15, 0, 0, 0, 0, time.UTC)
	date := func(year int, month time.Month, day int) *time.Time {
		value := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		return &value
	}
	tests := []struct {
		name     string
		birthday *time.Time
		now      time.Time
		maximum  string
	}{
		{name: "без даты рождения", now: now, maximum: "12+"},
		{name: "младше шести по продуктовой политике", birthday: date(2031, 6, 15), now: now, maximum: "6+"},
		{name: "накануне двенадцатилетия", birthday: date(2021, 6, 16), now: now, maximum: "6+"},
		{name: "двенадцатилетие", birthday: date(2021, 6, 15), now: now, maximum: "12+"},
		{name: "накануне шестнадцатилетия", birthday: date(2017, 6, 16), now: now, maximum: "12+"},
		{name: "шестнадцатилетие", birthday: date(2017, 6, 15), now: now, maximum: "16+"},
		{name: "накануне восемнадцатилетия", birthday: date(2015, 6, 16), now: now, maximum: "16+"},
		{name: "восемнадцатилетие", birthday: date(2015, 6, 15), now: now, maximum: "18+"},
		{name: "29 февраля до годовщины", birthday: date(2016, 2, 29), now: time.Date(2034, 2, 28, 23, 59, 59, 0, time.UTC), maximum: "16+"},
		{name: "29 февраля после годовщины", birthday: date(2016, 2, 29), now: time.Date(2034, 3, 1, 0, 0, 0, 0, time.UTC), maximum: "18+"},
		{name: "граница суток определяется по UTC", birthday: date(2015, 6, 15), now: time.Date(2033, 6, 15, 0, 30, 0, 0, time.FixedZone("UTC+14", 14*60*60)), maximum: "16+"},
	}
	all := []string{"0+", "6+", "12+", "16+", "18+"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var want []string
			for _, label := range all {
				want = append(want, label)
				if label == tt.maximum {
					break
				}
			}
			if got := AllowedAgeLimitNames(tt.birthday, tt.now); !reflect.DeepEqual(got, want) {
				t.Fatalf("допустимые маркировки = %v, ожидались %v", got, want)
			}
		})
	}
}
