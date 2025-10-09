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