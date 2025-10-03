// src/types/Groups.ts

// Интерфейсы для групп
export interface Contact {
  link: string;
  name: string;
}

export interface User {
  image: string;
  name: string;
}

export interface SessionMetadata {
  ageLimit: string;
  country: string;
  fields: Record<string, any>;
  genres: string[];
  location: string;
  notes: string;
  sessionID: number;
  year: number;
}

export interface Session {
  count_users_max: number;
  current_users: number;
  duration: number;
  end_time: string;
  group_id: number;
  id: number;
  image_url: string;
  session_place: string;
  session_type: string;
  start_time: string;
  title: string;
}

export interface SessionWithMetadata {
  metadata: SessionMetadata;
  session: Session;
}

export interface GroupData {
  categories: string[];
  city: string;
  contacts: Contact[];
  count_members: number;
  description: string;
  id: number;
  image: string;
  name: string;
  sessions: SessionWithMetadata[];
  users: User[];
  small_description?: string;
  private?: boolean;
}

export interface GroupProfileProps {
  groupData: GroupData;
}

export interface SmallGroup {
  id: string;
  name: string;
  description: string;
  memberCount: number;
  image?: string;
  categories: ('games' | 'movies' | 'board' | 'other')[];
  city: string;
}