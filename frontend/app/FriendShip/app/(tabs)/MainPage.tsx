import BottomBar from '@/components/BottomBar';
import CategoryButton from '@/components/CategoryButton';
import { Event } from '@/components/EventCard';
import EventCarousel from '@/components/EventCarousel';
import EventModal from '@/components/EventModal';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useEvents } from '@/hooks/useEvents';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useEffect, useState } from 'react';
import { ImageBackground, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

type MainPageRouteProp = RouteProp<RootStackParamList, 'MainPage'>;
type MainPageNavigationProp = StackNavigationProp<RootStackParamList, 'MainPage'>;

const MainPage = () => {
  const route = useRoute<MainPageRouteProp>();
  const navigation = useNavigation<MainPageNavigationProp>();

  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [modalVisible, setModalVisible] = useState(false);

  const { sortingState, sortingActions } = useSearchState();
  const { 
    movieEvents, 
    gameEvents, 
    tableGameEvents, 
    otherEvents, 
    popularEvents, 
    newEvents, 
    searchResults 
  } = useEvents(sortingState);

  useEffect(() => {
    if (route.params?.searchQuery) {
      sortingActions.setSearchQuery(route.params.searchQuery);
    }
  }, [route.params?.searchQuery]);

  const addOnPressToEvents = (events: Event[]) =>
    events.map(event => ({
      ...event,
      onPress: () => {
        setSelectedEvent(event);
        setModalVisible(true);
      },
    }));

  const navigateToCategory = (
    category: 'movie' | 'game' | 'table_game' | 'other' | 'popular' | 'new',
    title: string,
    imageSource: any
  ) => {
    try {
      navigation.navigate('CategoryPage', {
        category,
        title,
        imageSource,
      });
    } catch (error) {
      console.error('❌ Navigation error:', error);
    }
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
          {sortingState.searchQuery.trim() ? (
            <>
              <CategoryButton
                title="Результаты поиска"
                imageSource={require('../../assets/images/category/search-pattern.png')}
                onPress={() => {}}
              />
              {searchResults.length > 0 ? (
                <EventCarousel events={addOnPressToEvents(searchResults)} />
              ) : (
                <View style={styles.noResultsContainer}>
                  <Text style={styles.noResultsText}>
                    Ничего не найдено по запросу "{sortingState.searchQuery}"
                  </Text>
                </View>
              )}
            </>
          ) : (
            <>
              <CategoryButton
                title="Популярные события"
                imageSource={require('../../assets/images/category/popular-pattern.png')}
                onPress={() =>
                  navigateToCategory(
                    'popular',
                    'Популярные события',
                    require('../../assets/images/category/popular-pattern.png')
                  )
                }
              />
              <EventCarousel events={addOnPressToEvents(popularEvents)} />

              <CategoryButton
                title="Новые события"
                imageSource={require('../../assets/images/category/new-pattern.png')}
                onPress={() =>
                  navigateToCategory(
                    'new',
                    'Новые события',
                    require('../../assets/images/category/new-pattern.png')
                  )
                }
              />
              <EventCarousel events={addOnPressToEvents(newEvents)} />

              <View style={styles.category}>
                <Text style={styles.textCategory}>Категории</Text>
              </View>

              {movieEvents.length > 0 && (
                <>
                  <CategoryButton
                    title="Фильмы"
                    imageSource={require('../../assets/images/category/movies-pattern.png')}
                    onPress={() =>
                      navigateToCategory(
                        'movie',
                        'Фильмы',
                        require('../../assets/images/category/movies-pattern.png')
                      )
                    }
                  />
                  <EventCarousel events={addOnPressToEvents(movieEvents)} />
                </>
              )}

              {gameEvents.length > 0 && (
                <>
                  <CategoryButton
                    title="Игры"
                    imageSource={require('../../assets/images/category/games-pattern.png')}
                    onPress={() =>
                      navigateToCategory(
                        'game',
                        'Игры',
                        require('../../assets/images/category/games-pattern.png')
                      )
                    }
                  />
                  <EventCarousel events={addOnPressToEvents(gameEvents)} />
                </>
              )}

              {tableGameEvents.length > 0 && (
                <>
                  <CategoryButton
                    title="Настольные игры"
                    imageSource={require('../../assets/images/category/table_games-pattern.png')}
                    onPress={() =>
                      navigateToCategory(
                        'table_game',
                        'Настольные игры',
                        require('../../assets/images/category/table_games-pattern.png')
                      )
                    }
                  />
                  <EventCarousel events={addOnPressToEvents(tableGameEvents)} />
                </>
              )}

              {otherEvents.length > 0 && (
                <>
                  <CategoryButton
                    title="Другое"
                    imageSource={require('../../assets/images/category/other-pattern.png')}
                    onPress={() =>
                      navigateToCategory(
                        'other',
                        'Другое',
                        require('../../assets/images/category/other-pattern.png')
                      )
                    }
                  />
                  <EventCarousel events={addOnPressToEvents(otherEvents)} />
                </>
              )}
            </>
          )}
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
  noResultsContainer: {
    backgroundColor: Colors.white,
    borderRadius: 40,
    margin: 8,
    padding: 12,
    alignItems: 'center',
    borderWidth: 3,
    borderColor: Colors.lightBlue,
  },
  noResultsText: {
    fontFamily: inter.regular,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
  },
});

export default MainPage;
