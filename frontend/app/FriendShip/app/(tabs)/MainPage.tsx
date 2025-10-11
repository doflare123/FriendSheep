import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event } from '@/components/event/EventCard';
import EventCarousel from '@/components/event/EventCarousel';
import EventModal from '@/components/event/modal/EventModal';
import PageHeader from '@/components/PageHeader';
import SearchResultsSection from '@/components/search/SearchResultsSection';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { useEvents } from '@/hooks/useEvents';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useEffect, useState } from 'react';
import { ScrollView } from 'react-native';
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
          <SearchResultsSection
            title="Поиск по событиям"
            searchQuery={sortingState.searchQuery}
            hasResults={searchResults.length > 0}
          >
            <EventCarousel events={addOnPressToEvents(searchResults)} />
          </SearchResultsSection>
        ) : (
          <>
            <PageHeader title="С возвращением!" showWave />

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
                    'Фильмы',
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

export default MainPage;