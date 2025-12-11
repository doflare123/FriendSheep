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
import UserCard, { User } from '@/components/profile/UserCard';
import SearchResultsSection from '@/components/search/SearchResultsSection';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { useUserSearch } from '@/hooks/useUserSearch';
import { useUserSearchState } from '@/hooks/useUserSearchState';
import { RootStackParamList } from '@/navigation/types';
import { createUserWithHighlightedText } from '@/utils/userUtils';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';

const UserSearchPage: React.FC = () => {
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const { sortingState: globalSortingState, sortingActions: globalSortingActions } = useSearchState();
  const { searchState, searchActions } = useUserSearchState();

  const {
    users,
    isLoading,
    isLoadingMore,
    error,
    hasMore,
    totalUsers,
    searchUsers,
    loadMore,
    resetSearch,
  } = useUserSearch();

  useEffect(() => {
    const query = searchState.searchQuery.trim();
    
    if (query) {
      console.log('[UserSearchPage] 🔍 Выполняем поиск:', query);
      searchUsers(query, 1, false);
    } else {
      console.log('[UserSearchPage] 🧹 Очистка результатов поиска');
      resetSearch();
    }
  }, [searchState.searchQuery, searchUsers, resetSearch]);

  const formattedUsers: User[] = useMemo(() => {
    return users.map(user =>
      createUserWithHighlightedText(user, searchState.searchQuery)
    );
  }, [users, searchState.searchQuery]);

  const handleUserPress = (userId: string) => {
    console.log('[UserSearchPage] Переход к профилю:', userId);
    navigation.navigate('ProfilePage', { userId });
  };

  const handleLoadMore = () => {
    if (!isLoadingMore && hasMore && searchState.searchQuery.trim()) {
      console.log('[UserSearchPage] Загружаем больше пользователей');
      loadMore(searchState.searchQuery);
    }
  };

  const getUserWordForm = (count: number): string => {
    if (count % 10 === 1 && count % 100 !== 11) {
      return 'пользователь';
    } else if ([2, 3, 4].includes(count % 10) && ![12, 13, 14].includes(count % 100)) {
      return 'пользователя';
    } else {
      return 'пользователей';
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
          <Text style={styles.loadingText}>Поиск пользователей...</Text>
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

    if (!searchState.searchQuery.trim()) {
      return (
        <View style={styles.centerContainer}>
          <Text style={styles.emptyText}>Введите имя пользователя</Text>
          <Text style={styles.emptySubtext}>
            для начала поиска
          </Text>
        </View>
      );
    }

    return (
      <View style={styles.centerContainer}>
        <Text style={styles.emptyText}>Пользователи не найдены</Text>
        <Text style={styles.emptySubtext}>
          Попробуйте изменить запрос
        </Text>
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={globalSortingState} sortingActions={globalSortingActions} />

      <View style={styles.content}>
        <SearchResultsSection
          title="Поиск по профилям"
          searchQuery={searchState.searchQuery}
          hasResults={formattedUsers.length > 0}
          showWave
        >
          {totalUsers > 0 && (
            <View style={styles.resultsHeader}>
              <Text style={styles.resultsCount}>
                Найдено: {totalUsers} {getUserWordForm(totalUsers)}
              </Text>
            </View>
          )}

          <View style={styles.contentContainer}>
            <FlatList
              data={formattedUsers}
              keyExtractor={(item) => item.id}
              renderItem={({ item }) => (
                <View style={styles.cardWrapper}>
                  <UserCard
                    {...item}
                    onPress={() => handleUserPress(item.id)}
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
  content: {
    flex: 1,
  },
  contentContainer: {
    flex: 1,
    paddingHorizontal: 16,
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

export default UserSearchPage;