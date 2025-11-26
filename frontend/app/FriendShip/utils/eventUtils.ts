import { Event } from '@/components/event/EventCard';

export const categoryMapping: Record<string, Event['category']> = {
  'Фильмы': 'movie',
  'Игры': 'game', 
  'Настолки': 'table_game',
  'Другое': 'other'
};

export const getEventsByCategory = (events: any[], category: string) => {
  return events.filter(event => event.category === category);
};

export const filterEventsBySearch = (events: any[], searchQuery: string) => {
  return events.filter(event =>
    event.title.toLowerCase().includes(searchQuery.toLowerCase())
  );
};

export const filterEventsByCategories = (events: any[], categories: string[]) => {
  if (categories.includes('Все')) {
    return events;
  }
  
  const mappedCategories = categories
    .map(category => categoryMapping[category])
    .filter(Boolean);
  
  return events.filter(event => mappedCategories.includes(event.category));
};

export const filterEventsByCity = (events: Event[], cityFilter: string): Event[] => {
  if (!cityFilter.trim()) {
    return events;
  }

  const normalizedFilter = cityFilter.toLowerCase().trim();

  return events.filter(event => {
    if (event.typePlace !== 'offline') {
      return false;
    }

    if (!event.eventPlace) {
      return false;
    }

    const normalizedCity = event.eventPlace.toLowerCase().trim();

    return normalizedCity.includes(normalizedFilter);
  });
};

export const sortEventsByParticipants = (events: any[], order: 'asc' | 'desc' | 'none') => {
  if (order === 'none') return events;
  
  return [...events].sort((a, b) => {
    if (order === 'asc') {
      return a.currentParticipants - b.currentParticipants;
    } else {
      return b.currentParticipants - a.currentParticipants;
    }
  });
};

export const createEventWithHighlightedTitle = (event: Event, query: string): Event => {
  if (!query.trim()) return event;

  const title = event.title;
  const lowerTitle = title.toLowerCase();
  const lowerQuery = query.toLowerCase();
  const index = lowerTitle.indexOf(lowerQuery);

  if (index === -1) return event;

  const beforeMatch = title.substring(0, index);
  const match = title.substring(index, index + query.length);
  const afterMatch = title.substring(index + query.length);

  return {
    ...event,
    highlightedTitle: {
      before: beforeMatch,
      match: match,
      after: afterMatch
    }
  };
};