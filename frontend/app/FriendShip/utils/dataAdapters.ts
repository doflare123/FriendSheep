import { PublicGroupResponse } from '@/api/services/group';
import { SessionInfo } from '@/api/types/user';
import { formatSessionDate, sessionTypeToCategory } from '@/hooks/groups/groupManageHelpers';
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

    const category = sessionTypeToCategory[session.category_session] || 'other';

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
      eventPlace: eventPlace,
      publisher: '',
      publicationDate: '',
      ageRating: '',
      category: category,
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