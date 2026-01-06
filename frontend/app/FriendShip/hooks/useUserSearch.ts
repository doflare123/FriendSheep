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
    try {
      if (append) {
        setIsLoadingMore(true);
      } else {
        setIsLoading(true);
        setError(null);
      }

      const searchQuery = query.trim();
      const isEmptySearch = !searchQuery || searchQuery === '@';
      
      console.log('[useUserSearch] 🔍 Запрос:', {
        query: searchQuery,
        isEmptySearch,
        page,
        append,
      });
      
      const response = isEmptySearch 
        ? await userService.getAllUsers(page)
        : await userService.searchUsers(searchQuery, page);

      console.log('[useUserSearch] 📦 Ответ:', {
        users: response.users?.length,
        total: response.total,
        hasMore: response.has_more,
      });

      if (append) {
        setUsers(prev => {
          const newUsers = [...prev, ...(response.users || [])];
          console.log('[useUserSearch] ➕ Добавлено пользователей:', response.users?.length, 'Всего:', newUsers.length);
          return newUsers;
        });
      } else {
        setUsers(response.users || []);
        console.log('[useUserSearch] 🔄 Заменено пользователей:', response.users?.length);
      }

      setTotalUsers(response.total || 0);
      setHasMore(response.has_more || false);
      setCurrentPage(page);

      console.log('[useUserSearch] ✅ Состояние обновлено:', {
        totalUsers: response.total,
        hasMore: response.has_more,
        currentPage: page,
      });
    } catch (err: any) {
      console.error('[useUserSearch] ❌ Ошибка поиска:', err);
      setError(err.message || 'Не удалось выполнить поиск');
      if (!append) {
        setUsers([]);
        setTotalUsers(0);
        setHasMore(false);
      }
    } finally {
      setIsLoading(false);
      setIsLoadingMore(false);
    }
  }, []);

  const loadMore = useCallback(
    async (query: string) => {
      if (!hasMore || isLoadingMore) {
        console.log('[useUserSearch] ⚠️ Не можем загрузить больше:', { hasMore, isLoadingMore });
        return;
      }
      
      console.log('[useUserSearch] 📄 Загрузка страницы', currentPage + 1);
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