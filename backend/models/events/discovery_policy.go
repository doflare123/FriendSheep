package events

import "time"

// AllowedAgeLimitNames возвращает допустимые маркировки для календарного возраста на текущий день UTC.
func AllowedAgeLimitNames(birthday *time.Time, now time.Time) []string {
	maximum := 12
	if birthday != nil {
		today := now.UTC()
		age := today.Year() - birthday.Year()
		if today.Month() < birthday.Month() || (today.Month() == birthday.Month() && today.Day() < birthday.Day()) {
			age--
		}
		switch {
		case age < 12:
			maximum = 6
		case age < 16:
			maximum = 12
		case age < 18:
			maximum = 16
		default:
			maximum = 18
		}
	}
	result := []string{"0+", "6+"}
	if maximum >= 12 {
		result = append(result, "12+")
	}
	if maximum >= 16 {
		result = append(result, "16+")
	}
	if maximum >= 18 {
		result = append(result, "18+")
	}
	return result
}
