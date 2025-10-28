import React from 'react';
import { FlatList, StyleSheet, View } from 'react-native';
import EventCard, { Event } from './EventCard';

interface VerticalEventListProps {
  events: Event[];
}

const VerticalEventList: React.FC<VerticalEventListProps> = ({ events }) => {
  return (
    <FlatList
      data={events}
      keyExtractor={(item) => item.id}
      showsVerticalScrollIndicator={false}
      renderItem={({ item }) => (
        <View style={styles.cardContainer}>
          <EventCard {...item} />
        </View>
      )}
    />
  );
};

const styles = StyleSheet.create({
  cardContainer: {
    marginBottom: 16,
    alignItems: 'center',
  },
});

export default VerticalEventList;