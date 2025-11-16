import React, { useEffect, useMemo } from 'react';
import {
  ActivityIndicator,
  FlatList,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import GroupCard, { Group } from '@/components/groups/GroupCard';
import SearchResultsSection from '@/components/search/SearchResultsSection';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useGroupSearch } from '@/hooks/useGroupSearch';
import { useGroupSearchState } from '@/hooks/useGroupSearchState';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';

const CATEGORY_MAP: Record<string, string> = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};

const GroupSearchPage: React.FC = () => {
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const { searchState, searchActions } = useGroupSearchState();
  const { sortingState: globalSortingState, sortingActions: globalSortingActions } = useSearchState();
  
  const {
    groups,
    isLoading,
    isLoadingMore,
    error,
    hasMore,
    totalGroups,
    searchGroups,
    loadMore,
    resetSearch,
  } = useGroupSearch();

  useEffect(() => {
    console.log('[GroupSearchPage] Параметры поиска изменились, выполняем поиск');
    searchGroups(
      searchState.searchQuery,
      searchState.checkedCategories,
      searchState.sortByParticipants,
      searchState.sortByRegistration,
      1,
      false
    );
  }, [
    searchState.searchQuery,
    searchState.checkedCategories,
    searchState.sortByParticipants,
    searchState.sortByRegistration,
  ]);

  const formattedGroups: Group[] = useMemo(() => {
    return groups.map(group => ({
      id: group.id.toString(),
      name: group.name,
      participantsCount: group.count,
      description: group.description,
      imageUri: group.image,
      categories: group.category
        .map(cat => CATEGORY_MAP[cat])
        .filter(Boolean) as any[],
      isPrivate: group.isPrivate,
    }));
  }, [groups]);

  const handleGroupPress = (groupId: string) => {
    console.log('[GroupSearchPage] Переход к группе:', groupId);
    navigation.navigate('GroupPage', { groupId });
  };

  const handleLoadMore = () => {
    if (!isLoadingMore && hasMore) {
      console.log('[GroupSearchPage] Загружаем больше групп');
      loadMore(
        searchState.searchQuery,
        searchState.checkedCategories,
        searchState.sortByParticipants,
        searchState.sortByRegistration
      );
    }
  };

  const renderFooter = () => {
    if (!isLoadingMore) return null;

    return (
      <View style={styles.footerLoader}>
        <ActivityIndicator size="small" color={Colors.lightBlue} />
        <Text style={styles.footerText}>Загрузка...</Text>
      </View>
    );
  };

  const renderEmpty = () => {
    if (isLoading) {
      return (
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color={Colors.lightBlue} />
          <Text style={styles.loadingText}>Поиск групп...</Text>
        </View>
      );
    }

    if (error) {
      return (
        <View style={styles.centerContainer}>
          <Text style={styles.errorText}>Ошибка: {error}</Text>
        </View>
      );
    }

    return (
      <View style={styles.centerContainer}>
        <Text style={styles.emptyText}>Группы не найдены</Text>
        <Text style={styles.emptySubtext}>
          Попробуйте изменить параметры поиска
        </Text>
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={globalSortingState} sortingActions={globalSortingActions} />

      <View style={{ flex: 1 }}>
        <SearchResultsSection
          title="Поиск по группам"
          searchQuery={searchState.searchQuery}
          hasResults={formattedGroups.length > 0}
          showWave
        >
          {totalGroups > 0 && (
            <View style={styles.resultsHeader}>
              <Text style={styles.resultsCount}>
                Найдено: {totalGroups} {totalGroups === 1 ? 'группа' : 'групп'}
              </Text>
            </View>
          )}

          <View style={styles.contentContainer}>
            <FlatList
              data={formattedGroups}
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
              ListEmptyComponent={renderEmpty}
              ListFooterComponent={renderFooter}
              onEndReached={handleLoadMore}
              onEndReachedThreshold={0.5}
            />
          </View>
        </SearchResultsSection>
      </View>

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
    flexGrow: 1,
  },
  cardWrapper: {
    marginBottom: 16,
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
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 40,
  },
  loadingText: {
    marginTop: 12,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  errorText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.red,
    textAlign: 'center',
    paddingHorizontal: 32,
  },
  emptyText: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.grey,
    marginBottom: 8,
  },
  emptySubtext: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
    textAlign: 'center',
  },
  footerLoader: {
    paddingVertical: 20,
    alignItems: 'center',
  },
  footerText: {
    marginTop: 8,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
  },
});

export default GroupSearchPage;