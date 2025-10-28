import { useEffect, useState } from 'react';

export type GroupCategory = 'movie' | 'game' | 'table_game' | 'other';

export interface GroupSearchState {
  searchQuery: string;
  sortByParticipants: 'asc' | 'desc' | 'none';
  sortByRegistration: 'asc' | 'desc' | 'none';
  checkedCategories: GroupCategory[];
}

export interface GroupSearchActions {
  setSearchQuery: (query: string) => void;
  setSortByParticipants: (order: 'asc' | 'desc' | 'none') => void;
  setSortByRegistration: (order: 'asc' | 'desc' | 'none') => void;
  setCheckedCategories: (categories: GroupCategory[]) => void;
  toggleCategoryCheckbox: (category: GroupCategory) => void;
  resetFilters: () => void;
}

let globalGroupSearchState: GroupSearchState = {
  searchQuery: '',
  sortByParticipants: 'none',
  sortByRegistration: 'none',
  checkedCategories: [],
};

const listeners: Set<() => void> = new Set();

const notifyListeners = () => {
  listeners.forEach(listener => listener());
};

export const useGroupSearchState = () => {
  const [, forceUpdate] = useState({});

  useEffect(() => {
    const listener = () => forceUpdate({});
    listeners.add(listener);
    return () => {
      listeners.delete(listener);
    };
  }, []);

  const setSearchQuery = (query: string) => {
    globalGroupSearchState.searchQuery = query;
    notifyListeners();
  };

  const setSortByParticipants = (order: 'asc' | 'desc' | 'none') => {
    globalGroupSearchState.sortByParticipants = order;
    notifyListeners();
  };

  const setSortByRegistration = (order: 'asc' | 'desc' | 'none') => {
    globalGroupSearchState.sortByRegistration = order;
    notifyListeners();
  };

  const setCheckedCategories = (categories: GroupCategory[]) => {
    globalGroupSearchState.checkedCategories = categories;
    notifyListeners();
  };

  const toggleCategoryCheckbox = (category: GroupCategory) => {
    const currentCategories = [...globalGroupSearchState.checkedCategories];
    const index = currentCategories.indexOf(category);
    
    if (index > -1) {
      currentCategories.splice(index, 1);
    } else {
      currentCategories.push(category);
    }
    
    globalGroupSearchState.checkedCategories = currentCategories;
    notifyListeners();
  };

  const resetFilters = () => {
    globalGroupSearchState = {
      searchQuery: '',
      sortByParticipants: 'none',
      sortByRegistration: 'none',
      checkedCategories: [],
    };
    notifyListeners();
  };

  return {
    searchState: globalGroupSearchState,
    searchActions: {
      setSearchQuery,
      setSortByParticipants,
      setSortByRegistration,
      setCheckedCategories,
      toggleCategoryCheckbox,
      resetFilters,
    },
  };
};