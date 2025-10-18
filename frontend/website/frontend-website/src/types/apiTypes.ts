export interface Counters {
  count_all: boolean;
  count_films: boolean;
  count_games: boolean;
  count_other: boolean;
  count_table: boolean;
  spent_time: boolean;
}

export interface UpdateProfileRequest {
  image?: string;
  name?: string;
  status?: string;
  us?: string;
}

export interface SessionData {
  city: string;
  count_users_max: number;
  current_users: number;
  duration: number;
  genres: string[];
  group_id: number;
  id: number;
  image_url: string;
  session_place: string;
  session_type: string;
  start_time: string;
  title: string;
  fields?: string;
}

export interface EventMetadata {
  SessionID: number;
  Fields: {
    publisher?: string;
  };
  Location: string;
  Genres: string[];
  Year: number;
  Country: string;
  AgeLimit: string;
  Notes: string;
}

export interface EventSession {
  id: number;
  title: string;
  session_type: string;
  session_place: string;
  group_id: number;
  start_time: string;
  end_time: string;
  duration: number;
  current_users: number;
  count_users_max: number;
  image_url: string;
  is_sub?: boolean;
}

export interface EventFullResponse {
  session: EventSession;
  metadata: EventMetadata;
}