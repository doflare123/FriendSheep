import { PublicGroupResponse } from '@/api/services/group/groupService';
import { Session } from '@/api/types/user';

export const publicGroupSessionsToSessions = (
  apiSessions: PublicGroupResponse['sessions']
): Session[] => {
  return apiSessions.map(item => ({
    id: item.session.id,
    title: item.session.title,
    image_url: item.session.image_url,
    location: item.metadata?.location || '',
    city: item.metadata?.location?.split(',')[0]?.trim() || '',
    start_time: item.session.start_time,
    end_time: item.session.end_time,
    current_users: item.session.current_users,
    max_users: item.session.count_users_max,
    status: 'upcoming',
    category_session: item.session.session_type,
    type_session: item.session.session_type,
    genres: item.metadata?.genres || [],
  }));
};