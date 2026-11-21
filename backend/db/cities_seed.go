package db

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"friendship/models"
	"friendship/repository"

	"gorm.io/gorm/clause"
)

const citySeedBatchSize = 200

//go:embed data/cities_ru.csv
var citySeedFS embed.FS

func seedRussianCities(db repository.PostgresRepository) error {
	data, err := citySeedFS.ReadFile("data/cities_ru.csv")
	if err != nil {
		return fmt.Errorf("не удалось прочитать встроенный справочник: %w", err)
	}

	cities, err := parseRussianCitiesCSV(data)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx repository.PostgresRepository) error {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "fias_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "region", "latitude", "longitude"}),
		}).CreateInBatches(cities, citySeedBatchSize)
		if result.Error != nil {
			return fmt.Errorf("не удалось сохранить справочник: %w", result.Error)
		}
		return nil
	})
}

func parseRussianCitiesCSV(data []byte) ([]models.City, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать заголовок CSV: %w", err)
	}

	columns, err := requiredCityColumns(header)
	if err != nil {
		return nil, err
	}

	cities := make([]models.City, 0, 1200)
	seenFIAS := make(map[string]int, 1200)
	for rowNumber := 2; ; rowNumber++ {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("не удалось прочитать строку %d: %w", rowNumber, readErr)
		}

		city, parseErr := cityFromCSVRecord(record, columns)
		if parseErr != nil {
			return nil, fmt.Errorf("некорректная строка %d: %w", rowNumber, parseErr)
		}
		if firstRow, exists := seenFIAS[city.FIASID]; exists {
			return nil, fmt.Errorf("строки %d и %d содержат одинаковый fias_id %q", firstRow, rowNumber, city.FIASID)
		}
		seenFIAS[city.FIASID] = rowNumber
		cities = append(cities, city)
	}

	if len(cities) == 0 {
		return nil, fmt.Errorf("справочник не содержит городов")
	}
	return cities, nil
}

type cityCSVColumns struct {
	country        int
	regionType     int
	region         int
	areaType       int
	area           int
	city           int
	settlementType int
	settlement     int
	fiasID         int
	latitude       int
	longitude      int
	minimumLength  int
}

func requiredCityColumns(header []string) (cityCSVColumns, error) {
	indexes := make(map[string]int, len(header))
	for index, name := range header {
		indexes[strings.TrimSpace(name)] = index
	}

	required := []string{"country", "region_type", "region", "area_type", "area", "city", "settlement_type", "settlement", "fias_id", "geo_lat", "geo_lon"}
	for _, name := range required {
		if _, exists := indexes[name]; !exists {
			return cityCSVColumns{}, fmt.Errorf("в CSV отсутствует обязательная колонка %q", name)
		}
	}

	maximumIndex := 0
	for _, name := range required {
		if indexes[name] > maximumIndex {
			maximumIndex = indexes[name]
		}
	}
	return cityCSVColumns{
		country:        indexes["country"],
		regionType:     indexes["region_type"],
		region:         indexes["region"],
		areaType:       indexes["area_type"],
		area:           indexes["area"],
		city:           indexes["city"],
		settlementType: indexes["settlement_type"],
		settlement:     indexes["settlement"],
		fiasID:         indexes["fias_id"],
		latitude:       indexes["geo_lat"],
		longitude:      indexes["geo_lon"],
		minimumLength:  maximumIndex + 1,
	}, nil
}

func cityFromCSVRecord(record []string, columns cityCSVColumns) (models.City, error) {
	if len(record) < columns.minimumLength {
		return models.City{}, fmt.Errorf("ожидалось минимум %d колонок, получено %d", columns.minimumLength, len(record))
	}
	if country := strings.TrimSpace(record[columns.country]); country != "Россия" {
		return models.City{}, fmt.Errorf("страна %q не поддерживается", country)
	}

	regionName := strings.TrimSpace(record[columns.region])
	regionType := strings.TrimSpace(record[columns.regionType])
	region := formatCityRegion(regionName, regionType)
	if region == "" {
		return models.City{}, fmt.Errorf("регион не задан")
	}

	name := strings.TrimSpace(record[columns.city])
	if name == "" && strings.TrimSpace(record[columns.settlementType]) == "г" {
		name = strings.TrimSpace(record[columns.settlement])
	}
	if name == "" && strings.TrimSpace(record[columns.areaType]) == "г" {
		name = strings.TrimSpace(record[columns.area])
	}
	if name == "" && regionType == "г" {
		name = regionName
	}
	if name == "" {
		return models.City{}, fmt.Errorf("название города не задано")
	}

	fiasID := strings.TrimSpace(record[columns.fiasID])
	if fiasID == "" {
		return models.City{}, fmt.Errorf("fias_id не задан")
	}
	if len(fiasID) > 36 {
		return models.City{}, fmt.Errorf("fias_id %q длиннее 36 символов", fiasID)
	}

	latitude, longitude, err := parseCityCoordinates(record[columns.latitude], record[columns.longitude])
	if err != nil {
		return models.City{}, err
	}

	return models.City{Name: name, Region: region, FIASID: fiasID, Latitude: latitude, Longitude: longitude}, nil
}

func parseCityCoordinates(rawLatitude, rawLongitude string) (*float64, *float64, error) {
	rawLatitude = strings.TrimSpace(rawLatitude)
	rawLongitude = strings.TrimSpace(rawLongitude)
	if rawLatitude == "" && rawLongitude == "" {
		return nil, nil, nil
	}
	if rawLatitude == "" || rawLongitude == "" {
		return nil, nil, fmt.Errorf("координаты города должны быть заданы парой")
	}

	latitude, err := strconv.ParseFloat(rawLatitude, 64)
	if err != nil || math.IsNaN(latitude) || math.IsInf(latitude, 0) || latitude < -90 || latitude > 90 {
		return nil, nil, fmt.Errorf("некорректная широта %q", rawLatitude)
	}
	longitude, err := strconv.ParseFloat(rawLongitude, 64)
	if err != nil || math.IsNaN(longitude) || math.IsInf(longitude, 0) || longitude < -180 || longitude > 180 {
		return nil, nil, fmt.Errorf("некорректная долгота %q", rawLongitude)
	}

	return &latitude, &longitude, nil
}

func formatCityRegion(name, regionType string) string {
	name = strings.TrimSpace(name)
	regionType = strings.TrimSpace(regionType)
	if name == "" {
		return ""
	}
	if regionType == "г" || regionType == "" {
		return name
	}
	if regionType == "Респ" {
		return regionType + " " + name
	}
	return name + " " + regionType
}
