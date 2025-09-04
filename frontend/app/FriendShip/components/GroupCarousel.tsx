import React from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';
import GroupCard, { Group } from './GroupCard';

interface GroupCarouselProps {
  groups: Group[];
  actionText: string;
  actionColor?: string[];
}

const GroupCarousel: React.FC<GroupCarouselProps> = ({ 
  groups, 
  actionText,
  actionColor 
}) => {
  if (groups.length === 0) {
    return null;
  }

  return (
    <View style={styles.container}>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.scrollContent}
        decelerationRate="fast"
        snapToInterval={328}
        snapToAlignment="start"
      >
        {groups.map((group, index) => (
          <View key={group.id} style={[
            styles.cardWrapper,
            index === 0 && styles.firstCard,
            index === groups.length - 1 && styles.lastCard,
          ]}>
            <GroupCard 
              {...group} 
              actionText={actionText}
              actionColor={actionColor}
            />
          </View>
        ))}
      </ScrollView>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    marginBottom: 8,
  },
  scrollContent: {
    paddingVertical: 8,
  },
  cardWrapper: {
    marginHorizontal: 4,
  },
  firstCard: {
    marginLeft: 8,
  },
  lastCard: {
    marginRight: 8,
  },
});

export default GroupCarousel;