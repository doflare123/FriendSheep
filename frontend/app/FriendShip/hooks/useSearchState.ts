import { useState } from 'react';

export type SearchType = 'event' | 'profile' | 'group';

export interface SortingState {
  checkedCategories: string[];
  sortByDate: 'asc' | 'desc' | 'none';
  sortByParticipants: 'asc' | 'desc' | 'none';
  searchQuery: string;
  activeSearchType: SearchType;
  cityFilter: string;
}

export interface SortingActions {
  setCheckedCategories: (categories: string[]) => void;
  setSortByDate: (order: 'asc' | 'desc' | 'none') => void;
  setSortByParticipants: (order: 'asc' | 'desc' | 'none') => void;
  toggleCategoryCheckbox: (category: string) => void;
  setSearchQuery: (query: string) => void;
  setActiveSearchType: (type: SearchType) => void;
  setCityFilter: (city: string) => void;
}

let globalActiveSearchType: SearchType = 'event';
const searchTypeListeners: Set<(type: SearchType) => void> = new Set();

const notifySearchTypeListeners = (type: SearchType) => {
  searchTypeListeners.forEach(listener => listener(type));
};

export const useSearchState = () => {
  const [checkedCategories, setCheckedCategories] = useState<string[]>(['Все']);
  const [sortByDate, setSortByDate] = useState<'asc' | 'desc' | 'none'>('none');
  const [sortByParticipants, setSortByParticipants] = useState<'asc' | 'desc' | 'none'>('none');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [activeSearchType, setActiveSearchTypeLocal] = useState<SearchType>(globalActiveSearchType);
  const [cityFilter, setCityFilter] = useState<string>('');

  useState(() => {
    const listener = (type: SearchType) => {
      setActiveSearchTypeLocal(type);
    };
    searchTypeListeners.add(listener);
    return () => {
      searchTypeListeners.delete(listener);
    };
  });

  const setActiveSearchType = (type: SearchType) => {
    globalActiveSearchType = type;
    setActiveSearchTypeLocal(type);
    notifySearchTypeListeners(type);
  };

  const toggleCategoryCheckbox = (category: string) => {
    if (category === 'Все') {
      setCheckedCategories(['Все']);
    } else {
      const filtered = checkedCategories.filter(c => c !== 'Все');
      if (checkedCategories.includes(category)) {
        const newCategories = filtered.filter(c => c !== category);
        setCheckedCategories(newCategories.length === 0 ? ['Все'] : newCategories);
      } else {
        setCheckedCategories([...filtered, category]);
      }
    }
  };

  const sortingState: SortingState = {
    checkedCategories,
    sortByDate,
    sortByParticipants,
    searchQuery,
    activeSearchType,
    cityFilter,
  };

  const sortingActions: SortingActions = {
    setCheckedCategories,
    setSortByDate,
    setSortByParticipants,
    toggleCategoryCheckbox,
    setSearchQuery,
    setActiveSearchType,
    setCityFilter,
  };

  return {
    sortingState,
    sortingActions
  };
};