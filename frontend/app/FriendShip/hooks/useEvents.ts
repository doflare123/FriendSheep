import { eventsData } from '@/data/eventsData';
import { SortingState } from '@/hooks/useSearchState';
import {
  createEventWithHighlightedTitle,
  filterEventsByCategories,
  filterEventsBySearch,
  getEventsByCategory,
  sortEventsByParticipants
} from '@/utils/eventUtils';
import { useMemo } from 'react';

const parseDate = (dateString: string): Date => {
  const parts = dateString.split(' ');
  const datePart = parts[0];
  const timePart = parts[1] || '00:00';
  
  const [day, month, year] = datePart.split('.');
  const [hour, minute] = timePart.split(':');
  
  return new Date(
    parseInt(year), 
    parseInt(month) - 1,
    parseInt(day), 
    parseInt(hour) || 0, 
    parseInt(minute) || 0
  );
};

const sortEventsByDate = (events: any[], order: 'asc' | 'desc' | 'none') => {
  if (order === 'none') return events;
  
  return [...events].sort((a, b) => {
    const dateA = parseDate(a.date);
    const dateB = parseDate(b.date);
    
    if (order === 'asc') {
      return dateA.getTime() - dateB.getTime();
    } else {
      return dateB.getTime() - dateA.getTime();
    }
  });
};

export const useEvents = (sortingState: SortingState) => {
  const { checkedCategories, sortByDate, sortByParticipants, searchQuery } = sortingState;

  const movieEvents = useMemo(() => {
    let events = getEventsByCategory(eventsData, 'movie');
    events = sortEventsByDate(events, sortByDate);
    events = sortEventsByParticipants(events, sortByParticipants);
    return events;
  }, [sortByDate, sortByParticipants]);

  const gameEvents = useMemo(() => {
    let events = getEventsByCategory(eventsData, 'game');
    events = sortEventsByDate(events, sortByDate);
    events = sortEventsByParticipants(events, sortByParticipants);
    return events;
  }, [sortByDate, sortByParticipants]);

  const tableGameEvents = useMemo(() => {
    let events = getEventsByCategory(eventsData, 'table_game');
    events = sortEventsByDate(events, sortByDate);
    events = sortEventsByParticipants(events, sortByParticipants);
    return events;
  }, [sortByDate, sortByParticipants]);

  const otherEvents = useMemo(() => {
    let events = getEventsByCategory(eventsData, 'other');
    events = sortEventsByDate(events, sortByDate);
    events = sortEventsByParticipants(events, sortByParticipants);
    return events;
  }, [sortByDate, sortByParticipants]);

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