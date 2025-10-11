import { RouteProp, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useMemo, useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event } from '@/components/event/EventCard';
import EventModal from '@/components/event/EventModal';
import VerticalEventList from '@/components/event/VerticalEventList';
import CategorySearchBar from '@/components/search/CategorySearchBar';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useEvents } from '@/hooks/useEvents';
import { useSearchState } from '@/hooks/useSearchState';
import { sortEventsByDate, sortEventsByParticipants } from '@/utils/eventSorting';
import { filterEventsByCategories } from '@/utils/eventUtils';
import { highlightEventTitle } from '@/utils/textHighlight';

type RootStackParamList = {
  CategoryPage: {
    category: 'movie' | 'game' | 'table_game' | 'other' | 'popular' | 'new';
    title: string;
    imageSource: any;
  };
};

type CategoryPageRouteProp = RouteProp<RootStackParamList, 'CategoryPage'>;
type CategoryPageNavigationProp = StackNavigationProp<RootStackParamList, 'CategoryPage'>;

interface CategoryPageProps {
  navigation: CategoryPageNavigationProp;
}

const CategoryPage: React.FC<CategoryPageProps> = ({ navigation }) => {
  const route = useRoute<CategoryPageRouteProp>();
  const { category, title } = route.params;
  
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [modalVisible, setModalVisible] = useState(false);

  const { sortingState: globalSortingState, sortingActions: globalSortingActions } = useSearchState();

  const { sortingState: categorySortingState, sortingActions: categorySortingActions } = useSearchState();

  const { 
    movieEvents, 
    gameEvents, 
    tableGameEvents, 
    otherEvents,
    popularEvents,
    newEvents,
  } = useEvents(globalSortingState);

  const isSpecialCategory = category === 'popular' || category === 'new';

  const getBaseCategoryEvents = (): Event[] => {
    switch (category) {
      case 'movie':
        return movieEvents;
      case 'game':
        return gameEvents;
      case 'table_game':
        return tableGameEvents;
      case 'other':
        return otherEvents;
      case 'popular':
        return popularEvents;
      case 'new':
        return newEvents;
      default:
        return [];
    }
  };

  const sortedCategoryEvents = useMemo((): Event[] => {
    let events = getBaseCategoryEvents();
    
    console.log('Исходные события:', events.length);
    console.log('Категория:', category);
    console.log('Сортировка по дате:', categorySortingState.sortByDate);
    console.log('Сортировка по участникам:', categorySortingState.sortByParticipants);

    if (isSpecialCategory && categorySortingState.checkedCategories) {
      events = filterEventsByCategories(events, categorySortingState.checkedCategories);
      console.log('После фильтрации по категориям:', events.length);
    }

    if (categorySortingState.sortByDate !== 'none') {
      events = sortEventsByDate(events, categorySortingState.sortByDate);
      console.log('После сортировки по дате:', events.map(e => e.date));
    }

    if (categorySortingState.sortByParticipants !== 'none') {
      events = sortEventsByParticipants(events, categorySortingState.sortByParticipants);
      console.log('После сортировки по участникам:', events.map(e => e.currentParticipants));
    }
    
    return events;
  }, [
    category, 
    categorySortingState.sortByDate, 
    categorySortingState.sortByParticipants, 
    categorySortingState.checkedCategories,
    movieEvents, 
    gameEvents, 
    tableGameEvents, 
    otherEvents,
    popularEvents,
    newEvents
  ]);

  const addOnPressToEvents = (events: Event[]) => {
    return events.map(event => ({
      ...event,
      onPress: () => {
        setSelectedEvent(event);
        setModalVisible(true);
      },
    }));
  };

  const eventsToShow = useMemo(() => {
    if (categorySortingState.searchQuery.trim()) {
      let filtered = sortedCategoryEvents.filter(event =>
        event.title.toLowerCase().includes(categorySortingState.searchQuery.toLowerCase())
      );
      
      filtered = filtered.map(event => 
        highlightEventTitle(event, categorySortingState.searchQuery)
      );
      
      return filtered;
    } else {
      return sortedCategoryEvents;
    }
  }, [sortedCategoryEvents, categorySortingState.searchQuery]);

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={globalSortingState} sortingActions={globalSortingActions} />
      
      <CategorySection
        title={title}
        showBackButton
        onBackPress={() => navigation.goBack()}
      />

      <View style={styles.searchContainer}>
        <CategorySearchBar 
          sortingState={categorySortingState} 
          sortingActions={categorySortingActions}
          showCategoryFilter={isSpecialCategory}
        />
      </View>

      <View style={styles.contentContainer}>
        {eventsToShow.length > 0 ? (
          <VerticalEventList events={addOnPressToEvents(eventsToShow)} />
        ) : (
          <View style={styles.noEventsContainer}>
            <Text style={styles.noEventsText}>
              {categorySortingState.searchQuery.trim() 
                ? `Ничего не найдено по запросу "${categorySortingState.searchQuery}"` 
                : `В категории "${title}" пока нет событий`
              }
            </Text>
          </View>
        )}
      </View>

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
  searchContainer: {
    paddingHorizontal: 16,
    paddingBottom: 12,
  },
  contentContainer: {
    flex: 1,
    paddingHorizontal: 16,
  },
  noEventsContainer: {
    flex: 1,
    paddingHorizontal: 16,
  },
  noEventsText: {
    fontFamily: inter.regular,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
    backgroundColor: Colors.white,
    borderRadius: 40,
    padding: 12,
    borderWidth: 3,
    borderColor: Colors.lightBlue,
  },
});

export default CategoryPage;