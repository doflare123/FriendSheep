import { RouteProp, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useState } from 'react';
import { ImageBackground, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import CategoryButton from '@/components/CategoryButton';
import CategorySearchBar from '@/components/CategorySearchBar';
import { Event } from '@/components/EventCard';
import EventModal from '@/components/EventModal';
import TopBar from '@/components/TopBar';
import VerticalEventList from '@/components/VerticalEventList';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useEvents } from '@/hooks/useEvents';
import { useSearchState } from '@/hooks/useSearchState';

type RootStackParamList = {
  CategoryPage: {
    category: 'movie' | 'game' | 'table_game' | 'other';
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
  const { category, title, imageSource } = route.params;
  
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [modalVisible, setModalVisible] = useState(false);

  const { sortingState, sortingActions } = useSearchState();
  const { 
    movieEvents, 
    gameEvents, 
    tableGameEvents, 
    otherEvents,
    searchResults 
  } = useEvents(sortingState);

  const getCategoryEvents = (): Event[] => {
    switch (category) {
      case 'movie':
        return movieEvents;
      case 'game':
        return gameEvents;
      case 'table_game':
        return tableGameEvents;
      case 'other':
        return otherEvents;
      default:
        return [];
    }
  };

  const addOnPressToEvents = (events: Event[]) => {
    return events.map(event => ({
      ...event,
      onPress: () => {
        setSelectedEvent(event);
        setModalVisible(true);
      },
    }));
  };

  const eventsToShow = sortingState.searchQuery.trim() 
    ? searchResults.filter(event => event.category === category)
    : getCategoryEvents();

  return (
    <SafeAreaView style={styles.container} edges={['top', 'left', 'right']}>
      <ImageBackground
        source={require('../../assets/images/wallpaper.png')}
        style={styles.backgroundImage}
        resizeMode="cover"
      >
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />

        <CategoryButton
            title={title}
            imageSource={imageSource}
            onPress={() => navigation.goBack()}
        />

        <View style={styles.searchContainer}>
          <CategorySearchBar sortingState={sortingState} sortingActions={sortingActions} />
        </View>

        <View style={styles.contentContainer}>
          {eventsToShow.length > 0 ? (
            <VerticalEventList events={addOnPressToEvents(eventsToShow)} />
          ) : (
            <View style={styles.noEventsContainer}>
              <Text style={styles.noEventsText}>
                {sortingState.searchQuery.trim() 
                  ? `Ничего не найдено по запросу "${sortingState.searchQuery}"` 
                  : `В категории "${title}" пока нет событий`
                }
              </Text>
            </View>
          )}
        </View>
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
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  backgroundImage: {
    flex: 1,
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