import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event } from '@/components/event/EventCard';
import EventModal from '@/components/event/modal/EventModal';
import VerticalEventList from '@/components/event/VerticalEventList';
import PageHeader from '@/components/PageHeader';
import SearchResultsSection from '@/components/search/SearchResultsSection';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useEvents } from '@/hooks/useEvents';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import React, { useEffect, useMemo, useState } from 'react';
import { ActivityIndicator, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

type MainPageRouteProp = RouteProp<RootStackParamList, 'MainPage'>;
type MainPageNavigationProp = NativeStackNavigationProp<RootStackParamList, 'MainPage'>;

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
    searchResults,
    isLoading,
    error,
  } = useEvents(sortingState);

  useEffect(() => {
    if (route.params?.searchQuery) {
      sortingActions.setSearchQuery(route.params.searchQuery);
    }
  }, [route.params?.searchQuery]);

  const hasAnyEvents = useMemo(() => {
    return (
      popularEvents.length > 0 ||
      newEvents.length > 0 ||
      movieEvents.length > 0 ||
      gameEvents.length > 0 ||
      tableGameEvents.length > 0 ||
      otherEvents.length > 0
    );
  }, [popularEvents, newEvents, movieEvents, gameEvents, tableGameEvents, otherEvents]);

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

  const getEventWordForm = (count: number): string => {
    if (count % 10 === 1 && count % 100 !== 11) {
      return 'событие';
    } else if ([2, 3, 4].includes(count % 10) && ![12, 13, 14].includes(count % 100)) {
      return 'события';
    } else {
      return 'событий';
    }
  };

  if (isLoading) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color={Colors.lightBlue3} />
          <Text style={styles.loadingText}>
            Загрузка событий...
          </Text>
        </View>
        <BottomBar />
      </SafeAreaView>
    );
  }

  if (error) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <View style={styles.centerContainer}>
          <Text style={styles.errorTitle}>
            Ошибка загрузки
          </Text>
          <Text style={styles.errorText}>
            {error}
          </Text>
        </View>
        <BottomBar />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />

      {sortingState.searchQuery.trim() ? (
        <View style={{ flex: 1 }}>
          <SearchResultsSection
            title="Поиск по событиям"
            searchQuery={sortingState.searchQuery}
            hasResults={searchResults.length > 0}
          >
            {searchResults.length > 0 && (
              <View style={styles.resultsHeader}>
                <Text style={styles.resultsCount}>
                  Найдено: {searchResults.length} {getEventWordForm(searchResults.length)}
                </Text>
              </View>
            )}

            <View style={styles.searchResultsContainer}>
              <VerticalEventList events={addOnPressToEvents(searchResults)} />
            </View>
          </SearchResultsSection>
        </View>
      ) : (
        <ScrollView contentContainerStyle={styles.scrollContent}>
          <PageHeader title="С возвращением!" showWave />

          {!hasAnyEvents ? (
            <View style={styles.noEventsContainer}>
              <Text style={styles.noEventsTitle}>
                Пока нет доступных событий
              </Text>
              <Text style={styles.noEventsSubtext}>
                Создайте первое событие в вашей группе,{'\n'}
                чтобы оно появилось здесь
              </Text>
            </View>
          ) : (
            <>
              {popularEvents.length > 0 && (
                <CategorySection
                  title="Популярные:"
                  events={addOnPressToEvents(popularEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'popular',
                      'Популярные события',
                      require('../../assets/images/category/popular-pattern.png')
                    )
                  }
                />
              )}

              {newEvents.length > 0 && (
                <CategorySection
                  title="Новые:"
                  events={addOnPressToEvents(newEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'new',
                      'Новые события',
                      require('../../assets/images/category/new-pattern.png')
                    )
                  }
                />
              )}

              <CategorySection
                title="Категории"
                events={[]}
                centerTitle
                showLineVariant="line2"
                marginBottom={16}
              />

              {movieEvents.length > 0 && (
                <CategorySection
                  title="Медиа:"
                  events={addOnPressToEvents(movieEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'movie',
                      'Медиа',
                      require('../../assets/images/category/movies-pattern.png')
                    )
                  }
                />
              )}

              {gameEvents.length > 0 && (
                <CategorySection
                  title="Игры:"
                  events={addOnPressToEvents(gameEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'game',
                      'Игры',
                      require('../../assets/images/category/games-pattern.png')
                    )
                  }
                />
              )}

              {tableGameEvents.length > 0 && (
                <CategorySection
                  title="Настольные игры:"
                  events={addOnPressToEvents(tableGameEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'table_game',
                      'Настольные игры',
                      require('../../assets/images/category/table_games-pattern.png')
                    )
                  }
                />
              )}

              {otherEvents.length > 0 && (
                <CategorySection
                  title="Другое:"
                  events={addOnPressToEvents(otherEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'other',
                      'Другое',
                      require('../../assets/images/category/other-pattern.png')
                    )
                  }
                />
              )}
            </>
          )}
        </ScrollView>
      )}

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
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  scrollContent: {
    paddingBottom: 80,
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  loadingText: {
    marginTop: 12,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  errorTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.red,
    marginBottom: 8,
  },
  errorText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    textAlign: 'center',
  },
  resultsHeader: {
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  resultsCount: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
  },
  searchResultsContainer: {
    flex: 1,
    paddingHorizontal: 16,
  },
  noEventsContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
    paddingVertical: 60,
  },
  noEventsTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.blue2,
    textAlign: 'center',
    marginBottom: 12,
  },
  noEventsSubtext: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
    textAlign: 'center',
    lineHeight: 24,
  },
});

export default MainPage;