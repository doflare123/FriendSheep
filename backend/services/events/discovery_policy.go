package events

import (
	"math"
	"sort"
	"time"

	eventmodels "friendship/models/events"
)

// Дата рождения хранится как календарная дата; текущий день определяется в UTC.
func allowedEventAgeLimits(birthday *time.Time, now time.Time) []string {
	return eventmodels.AllowedAgeLimitNames(birthday, now)
}

func validDiscoveryCoordinates(latitude, longitude *float64) bool {
	return latitude != nil && longitude != nil && !math.IsNaN(*latitude) && !math.IsNaN(*longitude) &&
		!math.IsInf(*latitude, 0) && !math.IsInf(*longitude, 0) &&
		*latitude >= -90 && *latitude <= 90 && *longitude >= -180 && *longitude <= 180
}

// Расстояние между центрами городов задаёт порядок, но не заменяет расстояние до мероприятия.
func rankEventCities(preferredID uint, cities []EventDiscoveryCity) []uint {
	var origin *EventDiscoveryCity
	for i := range cities {
		if cities[i].ID == preferredID {
			origin = &cities[i]
			break
		}
	}
	if origin == nil {
		return nil
	}
	result := []uint{preferredID}
	if !validDiscoveryCoordinates(origin.Latitude, origin.Longitude) {
		return result
	}
	type cityDistance struct {
		id       uint
		distance float64
	}
	nearby := make([]cityDistance, 0, len(cities))
	for _, city := range cities {
		if city.ID == preferredID || !validDiscoveryCoordinates(city.Latitude, city.Longitude) {
			continue
		}
		const radians = math.Pi / 180
		lat1, lat2 := *origin.Latitude*radians, *city.Latitude*radians
		dlat, dlon := lat2-lat1, (*city.Longitude-*origin.Longitude)*radians
		a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
		nearby = append(nearby, cityDistance{id: city.ID, distance: a})
	}
	sort.Slice(nearby, func(i, j int) bool {
		if nearby[i].distance == nearby[j].distance {
			return nearby[i].id < nearby[j].id
		}
		return nearby[i].distance < nearby[j].distance
	})
	for _, city := range nearby {
		result = append(result, city.id)
	}
	return result
}
