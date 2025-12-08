import sessionService from '@/api/services/session/sessionService';
import { Event } from '@/components/event/EventCard';
import { SortingState } from '@/hooks/useSearchState';
import {
  createEventWithHighlightedTitle,
  filterEventsByCategories,
  filterEventsByCity,
  filterEventsBySearch,
  sortEventsByParticipants
} from '@/utils/eventUtils';
import { mapBackendSessionsToEvents } from '@/utils/sessionMapper';
import { getSessionStatus } from '@/utils/sessionStatusHelpers';
import { useEffect, useMemo, useState } from 'react';

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

const sortEventsByDate = (events: Event[], order: 'asc' | 'desc' | 'none') => {
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

const getEventsByCategory = (events: Event[], category: Event['category']) => {
  return events.filter(event => event.category === category);
};

function filterActiveEvents(events: Event[]): Event[] {
  return events.filter(event => {
    const durationMatch = event.duration.match(/\d+/);
    const duration = durationMatch ? parseInt(durationMatch[0]) : 0;

    const status = getSessionStatus(event.date, duration);
    
    const isActive = status === 'recruitment' || status === 'in_progress';
    
    if (!isActive) {
      console.log('[useEvents] 🚫 Событие завершено, фильтруем:', event.title);
    }
    
    return isActive;
  });
}

export const useEvents = (sortingState: SortingState) => {
  const { checkedCategories, sortByDate, sortByParticipants, searchQuery, cityFilter } = sortingState;
  
  const [allEvents, setAllEvents] = useState<Event[]>([]);
  const [popularEventsData, setPopularEventsData] = useState<Event[]>([]);
  const [newEventsData, setNewEventsData] = useState<Event[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadSessions = async () => {
    try {
      setIsLoading(true);
      setError(null);

      console.log('[useEvents] 🚀 Загрузка всех событий...');

      const popularResponse = await sessionService.getPopularSessions();
      const popularSessions = popularResponse?.sessions || [];
      
      console.log('[useEvents] 📊 Популярные сессии от API:', 
        popularSessions.map((s: any) => ({ 
          id: s.id, 
          title: s.title, 
          current_users: s.current_users 
        }))
      );
      
      const mappedPopular = mapBackendSessionsToEvents(popularSessions);
      const filteredPopular = filterActiveEvents(mappedPopular);
      
      console.log('[useEvents] 🔍 Популярные события после маппинга:', 
        filteredPopular.map(e => ({ 
          id: e.id, 
          title: e.title, 
          participants: e.currentParticipants 
        }))
      );

      console.log('[useEvents] ✅ Популярные события загружены:', filteredPopular.length);

      try {
        const newResponse = await sessionService.getNewSessions();
        const newSessions = newResponse?.sessions || [];
        const mappedNew = mapBackendSessionsToEvents(newSessions);

        const filteredNew = filterActiveEvents(mappedNew);
        setNewEventsData(filteredNew);
        console.log('[useEvents] ✅ Новые события загружены:', filteredNew.length);
      } catch (newError) {
        console.log('[useEvents] ⚠️ Эндпоинт для новых событий недоступен');
        setNewEventsData([]);
      }

      try {
        const allResponse = await sessionService.getAllSessions();
        const allSessions = allResponse?.sessions || [];
        const mappedAll = mapBackendSessionsToEvents(allSessions);

        const filteredAll = filterActiveEvents(mappedAll);
        setAllEvents(filteredAll);

        const updatedPopular = filteredPopular.map(popularEvent => {
          const freshEvent = filteredAll.find(e => e.id === popularEvent.id);
          if (freshEvent) {
            console.log(`[useEvents] 🔄 Обновляем популярное событие ${popularEvent.title}: ${popularEvent.currentParticipants} → ${freshEvent.currentParticipants}`);
            return freshEvent;
          }
          return popularEvent;
        });
        
        setPopularEventsData(updatedPopular);
        
        console.log('[useEvents] ✅ Все события загружены:', filteredAll.length);
      } catch (allError) {
        console.log('[useEvents] ⚠️ Эндпоинт для всех событий недоступен');
        setAllEvents([]);
        setPopularEventsData(filteredPopular);
      }

    } catch (error: any) {
      console.error('[useEvents] ❌ Ошибка загрузки событий:', error);
      setError(error.message || 'Не удалось загрузить события');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadSessions();
  }, []);

  const movieEvents = useMemo(() => {
    let events = getEventsByCategory(allEvents, 'movie');
    events = filterEventsByCity(events, cityFilter);
    events = sortEventsByDate(events, sortByDate);
    events = sortEventsByParticipants(events, sortByParticipants);
    return events;
  }, [allEvents, sortByDate, sortByParticipants, cityFilter]);

  const gameEvents = useMemo(() => {
    let events = getEventsByCategory(allEvents, 'game');
    events = filterEventsByCity(events, cityFilter);
    events = sortEventsByDate(events, sortByDate);
    events = sortEventsByParticipants(events, sortByParticipants);
    return events;
  }, [allEvents, sortByDate, sortByParticipants, cityFilter]);

  const tableGameEvents = useMemo(() => {
    let events = getEventsByCategory(allEvents, 'table_game');
    events = filterEventsByCity(events, cityFilter);
    events = sortEventsByDate(events, sortByDate);
    events = sortEventsByParticipants(events, sortByParticipants);
    return events;
  }, [allEvents, sortByDate, sortByParticipants, cityFilter]);

  const otherEvents = useMemo(() => {
    let events = getEventsByCategory(allEvents, 'other');
    events = filterEventsByCity(events, cityFilter);
    events = sortEventsByDate(events, sortByDate);
    events = sortEventsByParticipants(events, sortByParticipants);
    return events;
  }, [allEvents, sortByDate, sortByParticipants, cityFilter]); 

  const popularEvents = useMemo(() => {
    return filterEventsByCity(popularEventsData, cityFilter);
  }, [popularEventsData, cityFilter]);

  const newEvents = useMemo(() => {
    return filterEventsByCity(newEventsData, cityFilter);
  }, [newEventsData, cityFilter]);

  const searchResults = useMemo(() => {
    if (!searchQuery.trim()) return [];

    let filtered = filterEventsBySearch(allEvents, searchQuery);
    filtered = filterEventsByCategories(filtered, checkedCategories);
    filtered = filterEventsByCity(filtered, cityFilter);
    filtered = sortEventsByDate(filtered, sortByDate);
    filtered = sortEventsByParticipants(filtered, sortByParticipants);

    return filtered.map(event => createEventWithHighlightedTitle(event, searchQuery));
  }, [searchQuery, allEvents, checkedCategories, sortByDate, sortByParticipants, cityFilter]);

  return {
    movieEvents,
    gameEvents,
    tableGameEvents,
    otherEvents,
    popularEvents,
    newEvents,
    searchResults,
    isLoading,
    error,
    refreshEvents: loadSessions,
  };
};