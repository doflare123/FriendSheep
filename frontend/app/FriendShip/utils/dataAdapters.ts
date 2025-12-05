import { PublicGroupResponse } from '@/api/services/group/groupService';
import { Session, SessionInfo } from '@/api/types/user';
import { formatSessionDate, sessionTypeToCategory } from '@/hooks/groups/groupManageHelpers';
import { normalizeImageUrl } from '@/utils/imageUtils';
import { extractCityFromAddress } from './addressUtils';

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

export const sessionsToEvents = (sessions: SessionInfo[]): Event[] => {
  if (!sessions || sessions.length === 0) {
    return [];
  }

  return sessions.map((session) => {
    const typePlace = (session.type_session || '').toLowerCase();
    const eventTypePlace: 'online' | 'offline' = 
      typePlace === 'оффлайн' || typePlace === 'offline' ? 'offline' : 'online';

    let eventPlace = '';
    if (eventTypePlace === 'offline') {
      if (session.location) {
        const city = extractCityFromAddress(session.location);
        eventPlace = city || session.city || session.location;
      } else {
        eventPlace = session.city || '';
      }
    }

    return {
      id: session.id.toString(),
      title: session.title,
      date: formatSessionDate(session.start_time),
      genres: session.genres || [],
      currentParticipants: session.current_users || 0,
      maxParticipants: session.max_users,
      duration: `${Math.floor((new Date(session.end_time).getTime() - new Date(session.start_time).getTime()) / 60000)} мин`,
      imageUri: session.image_url || '',
      description: '',
      typeEvent: session.category_session,
      typePlace: eventTypePlace,
      eventPlace: session.location || session.city || '',
      publisher: '',
      publicationDate: '',
      ageRating: '',
      category: sessionTypeToCategory[session.category_session] || 'other',
      group: '',
    };
  });
};

export const groupSessionsToEvents = (
  groupData: PublicGroupResponse,
  onSessionUpdate?: () => void
): Event[] => {
  if (!groupData.sessions || groupData.sessions.length === 0) {
    return [];
  }

  return groupData.sessions.map((sessionWrapper) => {
    const session = sessionWrapper.session;
    const metadata = sessionWrapper.metadata;

    const sessionPlace = (session.session_place || '').toLowerCase();
    const typePlace: 'online' | 'offline' = 
      sessionPlace === 'оффлайн' || sessionPlace === 'offline' ? 'offline' : 'online';

    let eventPlace = '';
    if (typePlace === 'offline' && metadata?.Location) {
      const city = extractCityFromAddress(metadata.Location);
      eventPlace = city || metadata.Location;
    }

    console.log('[groupSessionsToEvents] 🔍', session.title);
    console.log('  - session_place:', session.session_place);
    console.log('  - Location:', metadata?.Location);
    console.log('  - typePlace:', typePlace);
    console.log('  - eventPlace:', eventPlace);

    return {
      id: session.id.toString(),
      title: session.title,
      date: formatSessionDate(session.start_time),
      genres: metadata?.Genres || [],
      currentParticipants: session.current_users || 0,
      maxParticipants: session.count_users_max,
      duration: `${session.duration} мин`,
      imageUri: session.image_url || '',
      description: metadata?.Notes || '',
      typeEvent: session.session_type,
      typePlace: typePlace,
      eventPlace: eventPlace,
      publisher: metadata?.Country || '',
      publicationDate: metadata?.Year?.toString() || '',
      ageRating: metadata?.AgeLimit || '',
      category: sessionTypeToCategory[session.session_type] || 'other',
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