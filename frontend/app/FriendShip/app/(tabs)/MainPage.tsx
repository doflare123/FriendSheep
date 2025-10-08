import BottomBar from '@/components/BottomBar';
import { Event } from '@/components/EventCard';
import EventCarousel from '@/components/EventCarousel';
import EventModal from '@/components/EventModal';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { Montserrat } from '@/constants/Montserrat';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import { useEvents } from '@/hooks/useEvents';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useEffect, useState } from 'react';
import { Image, ScrollView, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
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
    <SafeAreaView style={{ flex: 1, backgroundColor: Colors.white }}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />

        <ScrollView contentContainerStyle={{ paddingBottom: 80 }}>             
          {sortingState.searchQuery.trim() ? (
            <>
              <View style={styles.headerContainer}>
                <Text style={styles.headerTitle}>Поиск по событиям</Text>
              </View>
              <View style={styles.resultsHeader}>
                <Text style={styles.resultsText}>Результаты поиска:</Text>
                <Image
                  source={require('../../assets/images/line.png')}
                  style={{resizeMode: 'none'}}
                />
              </View>
              {searchResults.length > 0 ? (
                <EventCarousel events={addOnPressToEvents(searchResults)} />
              ) : (
                <Text style={styles.noResultsText}>
                  Ничего не найдено по запросу "{sortingState.searchQuery}"
                </Text>
              )}
            </>
          ) : (
            <>
              <View style={styles.headerContainer}>
                <Text style={styles.headerTitle}>С возвращением!</Text>
              </View>
              <Image
                source={require('../../assets/images/wave.png')}
                style={{resizeMode: 'none', marginBottom: -20}}
              />
              <View style={styles.resultsHeader}>
                <View style={{flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}}>
                  <Text style={styles.resultsText}>Популярные:</Text>
                  <TouchableOpacity onPress={() =>
                  navigateToCategory(
                    'popular',
                    'Популярные события',
                    require('../../assets/images/category/popular-pattern.png')
                  )
                }>
                    <Image
                      source={require('../../assets/images/arrow.png')}
                      style={{resizeMode: 'contain', width: 30, height: 30}}
                    />
                  </TouchableOpacity>
                </View>
                <Image
                    source={require('../../assets/images/line.png')}
                    style={{resizeMode: 'none'}}
                />
              </View>
              <EventCarousel events={addOnPressToEvents(popularEvents)} />

              <View style={styles.resultsHeader}>
                <View style={{flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}}>
                  <Text style={styles.resultsText}>Новые:</Text>
                  <TouchableOpacity onPress={() =>
                  navigateToCategory(
                    'new',
                    'Новые события',
                    require('../../assets/images/category/new-pattern.png')
                  )
                }>
                    <Image
                      source={require('../../assets/images/arrow.png')}
                      style={{resizeMode: 'contain', width: 30, height: 30}}
                    />
                  </TouchableOpacity>
                </View>
                <Image
                    source={require('../../assets/images/line.png')}
                    style={{resizeMode: 'none'}}
                />
              </View>
              <EventCarousel events={addOnPressToEvents(newEvents)} />

              <View style={[styles.resultsHeader, {marginBottom: 16}]}>
                <Text style={[styles.resultsText, {textAlign: 'center'}]}>Категории</Text>
                <Image
                    source={require('../../assets/images/line2.png')}
                    style={{resizeMode: 'none'}}
                />
              </View>

              {movieEvents.length > 0 && (
                <>
                  <View style={styles.resultsHeader}>
                    <View style={{flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}}>
                      <Text style={styles.resultsText}>Медиа:</Text>
                      <TouchableOpacity onPress={() =>
                      navigateToCategory(
                        'movie',
                        'Фильмы',
                        require('../../assets/images/category/movies-pattern.png')
                      )
                    }>
                        <Image
                          source={require('../../assets/images/arrow.png')}
                          style={{resizeMode: 'contain', width: 30, height: 30}}
                        />
                      </TouchableOpacity>
                    </View>
                    <Image
                        source={require('../../assets/images/line.png')}
                        style={{resizeMode: 'none'}}
                    />
                  </View>
                  <EventCarousel events={addOnPressToEvents(movieEvents)} />
                </>
              )}

              {gameEvents.length > 0 && (
                <>
                  <View style={styles.resultsHeader}>
                    <View style={{flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}}>
                      <Text style={styles.resultsText}>Игры:</Text>
                      <TouchableOpacity onPress={() =>
                      navigateToCategory(
                        'game',
                        'Игры',
                        require('../../assets/images/category/games-pattern.png')
                      )
                    }>
                        <Image
                          source={require('../../assets/images/arrow.png')}
                          style={{resizeMode: 'contain', width: 30, height: 30}}
                        />
                      </TouchableOpacity>
                    </View>
                    <Image
                        source={require('../../assets/images/line.png')}
                        style={{resizeMode: 'none'}}
                    />
                  </View>
                  <EventCarousel events={addOnPressToEvents(gameEvents)} />
                </>
              )}

              {tableGameEvents.length > 0 && (
                <>
                  <View style={styles.resultsHeader}>
                    <View style={{flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}}>
                      <Text style={styles.resultsText}>Настольные игры:</Text>
                      <TouchableOpacity onPress={() =>
                      navigateToCategory(
                        'table_game',
                        'Настольные игры',
                        require('../../assets/images/category/table_games-pattern.png')
                      )
                    }>
                        <Image
                          source={require('../../assets/images/arrow.png')}
                          style={{resizeMode: 'contain', width: 30, height: 30}}
                        />
                      </TouchableOpacity>
                    </View>
                    <Image
                        source={require('../../assets/images/line.png')}
                        style={{resizeMode: 'none'}}
                    />
                  </View>
                  <EventCarousel events={addOnPressToEvents(tableGameEvents)} />
                </>
              )}

              {otherEvents.length > 0 && (
                <>
                  <View style={styles.resultsHeader}>
                    <View style={{flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}}>
                      <Text style={styles.resultsText}>Другое:</Text>
                      <TouchableOpacity onPress={() =>
                      navigateToCategory(
                        'other',
                        'Другое',
                        require('../../assets/images/category/other-pattern.png')
                      )
                    }>
                        <Image
                          source={require('../../assets/images/arrow.png')}
                          style={{resizeMode: 'contain', width: 30, height: 30}}
                        />
                      </TouchableOpacity>
                    </View>
                    <Image
                        source={require('../../assets/images/line.png')}
                        style={{resizeMode: 'none'}}
                    />
                  </View>
                  <EventCarousel events={addOnPressToEvents(otherEvents)} />
                </>
              )}
            </>
          )}
        </ScrollView>
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
  noResultsText: {
    fontFamily: Montserrat.regular,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
  },
  headerContainer: {
    backgroundColor: Colors.lightBlue2,
    paddingVertical: 20,
    alignItems: 'center',
    marginBottom: 12,
  },
  headerTitle: {
    fontFamily: Montserrat_Alternates.medium,
    fontSize: 24,
    color: Colors.white,
    textTransform: 'none',
  },
  resultsHeader: {
    paddingHorizontal: 16,
  },
  resultsText: {
    fontFamily: Montserrat_Alternates.medium,
    fontSize: 20,
    color: Colors.blue2,
    marginBottom: 4
  },
});

export default MainPage;
