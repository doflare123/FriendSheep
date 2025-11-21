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