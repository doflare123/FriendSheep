import { Session } from '@/api/types/user';

export interface Event {
  id: string;
  title: string;
  date: string;
  genres: string[];
  currentParticipants: number;
  maxParticipants: number;
  duration: string;
  imageUri: string;
  description: string;
  typeEvent: string;
  typePlace: 'online' | 'offline';
  eventPlace: string;
  publisher: string;
  publicationDate: string;
  ageRating: string;
  category: 'movie' | 'game' | 'table_game' | 'other';
  group: string;
  onPress?: () => void;
  highlightedTitle?: {
    before: string;
    match: string;
    after: string;
  };
}

export interface StatisticsDataItem {
  name: string;
  percentage: number;
  color: string;
  legendFontColor: string;
}

export const sessionToEvent = (session: Session): Event => {
  const startDate = new Date(session.start_time);
  const endDate = new Date(session.end_time);

  const date = startDate.toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
  });

  const durationMs = endDate.getTime() - startDate.getTime();
  const hours = Math.floor(durationMs / (1000 * 60 * 60));
  const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60));
  const duration = hours > 0 ? `${hours}ч ${minutes}м` : `${minutes}м`;
  
  const categoryMap: Record<string, Event['category']> = {
    'movie': 'movie',
    'film': 'movie',
    'game': 'game',
    'video_game': 'game',
    'table_game': 'table_game',
    'board_game': 'table_game',
    'other': 'other',
  };
  
  const category = categoryMap[session.category_session?.toLowerCase()] || 
                   categoryMap[session.type_session?.toLowerCase()] || 
                   'other';
  
  return {
    id: session.id.toString(),
    title: session.title,
    date,
    genres: session.genres || [],
    currentParticipants: session.current_users,
    maxParticipants: session.max_users,
    duration,
    imageUri: session.image_url,
    description: '',
    typeEvent: session.type_session,
    typePlace: session.city ? 'offline' : 'online',
    eventPlace: session.city || 'Онлайн',
    publisher: '',
    publicationDate: session.start_time,
    ageRating: '',
    category,
    group: '',
  };
};

export const sessionsToEvents = (sessions?: Session[] | null): Event[] => {
  if (!sessions || sessions.length === 0) {
    return [];
  }
  return sessions.map(sessionToEvent);
};

export const generateStatisticsData = (): StatisticsDataItem[] => {
  return [
    { name: 'Фильмы', percentage: 30, color: '#FF6B6B', legendFontColor: '#000' },
    { name: 'Игры', percentage: 25, color: '#4ECDC4', legendFontColor: '#000' },
    { name: 'Настолки', percentage: 20, color: '#45B7D1', legendFontColor: '#000' },
    { name: 'Другое', percentage: 25, color: '#FFA07A', legendFontColor: '#000' },
  ];
};