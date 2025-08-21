import { eventsData } from '@/data/eventsData';
import { SortingState } from '@/hooks/useSearchState';
import {
    createEventWithHighlightedTitle,
    filterEventsByCategories,
    filterEventsBySearch,
    getEventsByCategory,
    sortEventsByDate,
    sortEventsByParticipants
} from '@/utils/eventUtils';
import { useMemo } from 'react';

export const useEvents = (sortingState: SortingState) => {
  const { checkedCategories, sortByDate, sortByParticipants, searchQuery } = sortingState;

  const movieEvents = useMemo(() => {
    return getEventsByCategory(eventsData, 'movie');
  }, []);

  const gameEvents = useMemo(() => {
    return getEventsByCategory(eventsData, 'game');
  }, []);

  const tableGameEvents = useMemo(() => {
    return getEventsByCategory(eventsData, 'table_game');
  }, []);

  const otherEvents = useMemo(() => {
    return getEventsByCategory(eventsData, 'other');
  }, []);

  const popularEvents = useMemo(() => {
    return [...eventsData].sort((a, b) => b.currentParticipants - a.currentParticipants);
  }, []);

  const newEvents = useMemo(() => {
    return sortEventsByDate(eventsData, 'desc');
  }, []);

  const searchResults = useMemo(() => {
    if (!searchQuery.trim()) return [];

    let filtered = filterEventsBySearch(eventsData, searchQuery);

    filtered = filterEventsByCategories(filtered, checkedCategories);

    filtered = sortEventsByDate(filtered, sortByDate);
    filtered = sortEventsByParticipants(filtered, sortByParticipants);

    return filtered.map(event => createEventWithHighlightedTitle(event, searchQuery));
  }, [searchQuery, checkedCategories, sortByDate, sortByParticipants]);

  return {
    movieEvents,
    gameEvents,
    tableGameEvents,
    otherEvents,
    popularEvents,
    newEvents,
    searchResults
  };
};