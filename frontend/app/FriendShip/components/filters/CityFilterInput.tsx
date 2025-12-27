import { Montserrat } from '@/constants/Montserrat';
import { RussianCities } from '@/constants/RussianCities';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useState } from 'react';
import { FlatList, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';

interface CityFilterInputProps {
  value: string;
  onChangeText: (text: string) => void;
}

const CityFilterInput: React.FC<CityFilterInputProps> = ({ value, onChangeText }) => {
  const colors = useThemedColors();
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);

  const filterCities = (query: string) => {
    if (query.length < 2) {
      setSuggestions([]);
      setShowSuggestions(false);
      return;
    }

    const lowerQuery = query.toLowerCase();
    const filtered = RussianCities
      .filter(city => city.toLowerCase().includes(lowerQuery))
      .slice(0, 5);

    setSuggestions(filtered);
    setShowSuggestions(filtered.length > 0);
  };

  const handleCitySelect = (city: string) => {
    onChangeText(city);
    setShowSuggestions(false);
    setSuggestions([]);
  };

  return (
    <View style={styles.container}>
      <Text style={[styles.dropdownTitle, { color: colors.black }]}>
        Фильтрация по городу
      </Text>
      
      <TextInput
        style={[styles.cityInput, { color: colors.black, backgroundColor: colors.veryLightGrey }]}
        placeholder="Введите город..."
        value={value}
        onChangeText={(text) => {
          onChangeText(text);
          filterCities(text);
        }}
        onBlur={() => {
          setTimeout(() => setShowSuggestions(false), 200);
        }}
        placeholderTextColor={colors.grey}
      />

      {showSuggestions && suggestions.length > 0 && (
        <View style={[styles.suggestionsContainer, { backgroundColor: colors.white }]}>
          <FlatList
            data={suggestions}
            keyExtractor={(item, index) => `${item}-${index}`}
            renderItem={({ item }) => (
              <TouchableOpacity
                style={[styles.suggestionItem, { borderBottomColor: colors.veryLightGrey }]}
                onPress={() => handleCitySelect(item)}
              >
                <Text style={[styles.suggestionText, { color: colors.black }]}>
                  {item}
                </Text>
              </TouchableOpacity>
            )}
            scrollEnabled={false}
            nestedScrollEnabled={true}
          />
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    marginBottom: 8,
  },
  dropdownTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    marginBottom: 8,
  },
  cityInput: {
    borderRadius: 20,
    paddingHorizontal: 10,
    paddingVertical: 8,
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  suggestionsContainer: {
    marginTop: 4,
    borderRadius: 12,
    overflow: 'hidden',
    elevation: 3,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.15,
    shadowRadius: 3.84,
  },
  suggestionItem: {
    paddingVertical: 12,
    paddingHorizontal: 12,
    borderBottomWidth: 1,
  },
  suggestionText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
});

export default CityFilterInput;