import React from 'react';
import { FlatList, StyleSheet, View } from 'react-native';
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

const EventCarousel: React.FC<EventCarouselProps> = ({ events }) => {
  return (
    <FlatList
      horizontal
      data={events}
      keyExtractor={(item) => item.id}
      showsHorizontalScrollIndicator={false}
      contentContainerStyle={styles.carousel}
      renderItem={({ item }) => (
        <View style={{ marginRight: 16 }}>
          <EventCard {...item} />
        </View>
      )}
    />
  );
};

const styles = StyleSheet.create({
  carousel: {
    paddingHorizontal: 16,
  },
});

export default EventCarousel;
