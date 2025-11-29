import { Event } from '@/components/event/EventCard';

interface BackendSession {
  id: number;
  title: string;
  session_type: string;
  session_place: string;
  start_time: string;
  end_time: string;
  duration: number;
  count_users_max: number;
  current_users: number;
  image_url: string;
  group_name?: string;
  genres?: string[];
  popularity_rate?: number;
  city?: string;
  location?: string;
  status?: string;
}

const formatDate = (isoDate: string): string => {
  if (!isoDate) return 'Дата не указана';
  
  const date = new Date(isoDate);
  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const year = date.getFullYear();
  
  return `${day}.${month}.${year}`;
};

const mapSessionTypeToCategory = (sessionType: string): Event['category'] => {
  const type = sessionType.toLowerCase();
  
  if (type.includes('movie') || type.includes('фильм') || type.includes('кино') || type.includes('медиа')) {
    return 'movie';
  }
  if (type.includes('game') || type.includes('игр') && !type.includes('настольн')) {
    return 'game';
  }
  if (type.includes('table') || type.includes('настольн')) {
    return 'table_game';
  }
  
  return 'other';
};

const mapSessionPlaceToType = (sessionPlace: string): Event['typePlace'] => {
  const place = sessionPlace.toLowerCase();
  
  if (place.includes('online') || place.includes('онлайн')) {
    return 'online';
  }
  
  return 'offline';
};

const getEventPlace = (
  session: BackendSession,
  metadata?: any
): string => {
  const typePlace = mapSessionPlaceToType(session.session_place);
  
  if (typePlace === 'offline') {
    const location = metadata?.Location?.trim() || 
                     session.location?.trim() || 
                     session.city?.trim();

    return location || '';
  }

  return '';
};

export const mapBackendSessionToEvent = (
  session: BackendSession,
  metadata?: any
): Event => {
  const category = mapSessionTypeToCategory(session.session_type);
  const typePlace = mapSessionPlaceToType(session.session_place);
  const eventPlace = getEventPlace(session, metadata);

  return {
    id: session.id.toString(),
    title: session.title || 'Без названия',
    date: formatDate(session.start_time),
    genres: session.genres || metadata?.Genres || [],
    currentParticipants: session.current_users || 0,
    maxParticipants: session.count_users_max || 0,
    duration: `${session.duration || 0} мин`,
    imageUri: session.image_url || '',
    description: metadata?.Notes || 'Описание отсутствует',
    typeEvent: session.session_type || 'Не указано',
    typePlace: typePlace,
    eventPlace: eventPlace,
    publisher: metadata?.Country || '',
    publicationDate: metadata?.Year ? metadata.Year.toString() : '',
    ageRating: metadata?.AgeLimit || '',
    category: category,
    group: session.group_name || 'Группа не указана',
  };
};

export const mapBackendSessionsToEvents = (
  sessions: BackendSession[]
): Event[] => {
  if (!sessions || sessions.length === 0) return [];
  
  return sessions.map(session => mapBackendSessionToEvent(session));
};