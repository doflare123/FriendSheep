import CategoryButton from '@/components/CategoryButton';
import { Event } from '@/components/EventCard';
import EventCarousel from '@/components/EventCarousel';
import EventModal from '@/components/EventModal';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useNavigation } from '@react-navigation/native';
import React, { useMemo, useState } from 'react';
import { ImageBackground, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import BottomBar from '../../components/BottomBar';
import TopBar from '../../components/TopBar';

export interface SortingState {
  checkedCategories: string[];
  sortByDate: 'asc' | 'desc';
  sortByParticipants: 'asc' | 'desc';
  sortByRegistration: 'asc' | 'desc';
}

export interface SortingActions {
  setCheckedCategories: (categories: string[]) => void;
  setSortByDate: (order: 'asc' | 'desc') => void;
  setSortByParticipants: (order: 'asc' | 'desc') => void;
  setSortByRegistration: (order: 'asc' | 'desc') => void;
  toggleCategoryCheckbox: (category: string) => void;
}

const dummyEvents: Event[] = [
  {
    id: '1',
    title: 'Крестный отец',
    date: '12.02.2004 19:20',
    imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
    description: "ЭЩКЕРЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕ",
    genres: ['Драма', 'Криминал'],
    currentParticipants: 52,
    maxParticipants: 52,
    duration: '175 минут',
    typeEvent: 'Фильм',
    typePlace: 'online',
    eventPlace: 'https://cinema.com',
    publisher: 'Paramount Pictures',
    publicationDate: 1972,
    ageRating: '18+',
    category: 'movie',
  },
  {
    id: '2',
    title: 'Матрица',
    date: '15.03.2004',
    imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
    description: "мяу",
    genres: ['Фантастика'],
    currentParticipants: 48,
    maxParticipants: 50,
    duration: '136 минут',
    typeEvent: 'Фильм',
    typePlace: 'offline',
    eventPlace: 'Кинотеатр «Октябрь»',
    publisher: 'Warner Bros',
    publicationDate: 1999,
    ageRating: '16+',
    category: 'movie',
  },
  {
    id: '3',
    title: 'CS:GO турнир',
    date: '10.01.2004',
    imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
    description: "Киберспортивный турнир",
    genres: ['Шутер'],
    currentParticipants: 32,
    maxParticipants: 64,
    duration: '240 минут',
    typeEvent: 'Турнир',
    typePlace: 'online',
    eventPlace: 'Steam',
    publisher: 'Valve',
    publicationDate: 2012,
    ageRating: '16+',
    category: 'game',
  },
  {
    id: '4',
    title: 'Monopoly Tournament',
    date: '20.04.2004',
    imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
    description: "Турнир по настольной игре Монополия",
    genres: ['Стратегия'],
    currentParticipants: 8,
    maxParticipants: 12,
    duration: '180 минут',
    typeEvent: 'Турнир',
    typePlace: 'offline',
    eventPlace: 'Игровой клуб "Кубик"',
    publisher: 'Hasbro',
    publicationDate: 1935,
    ageRating: '8+',
    category: 'table_game',
  },
];

const MainPage = () => {
  const navigation = useNavigation();

  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [modalVisible, setModalVisible] = useState(false);

  const [checkedCategories, setCheckedCategories] = useState<string[]>(['Все']);
  const [sortByDate, setSortByDate] = useState<'asc' | 'desc'>('asc');
  const [sortByParticipants, setSortByParticipants] = useState<'asc' | 'desc'>('asc');
  const [sortByRegistration, setSortByRegistration] = useState<'asc' | 'desc'>('asc');

  const toggleCategoryCheckbox = (category: string) => {
    if (category === 'Все') {
      setCheckedCategories(['Все']);
    } else {
      const filtered = checkedCategories.filter(c => c !== 'Все');
      if (checkedCategories.includes(category)) {
        const newCategories = filtered.filter(c => c !== category);
        setCheckedCategories(newCategories.length === 0 ? ['Все'] : newCategories);
      } else {
        setCheckedCategories([...filtered, category]);
      }
    }
  };

  const categoryMapping: Record<string, Event['category']> = {
    'Фильмы': 'movie',
    'Игры': 'game', 
    'Настолки': 'table_game',
    'Другое': 'other'
  };

  const sortedEvents = useMemo(() => {
    let filtered = [...dummyEvents];

    if (!checkedCategories.includes('Все')) {
      const mappedCategories = checkedCategories
        .map(cat => categoryMapping[cat])
        .filter(Boolean);
      
      filtered = filtered.filter(event => mappedCategories.includes(event.category));
    }

    filtered.sort((a, b) => {
      const participantsA = a.currentParticipants;
      const participantsB = b.currentParticipants;

      if (sortByParticipants === 'asc') {
        const result = participantsA - participantsB;
        return result;
      } else {
        const result = participantsB - participantsA;
        return result;
      }
    });

    return filtered;
  }, [dummyEvents, checkedCategories, sortByDate, sortByParticipants]);

  const sortingState: SortingState = {
    checkedCategories,
    sortByDate,
    sortByParticipants,
    sortByRegistration,
  };

  const sortingActions: SortingActions = {
    setCheckedCategories,
    setSortByDate,
    setSortByParticipants,
    setSortByRegistration,
    toggleCategoryCheckbox,
  };

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: Colors.white }} edges={['top', 'left', 'right']}>
      <ImageBackground
        source={require('../../assets/images/wallpaper.png')}
        style={{ flex: 1 }}
        resizeMode="cover"
      >
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        
        <ScrollView contentContainerStyle={{ paddingBottom: 80 }}>
          <CategoryButton
            title="Популярные события"
            imageSource={require('../../assets/images/category/popular-pattern.png')}
            onPress={() => {}}
          />
          <EventCarousel
            events={sortedEvents.map(event => ({
              ...event,
              onPress: () => {
                setSelectedEvent(event);
                setModalVisible(true);
              },
            }))}
          />

          <CategoryButton
            title="Новые события"
            imageSource={require('../../assets/images/category/new-pattern.png')}
            onPress={() => {}}
          />
          <EventCarousel
            events={sortedEvents.map(event => ({
              ...event,
              onPress: () => {
                setSelectedEvent(event);
                setModalVisible(true);
              },
            }))}
          />

          <View style={styles.category}>
            <Text style={styles.textCategory}>Категории</Text>
          </View>
          <CategoryButton
            title="Фильмы"
            imageSource={require('../../assets/images/category/movies-pattern.png')}
            onPress={() => {}}
          />
          <CategoryButton
            title="Игры"
            imageSource={require('../../assets/images/category/games-pattern.png')}
            onPress={() => {}}
          />
        </ScrollView>
      </ImageBackground>
      <BottomBar />

      {selectedEvent && (
        <EventModal
          visible={modalVisible}
          onClose={() => setModalVisible(false)}
          event={selectedEvent}
        />
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  category: {
    backgroundColor: Colors.white,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    overflow: 'hidden',
    margin: 8,
    marginTop: 16,
    height: 40,
    width: '90%',
    alignSelf: 'center',
  },
  textCategory: {
    color: Colors.blue,
    fontSize: 20,
    fontFamily: inter.bold,
  },
  debugPanel: {
    backgroundColor: 'rgba(255, 255, 255, 0.9)',
    margin: 8,
    padding: 12,
    borderRadius: 10,
  },
  debugText: {
    fontSize: 12,
    color: '#333',
    marginBottom: 2,
  },
  testButton: {
    backgroundColor: '#007AFF',
    padding: 8,
    borderRadius: 5,
    marginTop: 8,
    alignItems: 'center',
  },
  testButtonText: {
    color: 'white',
    fontSize: 12,
    fontWeight: 'bold',
  },
});

export default MainPage;