import userService from '@/api/services/userService';
import { SearchUserItem } from '@/api/types/user';
import { useCallback, useState } from 'react';

export const useUserSearch = () => {
  const [users, setUsers] = useState<SearchUserItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [totalUsers, setTotalUsers] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);

  const searchUsers = useCallback(async (query: string, page: number = 1, append: boolean = false) => {
    if (!query.trim()) {
      setUsers([]);
      setTotalUsers(0);
      setHasMore(false);
      return;
    }

    try {
      if (append) {
        setIsLoadingMore(true);
      } else {
        setIsLoading(true);
        setError(null);
      }

      const response = await userService.searchUsers(query, page);

      if (append) {
        setUsers(prev => [...prev, ...response.users]);
      } else {
        setUsers(response.users);
      }

      setTotalUsers(response.total);
      setHasMore(response.has_more);
      setCurrentPage(page);

      console.log('[useUserSearch] ✅ Пользователи загружены:', {
        page,
        total: response.total,
        loaded: response.users.length,
        hasMore: response.has_more,
      });
    } catch (err: any) {
      console.error('[useUserSearch] ❌ Ошибка поиска:', err);
      setError(err.message || 'Не удалось выполнить поиск');
    } finally {
      setIsLoading(false);
      setIsLoadingMore(false);
    }
  }, []);

  const loadMore = useCallback(
    async (query: string) => {
      if (!hasMore || isLoadingMore) return;
      await searchUsers(query, currentPage + 1, true);
    },
    [hasMore, isLoadingMore, currentPage, searchUsers]
  );

  const resetSearch = useCallback(() => {
    setUsers([]);
    setTotalUsers(0);
    setHasMore(false);
    setCurrentPage(1);
    setError(null);
  }, []);

  return {
    users,
    isLoading,
    isLoadingMore,
    error,
    hasMore,
    totalUsers,
    searchUsers,
    loadMore,
    resetSearch,
  };
};