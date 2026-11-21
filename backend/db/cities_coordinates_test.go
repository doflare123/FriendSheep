package db

import "testing"

func TestParseCityCoordinates(t *testing.T) {
	latitude, longitude, err := parseCityCoordinates("54.710426", "20.452214")
	if err != nil {
		t.Fatalf("parseCityCoordinates returned error: %v", err)
	}
	if latitude == nil || longitude == nil || *latitude != 54.710426 || *longitude != 20.452214 {
		t.Fatalf("coordinates = %v/%v", latitude, longitude)
	}

	latitude, longitude, err = parseCityCoordinates("", "")
	if err != nil || latitude != nil || longitude != nil {
		t.Fatalf("empty coordinates = %v/%v, err %v", latitude, longitude, err)
	}
}

func TestParseCityCoordinatesRejectsInvalidPair(t *testing.T) {
	tests := []struct {
		latitude  string
		longitude string
	}{
		{latitude: "54.7"},
		{latitude: "91", longitude: "20"},
		{latitude: "54", longitude: "181"},
		{latitude: "NaN", longitude: "20"},
	}

	for _, test := range tests {
		if _, _, err := parseCityCoordinates(test.latitude, test.longitude); err == nil {
			t.Fatalf("parseCityCoordinates(%q, %q) returned nil error", test.latitude, test.longitude)
		}
	}
}
