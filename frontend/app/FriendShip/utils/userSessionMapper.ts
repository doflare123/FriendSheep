import { Event } from '@/components/event/EventCard';

interface UserGroupSession {
  id: number;
  title: string;
  session_type: string;
  session_place: string;
  start_time: string;
  duration: number;
  count_users_max: number;
  current_users: number;
  image_url: string;
  group_name: string;
  group_id: number;
  genres: string[];
  city: string;
}

const formatDate = (isoDate: string): string => {
  if (!isoDate) return 'Дата не указана';
  
  const date = new Date(isoDate);
  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const year = date.getFullYear();
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  
  return `${day}.${month}.${year} ${hours}:${minutes}`;
};

const mapSessionTypeToCategory = (sessionType: string): Event['category'] => {
  const type = sessionType.toLowerCase();
  
  if (type.includes('фильм') || type.includes('кино') || type.includes('медиа')) {
    return 'movie';
  }
  if (type.includes('игр') && !type.includes('настольн')) {
    return 'game';
  }
  if (type.includes('настольн')) {
    return 'table_game';
  }
  
  return 'other';
};

const mapSessionPlaceToType = (sessionPlace: string): Event['typePlace'] => {
  const place = sessionPlace.toLowerCase();
  
  if (place.includes('онлайн') || place.includes('online')) {
    return 'online';
  }
  
  return 'offline';
};

export const mapUserGroupSessionToEvent = (session: UserGroupSession): Event => {
  const category = mapSessionTypeToCategory(session.session_type);
  const typePlace = mapSessionPlaceToType(session.session_place);
  const eventPlace = typePlace === 'offline' ? (session.city?.trim() || '') : '';

  return {
    id: session.id.toString(),
    title: session.title || 'Без названия',
    date: formatDate(session.start_time),
    genres: session.genres || [],
    currentParticipants: session.current_users || 0,
    maxParticipants: session.count_users_max || 0,
    duration: `${session.duration || 0} мин`,
    imageUri: session.image_url || '',
    description: 'Описание отсутствует',
    typeEvent: session.session_type || 'Не указано',
    typePlace: typePlace,
    eventPlace: eventPlace,
    publisher: '',
    publicationDate: '',
    ageRating: '',
    category: category,
    group: session.group_name || 'Группа не указана',
  };
};