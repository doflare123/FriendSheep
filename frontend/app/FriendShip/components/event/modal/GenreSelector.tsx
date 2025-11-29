import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React, { useState } from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

interface GenreSelectorProps {
  selected: string[];
  onToggle: (genre: string) => void;
  genres?: string[];
}

const DEFAULT_GENRES = [
  'Драма', 'Комедия', 'Боевик', 'Триллер', 'Ужасы', 'Фантастика',
  'Детектив', 'Приключения', 'Романтика', 'Криминал', 'Военный',
  'Исторический', 'Биография', 'Документальный', 'Анимация', 'Семейный',
  'Мюзикл', 'Вестерн', 'Спорт', 'Фэнтези'
];

const GenreSelector: React.FC<GenreSelectorProps> = ({
  selected,
  onToggle,
  genres = DEFAULT_GENRES,
}) => {
  const [showDropdown, setShowDropdown] = useState(false);

  return (
    <View>
      <TouchableOpacity
        style={styles.dropdownButton}
        onPress={() => setShowDropdown(!showDropdown)}
      >
        <Text style={[styles.dropdownText, selected.length > 0 && styles.dropdownTextActive]}>
          {selected.length > 0 ? selected.join(', ') : 'Выберите жанры...'}
        </Text>
        <Image
          source={require('@/assets/images/event_card/back.png')}
          style={[
            styles.dropdownArrow,
            { transform: [{ rotate: showDropdown ? '270deg' : '90deg' }] }
          ]}
        />
      </TouchableOpacity>

      {showDropdown && (
        <View style={styles.genreDropdown}>
          {genres.map((genre) => (
            <TouchableOpacity
              key={genre}
              style={styles.genreItem}
              onPress={() => onToggle(genre)}
            >
              <View style={styles.checkbox}>
                {selected.includes(genre) && <View style={styles.checkboxSelected} />}
              </View>
              <Text style={styles.genreText}>{genre}</Text>
            </TouchableOpacity>
          ))}
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  dropdownButton: {
    flexDirection: 'row',
    alignItems: 'center',
    borderBottomWidth: 1,
    borderBottomColor: Colors.grey,
    paddingVertical: 8,
    marginBottom: 8,
  },
  dropdownText: {
    flex: 1,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  dropdownTextActive: {
    color: Colors.black,
  },
  dropdownArrow: {
    width: 20,
    height: 20,
    tintColor: Colors.grey,
  },
  genreDropdown: {
    backgroundColor: Colors.white2,
    marginHorizontal: 20,
    borderRadius: 20,
    marginBottom: 16,
  },
  genreItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  checkbox: {
    width: 20,
    height: 20,
    borderRadius: 6,
    borderWidth: 2,
    borderColor: Colors.lightBlue2,
    marginRight: 12,
    backgroundColor: Colors.white,
    justifyContent: 'center',
    alignItems: 'center',
  },
  checkboxSelected: {
    width: 12,
    height: 12,
    borderRadius: 2,
    backgroundColor: Colors.blue2,
  },
  genreText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
  },
});

export default GenreSelector;