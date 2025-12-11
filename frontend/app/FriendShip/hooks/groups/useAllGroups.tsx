import groupService from '@/api/services/group/groupService';
import userService from '@/api/services/userService';
import { Group } from '@/components/groups/GroupCard';
import { GroupCategory, GroupSearchState } from '@/hooks/useGroupSearchState';
import { createGroupWithHighlightedName } from '@/utils/groupUtils';
import { useEffect, useMemo, useState } from 'react';

type GroupsMode = 'managed' | 'subscriptions';

const CATEGORY_MAPPING: Record<string, string> = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};

const transformToGroup = (data: any, onPress: (id: string) => void): Group => {
  const mappedCategories = (data.category || [])
    .map((cat: string) => CATEGORY_MAPPING[cat])
    .filter((cat: string | undefined) => cat !== undefined);

  return {
    id: data.id.toString(),
    name: data.name,
    participantsCount: data.member_count,
    description: data.small_description,
    imageUri: data.image,
    categories: mappedCategories as any[],
    isPrivate: data.type === 'private',
    onPress: () => onPress(data.id.toString()),
  };
};

function filterGroupsBySearch(groups: Group[], query: string): Group[] {
  if (!query.trim()) return groups;
  
  const lowerQuery = query.toLowerCase();
  return groups.filter(group => 
    group.name.toLowerCase().includes(lowerQuery)
  );
}

function filterGroupsByCategories(groups: Group[], categories: GroupCategory[]): Group[] {
  if (categories.length === 0) {
    return groups;
  }

  return groups.filter(group =>
    group.categories.some(cat => categories.includes(cat as GroupCategory))
  );
}

function sortGroupsByParticipants(groups: Group[], order: 'asc' | 'desc' | 'none'): Group[] {
  if (order === 'none') return groups;

  return [...groups].sort((a, b) => {
    const diff = a.participantsCount - b.participantsCount;
    return order === 'asc' ? diff : -diff;
  });
}

function sortGroupsByDate(groups: Group[], order: 'asc' | 'desc' | 'none'): Group[] {
  if (order === 'none') return groups;

  return [...groups].sort((a, b) => {
    const diff = parseInt(a.id) - parseInt(b.id);
    return order === 'asc' ? diff : -diff;
  });
}

export const useAllGroups = (
  mode: GroupsMode,
  searchState: GroupSearchState,
  onGroupPress: (groupId: string) => void
) => {
  const { checkedCategories, sortByParticipants, sortByRegistration, searchQuery } = searchState;
  
  const [allGroups, setAllGroups] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadGroups = async () => {
    try {
      setIsLoading(true);
      setError(null);

      console.log(`[useAllGroups] 🚀 Загрузка групп. Режим: ${mode}`);

      let rawGroups: any[] = [];

      if (mode === 'managed') {
        rawGroups = await groupService.getAdminGroups();
        console.log('[useAllGroups] 📊 Управляемых групп:', rawGroups.length);
      } else {
        rawGroups = await userService.getUserSubscriptions();
        console.log('[useAllGroups] 📊 Подписок:', rawGroups.length);
      }

      const transformedGroups = rawGroups.map(group => 
        transformToGroup(group, onGroupPress)
      );

      setAllGroups(transformedGroups);
      console.log('[useAllGroups] ✅ Группы загружены:', transformedGroups.length);

    } catch (error: any) {
      console.error('[useAllGroups] ❌ Ошибка загрузки групп:', error);
      setError(error.message || 'Не удалось загрузить группы');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadGroups();
  }, [mode]);

  const filteredAndSortedGroups = useMemo(() => {
    let groups = [...allGroups];

    console.log('[useAllGroups] 🔍 Начальное количество групп:', groups.length);

    if (searchQuery.trim()) {
      groups = filterGroupsBySearch(groups, searchQuery);
      console.log('[useAllGroups] После поиска:', groups.length);
    }

    groups = filterGroupsByCategories(groups, checkedCategories);
    console.log('[useAllGroups] После фильтра по категориям:', groups.length);

    if (sortByRegistration !== 'none') {
      groups = sortGroupsByDate(groups, sortByRegistration);
      console.log('[useAllGroups] После сортировки по дате регистрации');
    }

    if (sortByParticipants !== 'none') {
      groups = sortGroupsByParticipants(groups, sortByParticipants);
      console.log('[useAllGroups] После сортировки по участникам');
    }

    if (searchQuery.trim()) {
      groups = groups.map(group => createGroupWithHighlightedName(group, searchQuery));
    }
    
    return groups;
  }, [
    allGroups,
    searchQuery,
    checkedCategories, 
    sortByRegistration, 
    sortByParticipants
  ]);

  return {
    groups: filteredAndSortedGroups,
    allGroups,
    isLoading,
    error,
    refreshGroups: loadGroups,
  };
};