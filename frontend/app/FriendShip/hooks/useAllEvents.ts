import groupService from '@/api/services/group/groupService';
import userService from '@/api/services/userService';
import { SessionInfo } from '@/api/types/user';
import { Event } from '@/components/event/EventCard';
import { SortingState } from '@/hooks/useSearchState';
import { groupSessionsToEvents } from '@/utils/dataAdapters';
import {
  createEventWithHighlightedTitle,
  filterEventsByCategories,
  filterEventsByCity,
  filterEventsBySearch,
  sortEventsByParticipants
} from '@/utils/eventUtils';
import { filterActiveSessions, getSessionStatus } from '@/utils/sessionStatusHelpers';

import { sortEventsByDate } from '@/utils/eventSorting';
import { useEffect, useMemo, useState } from 'react';

type EventsMode = 'user' | 'group';
type UserEventFilter = 'current' | 'completed';

const mapSessionInfoToEvent = (session: SessionInfo, isFromRecent?: boolean): Event => {
  const formatDate = (isoDate: string): string => {
    if (!isoDate) return 'Дата не указана';
    
    if (isoDate.includes('.') && isoDate.includes(' ')) {
      return isoDate.split(' ')[0];
    }
    
    if (isoDate.includes('T') || isoDate.includes('Z')) {
      const date = new Date(isoDate);
      const day = String(date.getDate()).padStart(2, '0');
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const year = date.getFullYear();
      
      return `${day}.${month}.${year}`;
    }
    
    return isoDate;
  };

  const mapCategoryToEventCategory = (category: string): Event['category'] => {
    const cat = category.toLowerCase();
    if (cat.includes('фильм') || cat.includes('кино') || cat.includes('медиа')) return 'movie';
    if (cat.includes('игр') && !cat.includes('настольн')) return 'game';
    if (cat.includes('настольн')) return 'table_game';
    return 'other';
  };

  const mapTypeToPlace = (type: string): Event['typePlace'] => {
    const t = type.toLowerCase();
    if (t.includes('онлайн') || t.includes('online')) return 'online';
    return 'offline';
  };

  const category = mapCategoryToEventCategory(session.category_session);
  const typePlace = mapTypeToPlace(session.type_session);
  
  const start = new Date(session.start_time);
  const end = new Date(session.end_time);
  const durationMinutes = Math.round((end.getTime() - start.getTime()) / (1000 * 60));

  return {
    id: session.id.toString(),
    title: session.title || 'Без названия',
    date: formatDate(session.start_time),
    genres: session.genres || [],
    currentParticipants: session.current_users || 0,
    maxParticipants: session.max_users || 0,
    duration: `${durationMinutes} мин`,
    imageUri: session.image_url || '',
    description: 'Описание отсутствует',
    typeEvent: session.category_session || 'Не указано',
    typePlace: typePlace,
    eventPlace: typePlace === 'offline' ? (session.city?.trim() || session.location?.trim() || '') : '',
    publisher: '',
    publicationDate: '',
    ageRating: '',
    category: category,
    group: 'Группа не указана',
    _isFromRecent: isFromRecent,
  } as Event & { _isFromRecent?: boolean };
};

function filterUserEventsByStatus(events: (Event & { _isFromRecent?: boolean })[], filter: UserEventFilter): Event[] {
  console.log(`[filterUserEventsByStatus] 🔍 Фильтруем события. Фильтр: ${filter}`);
  console.log(`[filterUserEventsByStatus] 📊 Всего событий на входе: ${events.length}`);
  
  const eventsWithMark = events.filter(e => e._isFromRecent !== undefined).length;
  const eventsFromRecent = events.filter(e => e._isFromRecent === true).length;
  const eventsFromUpcoming = events.filter(e => e._isFromRecent === false).length;
  
  console.log(`[filterUserEventsByStatus] 📊 Событий с меткой: ${eventsWithMark}`);
  console.log(`[filterUserEventsByStatus] 📊 Из recent_sessions: ${eventsFromRecent}`);
  console.log(`[filterUserEventsByStatus] 📊 Из upcoming_sessions: ${eventsFromUpcoming}`);
  
  const filtered = events.filter(event => {
    if (event._isFromRecent !== undefined) {
      if (filter === 'completed') {
        const result = event._isFromRecent === true;
        console.log(`[filterUserEventsByStatus] 📝 "${event.title}" - isFromRecent: ${event._isFromRecent}, показываем: ${result}`);
        return result;
      } else {
        const result = event._isFromRecent === false;
        console.log(`[filterUserEventsByStatus] 📝 "${event.title}" - isFromRecent: ${event._isFromRecent}, показываем: ${result}`);
        return result;
      }
    }

    const durationMatch = event.duration.match(/\d+/);
    const duration = durationMatch ? parseInt(durationMatch[0]) : 0;
    const status = getSessionStatus(event.date, duration);
    
    console.log(`[filterUserEventsByStatus] 📝 "${event.title}" - статус по времени: ${status}`);
    
    if (filter === 'current') {
      return status === 'recruitment' || status === 'in_progress';
    } else {
      return status === 'completed';
    }
  });
  
  console.log(`[filterUserEventsByStatus] ✅ После фильтрации: ${filtered.length} событий`);
  return filtered;
}

export const useAllEvents = (
  mode: EventsMode,
  sortingState: SortingState,
  userEventFilter?: UserEventFilter,
  groupId?: string,
  userId?: string
) => {
  const { checkedCategories, sortByDate, sortByParticipants, searchQuery, cityFilter } = sortingState;
  
  const [allEvents, setAllEvents] = useState<Event[]>([]);
  const [userName, setUserName] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadEvents = async () => {
    try {
      setIsLoading(true);
      setError(null);

      if (mode === 'user') {
        console.log('[useAllEvents] 🚀 Загрузка событий пользователя...');

        let profile;
        
        if (userId) {
          console.log('[useAllEvents] 📋 Загрузка событий для userId:', userId);
          profile = await userService.getUserProfileById(userId);
          setUserName(profile.name);
        } else {
          console.log('[useAllEvents] 📋 Загрузка событий текущего пользователя');
          profile = await userService.getCurrentUserProfile();
          setUserName('');
        }
        
        const upcomingSessions = profile.upcoming_sessions || [];
        const recentSessions = profile.recent_sessions || [];
        
        console.log('[useAllEvents] 📊 Предстоящие сессии:', upcomingSessions.length);
        console.log('[useAllEvents] 📊 Завершённые сессии:', recentSessions.length);

        const upcomingEvents = upcomingSessions.map(session => mapSessionInfoToEvent(session, false));
        const recentEvents = recentSessions.map(session => mapSessionInfoToEvent(session, true));
        
        const allSessions = [...upcomingEvents, ...recentEvents];
        
        setAllEvents(allSessions);
        console.log('[useAllEvents] ✅ События пользователя загружены:', allSessions.length);
        console.log('[useAllEvents] 🔍 Из них предстоящих:', upcomingEvents.length);
        console.log('[useAllEvents] 🔍 Из них завершённых:', recentEvents.length);

      } else if (mode === 'group') {
        if (!groupId) {
          throw new Error('groupId обязателен для режима group');
        }

        console.log('[useAllEvents] 🚀 Загрузка событий группы:', groupId);

        const groupData = await groupService.getPublicGroupDetail(groupId);
        
        console.log('[useAllEvents] 📊 Получено сессий:', groupData.sessions?.length || 0);
        
        const activeSessions = groupData.sessions 
          ? filterActiveSessions(groupData.sessions)
          : [];
        
        console.log('[useAllEvents] 📊 Активных сессий:', activeSessions.length);
        
        const mappedEvents = groupSessionsToEvents(
          { ...groupData, sessions: activeSessions }
        );
        
        setAllEvents(mappedEvents);
        console.log('[useAllEvents] ✅ События группы загружены:', mappedEvents.length);
      }

    } catch (error: any) {
      console.error('[useAllEvents] ❌ Ошибка загрузки событий:', error);
      setError(error.message || 'Не удалось загрузить события');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadEvents();
  }, [mode, groupId, userId]);

  const filteredAndSortedEvents = useMemo(() => {
    let events = [...allEvents];

    if (mode === 'user' && userEventFilter) {
      events = filterUserEventsByStatus(events, userEventFilter);
      console.log(`[useAllEvents] После фильтра по статусу (${userEventFilter}):`, events.length);
    }

    if (searchQuery.trim()) {
      events = filterEventsBySearch(events, searchQuery);
      console.log('[useAllEvents] После поиска:', events.length);
    }

    events = filterEventsByCategories(events, checkedCategories);
    console.log('[useAllEvents] После фильтра по категориям:', events.length);

    events = filterEventsByCity(events, cityFilter);
    console.log('[useAllEvents] После фильтра по городу:', events.length);

    if (sortByDate !== 'none') {
      events = sortEventsByDate(events, sortByDate);
      console.log('[useAllEvents] После сортировки по дате:', events.map(e => e.date));
    }

    if (sortByParticipants !== 'none') {
      events = sortEventsByParticipants(events, sortByParticipants);
      console.log('[useAllEvents] После сортировки по участникам:', events.map(e => e.currentParticipants));
    }

    if (searchQuery.trim()) {
      events = events.map(event => createEventWithHighlightedTitle(event, searchQuery));
    }
    
    return events;
  }, [
    allEvents,
    mode,
    userEventFilter,
    searchQuery,
    checkedCategories, 
    sortByDate, 
    sortByParticipants, 
    cityFilter
  ]);

  return {
    events: filteredAndSortedEvents,
    allEvents,
    userName,
    isLoading,
    error,
    refreshEvents: loadEvents,
  };
};