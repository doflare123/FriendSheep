import { RouteProp, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useMemo, useState } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event } from '@/components/event/EventCard';
import EventModal from '@/components/event/modal/EventModal';
import VerticalEventList from '@/components/event/VerticalEventList';
import CategorySearchBar from '@/components/search/CategorySearchBar';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useAllEvents } from '@/hooks/useAllEvents';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';

type AllEventsPageRouteProp = RouteProp<RootStackParamList, 'AllEventsPage'>;
type AllEventsPageNavigationProp = StackNavigationProp<RootStackParamList, 'AllEventsPage'>;

interface AllEventsPageProps {
  navigation: AllEventsPageNavigationProp;
}

const AllEventsPage: React.FC<AllEventsPageProps> = ({ navigation }) => {
  const route = useRoute<AllEventsPageRouteProp>();
  const { mode, groupId, userId } = route.params;
  
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [userEventFilter, setUserEventFilter] = useState<'current' | 'completed'>('current');

  const { sortingState, sortingActions } = useSearchState();

  const { 
    events,
    allEvents,
    userName,
    isLoading,
    error,
    refreshEvents 
  } = useAllEvents(mode, sortingState, mode === 'user' ? userEventFilter : undefined, groupId, userId);

  const hasCompletedEvents = useMemo(() => {
    return allEvents.some((event: any) => event._isFromRecent === true);
  }, [allEvents]);

  const shouldShowToggle = mode === 'user' && hasCompletedEvents;

  const addOnPressToEvents = (events: Event[]) => {
    return events.map(event => ({
      ...event,
      onPress: () => {
        setSelectedEvent(event);
        setModalVisible(true);
      },
      onSessionUpdate: refreshEvents,
    }));
  };

  const eventsToShow = useMemo(() => {
    return addOnPressToEvents(events);
  }, [events]);

  const getEventWordForm = (count: number): string => {
    if (count % 10 === 1 && count % 100 !== 11) {
      return 'событие';
    } else if ([2, 3, 4].includes(count % 10) && ![12, 13, 14].includes(count % 100)) {
      return 'события';
    } else {
      return 'событий';
    }
  };

  const handleToggleUserFilter = () => {
    setUserEventFilter(prev => prev === 'current' ? 'completed' : 'current');
  };

  const getTitle = () => {
    if (mode === 'group') {
      return 'События группы';
    }
    if (userName) {
      return `События ${userName}`;
    }
    return 'Мои события';
  };

  const getEmptyText = () => {
    if (sortingState.searchQuery.trim()) {
      return `Ничего не найдено по запросу "${sortingState.searchQuery}"`;
    }
    
    if (mode === 'user') {
      if (userName) {
        return `У ${userName} пока нет текущих событий`;
      }
      return userEventFilter === 'current' 
        ? 'У вас пока нет текущих событий'
        : 'У вас пока нет завершённых событий';
    } else {
      return 'В группе пока нет активных событий';
    }
  };

  if (isLoading) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        
        <CategorySection
          title={getTitle()}
          showBackButton
          onBackPress={() => navigation.goBack()}
        />
        
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color={Colors.lightBlue} />
          <Text style={styles.loadingText}>Загрузка событий...</Text>
        </View>
        
        <BottomBar />
      </SafeAreaView>
    );
  }

  if (error) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        
        <CategorySection
          title={getTitle()}
          showBackButton
          onBackPress={() => navigation.goBack()}
        />
        
        <View style={styles.centerContainer}>
          <Text style={styles.errorTitle}>Ошибка загрузки</Text>
          <Text style={styles.errorText}>{error}</Text>
        </View>
        
        <BottomBar />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      
      <CategorySection
        title={getTitle()}
        showBackButton
        onBackPress={() => navigation.goBack()}
        customActionButton={shouldShowToggle ? {
          icon: userEventFilter === 'completed' 
            ? require('@/assets/images/profile/finished.png')
            : require('@/assets/images/profile/current.png'),
          onPress: handleToggleUserFilter,
        } : undefined}
      />

      <View style={styles.searchContainer}>
        <CategorySearchBar 
          sortingState={sortingState} 
          sortingActions={sortingActions}
          showCategoryFilter={true}
        />
      </View>

      {eventsToShow.length > 0 && (
        <View style={styles.resultsHeader}>
          <Text style={styles.resultsCount}>
            Найдено: {eventsToShow.length} {getEventWordForm(eventsToShow.length)}
          </Text>
        </View>
      )}

      <View style={styles.contentContainer}>
        {eventsToShow.length > 0 ? (
          <VerticalEventList events={eventsToShow} />
        ) : (
          <View style={styles.noEventsContainer}>
            <Text style={styles.noEventsText}>
              {getEmptyText()}
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
          onSessionUpdate={refreshEvents}
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
  contentContainer: {
    flex: 1,
    paddingHorizontal: 16,
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
    color: Colors.grey,
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
    color: Colors.grey,
  },
  noEventsContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 16,
  },
  noEventsText: {
    fontFamily: Montserrat.regular,
    fontSize: 20,
    textAlign: 'center',
    color: Colors.blue2,
    marginBottom: 4,
  },
});

export default AllEventsPage;