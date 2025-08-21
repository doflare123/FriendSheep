import { useState } from 'react';

export interface SortingState {
  checkedCategories: string[];
  sortByDate: 'asc' | 'desc';
  sortByParticipants: 'asc' | 'desc';
  searchQuery: string;
}

export interface SortingActions {
  setCheckedCategories: (categories: string[]) => void;
  setSortByDate: (order: 'asc' | 'desc') => void;
  setSortByParticipants: (order: 'asc' | 'desc') => void;
  toggleCategoryCheckbox: (category: string) => void;
  setSearchQuery: (query: string) => void;
}

export const useSearchState = () => {
  const [checkedCategories, setCheckedCategories] = useState<string[]>(['Все']);
  const [sortByDate, setSortByDate] = useState<'asc' | 'desc'>('asc');
  const [sortByParticipants, setSortByParticipants] = useState<'asc' | 'desc'>('asc');
  const [searchQuery, setSearchQuery] = useState<string>('');

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
  };

  const sortingActions: SortingActions = {
    setCheckedCategories,
    setSortByDate,
    setSortByParticipants,
    toggleCategoryCheckbox,
    setSearchQuery,
  };

  return {
    sortingState,
    sortingActions
  };
};