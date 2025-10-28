import { Event } from '@/components/event/EventCard';

export const parseDate = (dateString: string): Date => {
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

export const sortEventsByDate = (events: Event[], order: 'asc' | 'desc' | 'none'): Event[] => {
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

export const sortEventsByParticipants = (events: Event[], order: 'asc' | 'desc' | 'none'): Event[] => {
  if (order === 'none') return events;
  
  return [...events].sort((a, b) => {
    if (order === 'asc') {
      return a.currentParticipants - b.currentParticipants;
    } else {
      return b.currentParticipants - a.currentParticipants;
    }
  });
};