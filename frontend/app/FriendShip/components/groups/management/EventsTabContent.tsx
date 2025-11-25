import EventCard, { Event } from '@/components/event/EventCard';
import GroupEventsSearchBar from '@/components/groups/GroupEventsSearchBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { sortEventsByDate, sortEventsByParticipants } from '@/utils/eventSorting';
import { highlightEventTitle } from '@/utils/textHighlight';
import React, { useMemo, useState } from 'react';
import {
  FlatList,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';

interface EventsTabContentProps {
  events: Event[];
  onCreateEvent: () => void;
  onEditEvent: (eventId: string) => void;
}

const EventsTabContent: React.FC<EventsTabContentProps> = ({
  events,
  onCreateEvent,
  onEditEvent,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const { sortingState, sortingActions } = useSearchState();

  const categoryMapping: Record<string, Event['category']> = {
    'Игры': 'game',
    'Фильмы': 'movie',
    'Настолки': 'table_game',
    'Другое': 'other'
  };

  const filteredAndSortedEvents = useMemo(() => {
    let filtered = events;

    if (searchQuery.trim()) {
      filtered = filtered.filter(event => 
        event.title.toLowerCase().includes(searchQuery.toLowerCase())
      );

      filtered = filtered.map(event => 
        highlightEventTitle(event, searchQuery)
      );
    }

    if (!sortingState.checkedCategories.includes('Все')) {
      filtered = filtered.filter(event => {
        return sortingState.checkedCategories.some(cat => {
          const mappedCategory = categoryMapping[cat];
          return mappedCategory && event.category === mappedCategory;
        });
      });
    }

    if (sortingState.sortByDate !== 'none') {
      filtered = sortEventsByDate(filtered, sortingState.sortByDate);
    }

    if (sortingState.sortByParticipants !== 'none') {
      filtered = sortEventsByParticipants(filtered, sortingState.sortByParticipants);
    }

    return filtered;
  }, [events, searchQuery, sortingState.checkedCategories, sortingState.sortByDate, sortingState.sortByParticipants]);

  const renderEventCard = ({ item }: { item: Event }) => (
    <View style={styles.cardContainer}>
      <EventCard
        {...item}
        onPress={() => onEditEvent(item.id)}
      />
    </View>
  );

  return (
    <View style={styles.container}>
      <View style={styles.searchContainer}>
        <GroupEventsSearchBar
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          sortingState={sortingState}
          sortingActions={sortingActions}
        />
      </View>

      <TouchableOpacity style={styles.createButton} onPress={onCreateEvent}>
        <Text style={styles.createButtonText}>Создать новое событие</Text>
      </TouchableOpacity>

      <FlatList
        data={filteredAndSortedEvents}
        keyExtractor={(item) => item.id}
        renderItem={renderEventCard}
        showsVerticalScrollIndicator={false}
        contentContainerStyle={styles.listContainer}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>
              {searchQuery.trim() || !sortingState.checkedCategories.includes('Все')
                ? 'События не найдены' 
                : 'У группы пока нет событий'
              }
            </Text>
          </View>
        }
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  searchContainer: {
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: Colors.white,
  },
  createButton: {
    marginHorizontal: 20,
    marginBottom: 16,
    backgroundColor: Colors.lightBlue3,
    borderRadius: 40,
    paddingVertical: 6,
    alignItems: 'center',
    elevation: 2,
    shadowColor: Colors.black,
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  createButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.white,
  },
  listContainer: {
    paddingBottom: 20,
  },
  cardContainer: {
    marginBottom: 16,
    alignItems: 'center',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingTop: 50,
  },
  emptyText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
    textAlign: 'center',
  },
});

export default EventsTabContent;