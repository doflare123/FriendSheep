import type { SearchGroupItem, SearchGroupsParams } from '@/api/services/group';
import { groupSearchService } from '@/api/services/group';
import { useCallback, useState } from 'react';
import { GroupCategory } from './useGroupSearchState';

export const useGroupSearch = () => {
  const [groups, setGroups] = useState<SearchGroupItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [totalGroups, setTotalGroups] = useState(0);

  const searchGroups = useCallback(async (
    searchQuery: string,
    categories: GroupCategory[],
    sortByParticipants: 'asc' | 'desc' | 'none',
    sortByRegistration: 'asc' | 'desc' | 'none',
    page: number = 1,
    append: boolean = false
  ) => {
    try {
      if (append) {
        setIsLoadingMore(true);
      } else {
        setIsLoading(true);
        setGroups([]);
      }
      setError(null);

      const params: SearchGroupsParams = {
        page,
      };

      if (searchQuery.trim()) {
        params.name = searchQuery.trim();
      }

      if (categories.length > 0) {
        const categoryMap: Record<GroupCategory, string> = {
          'movie': 'Фильмы',
          'game': 'Игры',
          'table_game': 'Настольные игры',
          'other': 'Другое',
        };
        params.category = categoryMap[categories[0]];
      }

      if (sortByParticipants !== 'none') {
        params.sort_by = 'members';
        params.order = sortByParticipants;
      } else if (sortByRegistration !== 'none') {
        params.sort_by = 'date';
        params.order = sortByRegistration;
      }

      console.log('[useGroupSearch] Выполняем поиск:', params);

      const response = await groupSearchService.searchGroups(params);

      if (append) {
        setGroups(prev => [...prev, ...response.groups]);
      } else {
        setGroups(response.groups);
      }

      setCurrentPage(response.page);
      setHasMore(response.has_more);
      setTotalGroups(response.total);

      console.log('[useGroupSearch] Поиск завершён:', {
        groupsCount: response.groups.length,
        totalGroups: response.total,
        hasMore: response.has_more,
      });

    } catch (err: any) {
      console.error('[useGroupSearch] Ошибка поиска:', err);
      setError(err.message || 'Ошибка поиска групп');
    } finally {
      setIsLoading(false);
      setIsLoadingMore(false);
    }
  }, []);

  const loadMore = useCallback((
    searchQuery: string,
    categories: GroupCategory[],
    sortByParticipants: 'asc' | 'desc' | 'none',
    sortByRegistration: 'asc' | 'desc' | 'none'
  ) => {
    if (!isLoadingMore && hasMore) {
      searchGroups(
        searchQuery,
        categories,
        sortByParticipants,
        sortByRegistration,
        currentPage + 1,
        true
      );
    }
  }, [isLoadingMore, hasMore, currentPage, searchGroups]);

  const resetSearch = useCallback(() => {
    setGroups([]);
    setCurrentPage(1);
    setHasMore(false);
    setTotalGroups(0);
    setError(null);
  }, []);

  return {
    groups,
    isLoading,
    isLoadingMore,
    error,
    hasMore,
    totalGroups,
    currentPage,
    searchGroups,
    loadMore,
    resetSearch,
  };
};