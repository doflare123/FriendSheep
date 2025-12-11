import React from 'react';
import { FlatList, StyleSheet, View } from 'react-native';
import GroupCard, { Group } from './GroupCard';

interface VerticalGroupListProps {
  groups: Group[];
}

const VerticalGroupList: React.FC<VerticalGroupListProps> = ({ groups }) => {
  return (
    <FlatList
      data={groups}
      keyExtractor={(item) => item.id}
      renderItem={({ item }) => (
        <View style={styles.cardWrapper}>
          <GroupCard {...item} />
        </View>
      )}
      showsVerticalScrollIndicator={false}
      contentContainerStyle={styles.listContent}
    />
  );
};

const styles = StyleSheet.create({
  listContent: {
    paddingBottom: 16,
  },
  cardWrapper: {
    marginBottom: 16,
  },
});

export default VerticalGroupList;