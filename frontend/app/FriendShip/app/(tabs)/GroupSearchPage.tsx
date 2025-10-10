import React, { useMemo } from 'react';
import { FlatList, StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import GroupCard, { Group } from '@/components/groups/GroupCard';
import SearchResultsSection from '@/components/search/SearchResultsSection';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { useGroupSearchState } from '@/hooks/useGroupSearchState';
import { useSearchState } from '@/hooks/useSearchState';

const mockGroups: Group[] = [
  {
    id: '1',
    name: 'Группа крутышек',
    participantsCount: 52,
    description: 'Мы клуб фанатов соника. Присоединяйтесь к нам, если вы его любите!',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
  {
    id: '2',
    name: 'Группа крутышек',
    participantsCount: 52,
    description: 'Мы клуб фанатов соника. Присоединяйтесь к нам, если вы его любите!',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
  {
    id: '3',
    name: 'Группа крутышек',
    participantsCount: 52,
    description: 'Мы клуб фанатов соника. Присоединяйтесь к нам, если вы его любите!',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
  {
    id: '4',
    name: 'Группа крутышек',
    participantsCount: 52,
    description: 'Мы клуб фанатов соника. Присоединяйтесь к нам, если вы его любите!',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
];

const filterGroupsByCategories = (groups: Group[], categories: string[]): Group[] => {
  if (categories.includes('Все') || categories.length === 0) {
    return groups;
  }
  return groups;
};

const sortGroupsByParticipants = (groups: Group[], order: 'asc' | 'desc' | 'none') => {
  if (order === 'none') return groups;
  
  return [...groups].sort((a, b) => {
    if (order === 'asc') {
      return a.participantsCount - b.participantsCount;
    } else {
      return b.participantsCount - a.participantsCount;
    }
  });
};

const sortGroupsByRegistration = (groups: Group[], order: 'asc' | 'desc' | 'none') => {
  if (order === 'none') return groups;
  
  return [...groups].sort((a, b) => {
    if (order === 'asc') {
      return parseInt(a.id) - parseInt(b.id);
    } else {
      return parseInt(b.id) - parseInt(a.id);
    }
  });
};

const GroupSearchPage: React.FC = () => {
  const { searchState, searchActions } = useGroupSearchState();
  const { sortingState: globalSortingState, sortingActions: globalSortingActions } = useSearchState();

  const createGroupWithHighlightedText = (group: Group, query: string): Group => {
    if (!query.trim()) return group;

    const lowerQuery = query.toLowerCase();

    const lowerName = group.name.toLowerCase();
    const nameIndex = lowerName.indexOf(lowerQuery);
    let highlightedName = undefined;
    
    if (nameIndex !== -1) {
      highlightedName = {
        before: group.name.substring(0, nameIndex),
        match: group.name.substring(nameIndex, nameIndex + query.length),
        after: group.name.substring(nameIndex + query.length)
      };
    }

    const lowerDescription = group.description.toLowerCase();
    const descIndex = lowerDescription.indexOf(lowerQuery);
    let highlightedDescription = undefined;
    
    if (descIndex !== -1) {
      highlightedDescription = {
        before: group.description.substring(0, descIndex),
        match: group.description.substring(descIndex, descIndex + query.length),
        after: group.description.substring(descIndex + query.length)
      };
    }

    return {
      ...group,
      highlightedName,
      highlightedDescription
    };
  };

  const filteredAndSortedGroups = useMemo((): Group[] => {
    let groups = [...mockGroups];
    
    if (searchState.searchQuery.trim()) {
      groups = groups.filter(group =>
        group.name.toLowerCase().includes(searchState.searchQuery.toLowerCase()) ||
        group.description.toLowerCase().includes(searchState.searchQuery.toLowerCase())
      );
      
      groups = groups.map(group => 
        createGroupWithHighlightedText(group, searchState.searchQuery)
      );
    }

    if (searchState.checkedCategories && searchState.checkedCategories.length > 0) {
      groups = filterGroupsByCategories(groups, searchState.checkedCategories);
    }

    if (searchState.sortByParticipants !== 'none') {
      groups = sortGroupsByParticipants(groups, searchState.sortByParticipants);
    }

    if (searchState.sortByRegistration !== 'none') {
      groups = sortGroupsByRegistration(groups, searchState.sortByRegistration);
    }
    
    return groups;
  }, [
    searchState.searchQuery,
    searchState.checkedCategories,
    searchState.sortByParticipants,
    searchState.sortByRegistration,
  ]);

  const handleGroupPress = (groupId: string) => {
    console.log('Group pressed:', groupId);
  };

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={globalSortingState} sortingActions={globalSortingActions} />

      <SearchResultsSection
        title="Поиск по группам"
        searchQuery={searchState.searchQuery}
        hasResults={filteredAndSortedGroups.length > 0}
        showWave
      >
        <View style={styles.contentContainer}>
          <FlatList
            data={filteredAndSortedGroups}
            keyExtractor={(item) => item.id}
            renderItem={({ item }) => (
              <View style={styles.cardWrapper}>
                <GroupCard
                  {...item}
                  onPress={() => handleGroupPress(item.id)}
                />
              </View>
            )}
            showsVerticalScrollIndicator={false}
            contentContainerStyle={styles.listContent}
          />
        </View>
      </SearchResultsSection>

      <BottomBar />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  contentContainer: {
    flex: 1,
    paddingHorizontal: 8,
  },
  listContent: {
    paddingBottom: 16,
  },
  cardWrapper: {
    marginBottom: 16,
  },
});

export default GroupSearchPage;