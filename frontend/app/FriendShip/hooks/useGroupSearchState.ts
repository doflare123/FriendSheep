import { useState } from 'react';

export interface GroupSearchState {
  searchQuery: string;
  sortByParticipants: 'asc' | 'desc' | 'none';
  sortByRegistration: 'asc' | 'desc' | 'none';
  checkedCategories: string[];
}

export interface GroupSearchActions {
  setSearchQuery: (query: string) => void;
  setSortByParticipants: (order: 'asc' | 'desc' | 'none') => void;
  setSortByRegistration: (order: 'asc' | 'desc' | 'none') => void;
  setCheckedCategories: (categories: string[]) => void;
  toggleCategoryCheckbox: (category: string) => void;
  resetFilters: () => void;
}

let globalGroupSearchState: GroupSearchState = {
  searchQuery: '',
  sortByParticipants: 'none',
  sortByRegistration: 'none',
  checkedCategories: ['Все'],
};

const listeners: Set<() => void> = new Set();

const notifyListeners = () => {
  listeners.forEach(listener => listener());
};

export const useGroupSearchState = () => {
  const [, forceUpdate] = useState({});

  const subscribe = (listener: () => void) => {
    listeners.add(listener);
    return () => listeners.delete(listener);
  };

  useState(() => {
    const unsubscribe = subscribe(() => forceUpdate({}));
    return unsubscribe;
  });

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

  const setCheckedCategories = (categories: string[]) => {
    globalGroupSearchState.checkedCategories = categories;
    notifyListeners();
  };

  const toggleCategoryCheckbox = (category: string) => {
    const currentCategories = [...globalGroupSearchState.checkedCategories];
    
    if (category === 'Все') {
      globalGroupSearchState.checkedCategories = ['Все'];
    } else {
      const index = currentCategories.indexOf(category);
      if (index > -1) {
        currentCategories.splice(index, 1);
        const allIndex = currentCategories.indexOf('Все');
        if (allIndex > -1) {
          currentCategories.splice(allIndex, 1);
        }
      } else {
        currentCategories.push(category);
        const allIndex = currentCategories.indexOf('Все');
        if (allIndex > -1) {
          currentCategories.splice(allIndex, 1);
        }
      }
      
      globalGroupSearchState.checkedCategories = currentCategories.length > 0 ? currentCategories : ['Все'];
    }
    
    notifyListeners();
  };

  const resetFilters = () => {
    globalGroupSearchState = {
      searchQuery: '',
      sortByParticipants: 'none',
      sortByRegistration: 'none',
      checkedCategories: ['Все'],
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