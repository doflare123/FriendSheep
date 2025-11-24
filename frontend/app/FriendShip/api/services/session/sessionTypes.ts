export interface CreateSessionData {
  title: string;
  session_type: string;
  session_place: number;
  group_id: number;
  start_time: string;
  duration?: number;
  count_users: number;
  genres?: string;
  fields?: string;
  location?: string;
  year?: number;
  country?: string;
  age_limit?: string;
  notes?: string;
  image: {
    uri: string;
    name: string;
    type: string;
  };
}

export interface UpdateSessionData {
  title?: string;
  session_type_id?: number;
  session_place_id?: number;
  start_time?: string;
  end_time?: string;
  duration?: number;
  count_users_max?: number;
  genres?: string[];
  image_url?: string;
  location?: string;
  year?: number;
  country?: string;
  age_limit?: string;
  notes?: string;
}

export interface JoinSessionData {
  group_id: number;
  session_id: number;
}

export interface SessionInfo {
  id: number;
  title: string;
  category_session: string;
  type_session: string;
  city: string;
  start_time: string;
  end_time: string;
  current_users: number;
  max_users: number;
  image_url: string;
  genres: string[];
  location: string;
  status: string;
}

export interface SessionDetailResponse {
  session: {
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
    group_id: number;
  };
  metadata: {
    sessionID: number;
    genres: string[];
    location: string;
    year: number;
    country: string;
    ageLimit: string;
    notes: string;
    fields: Record<string, any>;
  };
}