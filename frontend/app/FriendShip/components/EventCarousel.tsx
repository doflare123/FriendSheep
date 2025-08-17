import React from 'react';
import { Dimensions, FlatList, StyleSheet, View } from 'react-native';
import EventCard from './EventCard';

interface EventCarouselProps {
  events: {
    id: string;
    title: string;
    date: string;
    genres: string[];
    participants: string;
    duration: string;
    imageUri: string;
  }[];
}

const { width } = Dimensions.get('window');
const CARD_WIDTH = 320;
const CARD_SPACING = 16;

const EventCarousel: React.FC<EventCarouselProps> = ({ events }) => {
  return (
    <FlatList
      horizontal
      data={events}
      keyExtractor={(item) => item.id}
      showsHorizontalScrollIndicator={false}
      snapToInterval={CARD_WIDTH + CARD_SPACING}
      decelerationRate="fast"
      contentContainerStyle={styles.carousel}
      renderItem={({ item, index }) => (
        <View
          style={{
            marginLeft: index === 0 ? CARD_SPACING : 0,
            marginRight: CARD_SPACING,
          }}
        >
          <EventCard {...item} />
        </View>
      )}
    />
  );
};

const styles = StyleSheet.create({
  carousel: {
    paddingVertical: 16,
  },
});

export default EventCarousel;
