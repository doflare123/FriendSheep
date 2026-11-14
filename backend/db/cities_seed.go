package db

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
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
			DoUpdates: clause.AssignmentColumns([]string{"name", "region"}),
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
	minimumLength  int
}

func requiredCityColumns(header []string) (cityCSVColumns, error) {
	indexes := make(map[string]int, len(header))
	for index, name := range header {
		indexes[strings.TrimSpace(name)] = index
	}

	required := []string{"country", "region_type", "region", "area_type", "area", "city", "settlement_type", "settlement", "fias_id"}
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

	return models.City{Name: name, Region: region, FIASID: fiasID}, nil
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
