import { PublicGroupResponse } from '@/api/services/groupService';
import { Session } from '@/api/types/user';
import { normalizeImageUrl } from '@/utils/imageUtils';

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
  onSessionUpdate?: () => void;
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

  const day = String(startDate.getDate()).padStart(2, '0');
  const month = String(startDate.getMonth() + 1).padStart(2, '0');
  const year = startDate.getFullYear();
  const date = `${day}.${month}.${year}`;

  const durationMs = endDate.getTime() - startDate.getTime();
  const totalMinutes = Math.floor(durationMs / (1000 * 60));
  const duration = `${totalMinutes} мин`;
  
  const categoryMap: Record<string, Event['category']> = {
    'movie': 'movie',
    'film': 'movie',
    'фильмы': 'movie',
    'фильм': 'movie',
    'game': 'game',
    'video_game': 'game',
    'игры': 'game',
    'игра': 'game',
    'table_game': 'table_game',
    'board_game': 'table_game',
    'настольные игры': 'table_game',
    'настольная игра': 'table_game',
    'other': 'other',
    'другое': 'other',
  };
  
  const category = categoryMap[session.category_session?.toLowerCase()] || 
                   categoryMap[session.type_session?.toLowerCase()] || 
                   'other';

  const typeSessionLower = session.type_session?.toLowerCase() || '';
  const typePlace: 'online' | 'offline' = 
    typeSessionLower === 'офлайн' || typeSessionLower === 'offline' ? 'offline' : 'online';

  const eventPlace = typePlace === 'offline' 
    ? (session.location || session.city || '') 
    : '';

  console.log('[sessionToEvent] 🔍', session.title);
  console.log('  - type_session:', session.type_session);
  console.log('  - location:', session.location);
  console.log('  - typePlace:', typePlace);
  console.log('  - eventPlace:', eventPlace);
  
  return {
    id: session.id.toString(),
    title: session.title,
    date,
    genres: session.genres || [],
    currentParticipants: session.current_users,
    maxParticipants: session.max_users,
    duration,
    imageUri: normalizeImageUrl(session.image_url),
    description: '',
    typeEvent: session.type_session,
    typePlace,
    eventPlace,
    publisher: '',
    publicationDate: year.toString(),
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

export const groupSessionsToEvents = (
  groupData: PublicGroupResponse,
  onSessionUpdate?: () => void
): Event[] => {
  if (!groupData.sessions || groupData.sessions.length === 0) {
    return [];
  }

  const sessionTypeToCategory: { [key: string]: Event['category'] } = {
    'Фильмы': 'movie',
    'Игры': 'game',
    'Настольные игры': 'table_game',
    'Другое': 'other',
  };

  return groupData.sessions.map(item => {
    const sessionPlace = item.session.session_place.toLowerCase();
    const typePlace: 'online' | 'offline' = 
      sessionPlace === 'офлайн' || sessionPlace === 'offline' ? 'offline' : 'online';

    const location = item.metadata?.Location || (item.metadata as any)?.location || '';

    const eventPlace = typePlace === 'offline' ? location : '';

    console.log('[groupSessionsToEvents] 🔍', item.session.title);
    console.log('  - session_place:', item.session.session_place);
    console.log('  - Location:', location);
    console.log('  - typePlace:', typePlace);
    console.log('  - eventPlace:', eventPlace);

    return {
      id: item.session.id.toString(),
      title: item.session.title,
      date: new Date(item.session.start_time).toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric'
      }),
      genres: item.metadata?.Genres || (item.metadata as any)?.genres || [],
      currentParticipants: item.session.current_users,
      maxParticipants: item.session.count_users_max,
      duration: `${item.session.duration} мин`,
      imageUri: item.session.image_url,
      description: item.metadata?.Notes || (item.metadata as any)?.notes || '',
      typeEvent: item.session.session_type,
      typePlace,
      eventPlace,
      publisher: groupData.name,
      publicationDate: item.metadata?.Year?.toString() || (item.metadata as any)?.year?.toString() || '',
      ageRating: item.metadata?.AgeLimit || (item.metadata as any)?.ageLimit || '',
      category: sessionTypeToCategory[item.session.session_type] || 'other',
      group: groupData.name,
      onSessionUpdate,
    };
  });
};

export const generateStatisticsData = (): StatisticsDataItem[] => {
  return [
    { name: 'Фильмы', percentage: 30, color: '#FF6B6B', legendFontColor: '#000' },
    { name: 'Игры', percentage: 25, color: '#4ECDC4', legendFontColor: '#000' },
    { name: 'Настолки', percentage: 20, color: '#45B7D1', legendFontColor: '#000' },
    { name: 'Другое', percentage: 25, color: '#FFA07A', legendFontColor: '#000' },
  ];
};