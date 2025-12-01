import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React, { useMemo, useState } from 'react';
import { Image, ScrollView, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';

interface GenreSelectorProps {
  selected: string[];
  onToggle: (genre: string) => void;
  genres?: string[];
}

const GENRE_CATEGORIES = {
  media: {
    id: 'media',
    label: 'Медиа',
    icon: require('@/assets/images/event_card/movie.png'),
    genres: [
      'Драма', 'Комедия', 'Боевик', 'Триллер', 'Ужасы', 'Фантастика',
      'Детектив', 'Приключения', 'Романтика', 'Криминал', 'Военный',
      'Исторический', 'Биография', 'Документальный', 'Анимация', 'Семейный',
      'Мюзикл', 'Вестерн', 'Спорт', 'Фэнтези', 'Нуар', 'Киберпанк',
      'Мистика', 'Психологический', 'Постапокалипсис', 'Супергерои', 'Антиутопия',
      'Сатира', 'Пародия', 'Абсурд', 'Артхаус', 'Слэшер', 'Сплэттер',
      'Монстры', 'Зомби', 'Вампиры', 'Космоопера', 'Стимпанк', 'Дизельпанк',
      'Самураи', 'Якудза', 'Гангстеры', 'Шпионы', 'Катастрофа', 'Survival',
      'Подростковый', 'Мелодрама', 'Нео-нуар', 'Эротика', 'Философский',
      'Социальный', 'Политический', 'Экранизация', 'Ремейк', 'Сиквел'
    ]
  },
  game: {
    id: 'game',
    label: 'Видеоигры',
    icon: require('@/assets/images/event_card/game.png'),
    genres: [
      'Action', 'RPG', 'Стратегия', 'Шутер', 'Платформер', 'Симулятор',
      'Хоррор', 'Выживание', 'Песочница', 'MOBA', 'Battle Royale', 
      'Головоломка', 'Гонки', 'Файтинг', 'Метроидвания', 'Roguelike',
      'Инди', 'MMO', 'Стелс', 'Аркада', 'JRPG', 'CRPG', 'Action-RPG',
      'Hack and Slash', 'Souls-like', 'Tower Defense', 'RTS', 'TBS',
      '4X', 'Grand Strategy', 'Tycoon', 'Menedžment', 'City Builder',
      'FPS', 'TPS', 'Tactical Shooter', 'Hero Shooter', 'Looter Shooter',
      'Racing Sim', 'Kart Racing', 'Walking Simulator', 'Visual Novel',
      'Dating Sim', 'Rhythm', 'Music', 'Party', 'Sports', 'Fighting',
      'Beat \'em up', 'Bullet Hell', 'Twin-stick Shooter', 'Shoot \'em up',
      'Point-and-click', 'Adventure', 'Interactive Fiction', 'Narrative',
      'Survival Horror', 'Psychological Horror', 'Open World', 'Linear',
      'Co-op', 'PvP', 'PvE', 'Auto Battler', 'Idle Game', 'Clicker'
    ]
  },
  table_game: {
    id: 'table_game',
    label: 'Настолки',
    icon: require('@/assets/images/event_card/table_game.png'),
    genres: [
      'Стратегическая', 'Экономическая', 'Кооперативная', 'Карточная',
      'Дедукция', 'Блеф', 'Викторина', 'Ролевая', 'Абстрактная',
      'Варгейм', 'Семейная', 'Детская', 'Вечериночная', 'Legacy',
      'Deck-building', 'Worker Placement', 'Area Control', 'Eurogame',
      'Ameritrash', 'Drafting', 'Tile Placement', 'Set Collection',
      'Hand Management', 'Engine Building', 'Push Your Luck', 'Memory',
      'Dexterity', 'Real-time', 'Word Game', 'Trivia', 'Roll and Write',
      'Dice', 'Dungeon Crawler', 'Campaign', 'Storytelling', 'Negotiation',
      'Trading', 'Auction', 'Bidding', 'Hidden Role', 'Social Deduction',
      'Party Game', 'Team-based', 'Solo', '2-player', 'Miniatures',
      'Escape Room', 'Puzzle', 'Pattern Recognition', 'Speed', 'Reaction',
      'Coordination', 'Communication', 'Drawing', 'Acting', 'Singing'
    ]
  },
  other: {
    id: 'other',
    label: 'Мероприятия',
    icon: require('@/assets/images/event_card/other.png'),
    genres: [
      'Концерт', 'Выставка', 'Фестиваль', 'Мастер-класс', 'Лекция',
      'Турнир', 'Квиз', 'Квест', 'Вечеринка', 'Кулинария',
      'Спортивное', 'Культурное', 'Образовательное', 'Благотворительность',
      'Нетворкинг', 'Stand-up', 'Караоке', 'Танцы', 'Творчество',
      'Рок-концерт', 'Джаз', 'Классика', 'Электроника', 'Хип-хоп', 'Поп',
      'Арт-выставка', 'Фотовыставка', 'Музей', 'Галерея', 'Инсталляция',
      'Кинофестиваль', 'Музыкальный фестиваль', 'Гастрономический', 'Косплей',
      'Аниме', 'Комикс-кон', 'Игровая конвенция', 'Стрим', 'Митап',
      'Воркшоп', 'Семинар', 'Конференция', 'Вебинар', 'Тренинг', 'Коучинг',
      'Киберспорт', 'Боулинг', 'Бильярд', 'Дартс', 'Пейнтбол', 'Лазертаг',
      'Йога', 'Медитация', 'Фитнес', 'Бег', 'Велопрогулка', 'Поход',
      'Экскурсия', 'Сити-тур', 'Дегустация', 'Винный вечер', 'Барбекю',
      'Бранч', 'Ужин', 'Пикник', 'Тимбилдинг', 'Корпоратив', 'День рождения',
      'Свадьба', 'Baby shower', 'Встреча выпускников', 'Книжный клуб',
      'Кинопоказ', 'Театр', 'Опера', 'Балет', 'Цирк', 'Планетарий'
    ]
  }
};

const DEFAULT_GENRES = [
  ...GENRE_CATEGORIES.media.genres,
  ...GENRE_CATEGORIES.game.genres,
  ...GENRE_CATEGORIES.table_game.genres,
  ...GENRE_CATEGORIES.other.genres
];

const MAX_GENRES = 9;

const GenreSelector: React.FC<GenreSelectorProps> = ({
  selected,
  onToggle,
  genres,
}) => {
  const useDefaultGenres = !genres || genres.length === 0;
  
  const [showDropdown, setShowDropdown] = useState(false);
  const [activeCategory, setActiveCategory] = useState<string>('media');
  const [searchQuery, setSearchQuery] = useState('');


   const categoryGenres = useDefaultGenres
    ? GENRE_CATEGORIES[activeCategory as keyof typeof GENRE_CATEGORIES].genres
    : genres;

  const displayGenres = useMemo(() => {
    if (!searchQuery.trim()) {
      return categoryGenres;
    }
    
    const query = searchQuery.toLowerCase().trim();
    
    return categoryGenres.filter(genre => 
      genre.toLowerCase().includes(query)
    );
  }, [searchQuery, categoryGenres]);

  const handleGenreToggle = (genre: string) => {
    if (selected.includes(genre)) {
      onToggle(genre);
      return;
    }

    if (selected.length >= MAX_GENRES) {
      return;
    }

    onToggle(genre);
  };

  return (
    <View>
      <TouchableOpacity
        style={styles.dropdownButton}
        onPress={() => setShowDropdown(!showDropdown)}
      >
        <Text style={[styles.dropdownText, selected.length > 0 && styles.dropdownTextActive]}>
          {selected.length > 0 ? selected.join(', ') : 'Выберите жанры... *'}
        </Text>
        <Image
          source={require('@/assets/images/event_card/back.png')}
          style={[
            styles.dropdownArrow,
            { transform: [{ rotate: showDropdown ? '270deg' : '90deg' }] }
          ]}
        />
      </TouchableOpacity>

      {selected.length > 0 && (
        <Text style={styles.genreCounter}>
          {selected.length} / {MAX_GENRES}
        </Text>
      )}

      {showDropdown && (
        <View style={styles.genreDropdown}>
          {useDefaultGenres && (
            <View style={styles.categorySelector}>
              <View style={styles.categoriesContainer}>
                {Object.values(GENRE_CATEGORIES).map((category) => {
                  const isSelected = activeCategory === category.id;
                  
                  return (
                    <TouchableOpacity
                      key={category.id}
                      style={[
                        styles.categoryButton,
                        isSelected && styles.categorySelected
                      ]}
                      onPress={() => setActiveCategory(category.id)}
                    >
                      <Image source={category.icon} style={styles.categoryIcon} />
                    </TouchableOpacity>
                  );
                })}
              </View>
            </View>
          )}

          <View style={styles.searchContainer}>
            <Image
              source={require('@/assets/images/top_bar/search_bar/search.png')}
              style={styles.searchIcon}
            />
            <TextInput
              style={styles.searchInput}
              placeholder="Поиск жанров..."
              placeholderTextColor={Colors.grey}
              value={searchQuery}
              onChangeText={setSearchQuery}
              autoCapitalize="none"
              autoCorrect={false}
            />
            {searchQuery.length > 0 && (
              <TouchableOpacity
                onPress={() => setSearchQuery('')}
                style={styles.clearButton}
              >
                <Image
                  source={require('@/assets/images/event_card/back.png')}
                  style={styles.clearIcon}
                />
              </TouchableOpacity>
            )}
          </View>

          <ScrollView 
            style={styles.genreScrollView}
            nestedScrollEnabled
            showsVerticalScrollIndicator={false}
          >
            {displayGenres.length > 0 ? (
              displayGenres.map((genre) => (
                <TouchableOpacity
                  key={genre}
                  style={styles.genreItem}
                  onPress={() => handleGenreToggle(genre)}
                >
                  <View style={styles.checkbox}>
                    {selected.includes(genre) && <View style={styles.checkboxSelected} />}
                  </View>
                  <Text style={styles.genreText}>{genre}</Text>
                </TouchableOpacity>
              ))
            ) : (
              <View style={styles.noResultsContainer}>
                <Text style={styles.noResultsText}>Ничего не найдено</Text>
              </View>
            )}
          </ScrollView>
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
    paddingHorizontal: 0,
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
  genreCounter: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.grey,
    marginBottom: 8,
  },
  genreDropdown: {
    backgroundColor: Colors.white2,
    borderRadius: 20,
    marginBottom: 16,
    maxHeight: 350,
  },
  categorySelector: {
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
  },
  categoriesContainer: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 30,
    paddingHorizontal: 16,
  },
  categoryButton: {
    width: 40,
    height: 40,
    backgroundColor: Colors.lightLightGrey,
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  categorySelected: {
    backgroundColor: Colors.lightBlue,
  },
  categoryIcon: {
    width: 28,
    height: 28,
    resizeMode: 'contain',
  },
  searchContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
  },
  searchIcon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
    tintColor: Colors.grey,
    marginRight: 12,
  },
  searchInput: {
    flex: 1,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    padding: 0,
  },
  clearButton: {
    padding: 4,
  },
  clearIcon: {
    width: 16,
    height: 16,
    resizeMode: 'contain',
    tintColor: Colors.grey,
    transform: [{ rotate: '180deg' }],
  },
  genreScrollView: {
    maxHeight: 220,
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
    borderColor: Colors.blue2,
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
  noResultsContainer: {
    paddingVertical: 20,
    alignItems: 'center',
  },
  noResultsText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
  },
});

export default GenreSelector;