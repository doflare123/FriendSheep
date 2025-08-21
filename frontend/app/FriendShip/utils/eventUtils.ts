import { Event } from '@/components/EventCard';

export const categoryMapping: Record<string, Event['category']> = {
  'Фильмы': 'movie',
  'Игры': 'game', 
  'Настолки': 'table_game',
  'Другое': 'other'
};

export const sortEventsByDate = (events: Event[], order: 'asc' | 'desc'): Event[] => {
  return [...events].sort((a, b) => {
    const dateA = new Date(a.date.split(' ')[0].split('.').reverse().join('-'));
    const dateB = new Date(b.date.split(' ')[0].split('.').reverse().join('-'));
    
    if (order === 'asc') {
      return dateA.getTime() - dateB.getTime();
    } else {
      return dateB.getTime() - dateA.getTime();
    }
  });
};

export const sortEventsByParticipants = (events: Event[], order: 'asc' | 'desc'): Event[] => {
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

export const getEventsByCategory = (events: Event[], category: Event['category']): Event[] => {
  return events.filter(event => event.category === category);
};

export const filterEventsBySearch = (events: Event[], query: string): Event[] => {
  if (!query.trim()) return events;
  return events.filter(event => 
    event.title.toLowerCase().includes(query.toLowerCase())
  );
};

export const filterEventsByCategories = (
  events: Event[], 
  checkedCategories: string[]
): Event[] => {
  if (checkedCategories.includes('Все')) return events;
  
  const mappedCategories = checkedCategories
    .map(cat => categoryMapping[cat])
    .filter(Boolean);
  
  return events.filter(event => mappedCategories.includes(event.category));
};