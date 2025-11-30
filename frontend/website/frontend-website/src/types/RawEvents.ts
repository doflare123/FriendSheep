export interface RawSession {
  id: number;
  title: string;
  start_time: string;
  end_time: string | null;
  current_users: number;
  max_users: number;
  image_url: string;
  status: string;
  type_session: string;   
  category_session: string;
  location: string;  
  genres: string[];
  city?: string;
  date?: string;
}


export interface RawUserStats {
  count_all: number;
  count_another: number;
  count_create_session: number;
  count_films: number;
  count_games: number;
  count_table_games: number;
  most_big_session: number;
  most_pop_day: string;
  series_session_count: number;
  spent_time: number;
}


export interface RawPopularGenre {
  count: number;
  name: string;
}


export interface RawUserDataResponse {
  data_register: string;
  enterprise: boolean;
  image: string;
  name: string;
  us: string;
  popular_genres: RawPopularGenre[];
  recent_sessions: RawSession[];
  status: string;
  telegram_link: boolean;
  tiles: string[];
  upcoming_sessions: RawSession[];
  user_stats: RawUserStats;
}