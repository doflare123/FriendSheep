import { useEffect, useState } from 'react';

export interface UserSearchState {
  searchQuery: string;
}

export interface UserSearchActions {
  setSearchQuery: (query: string) => void;
}

let globalUserSearchQuery: string = '';
const userSearchListeners: Set<(query: string) => void> = new Set();

const notifyUserSearchListeners = (query: string) => {
  userSearchListeners.forEach(listener => listener(query));
};

export const useUserSearchState = () => {
  const [searchQuery, setSearchQueryLocal] = useState<string>(globalUserSearchQuery);

  useEffect(() => {
    const listener = (query: string) => {
      setSearchQueryLocal(query);
    };
    userSearchListeners.add(listener);
    return () => {
      userSearchListeners.delete(listener);
    };
  }, []);

  const setSearchQuery = (query: string) => {
    globalUserSearchQuery = query;
    setSearchQueryLocal(query);
    notifyUserSearchListeners(query);
  };

  const searchState: UserSearchState = {
    searchQuery,
  };

  const searchActions: UserSearchActions = {
    setSearchQuery,
  };

  return {
    searchState,
    searchActions
  };
};