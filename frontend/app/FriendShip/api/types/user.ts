
export interface Genre {
  name: string;
  count: number;
}

export interface Session {
  id: number;
  title: string;
  image_url: string;
  location: string;
  city: string;
  start_time: string;
  end_time: string;
  current_users: number;
  max_users: number;
  status: string;
  category_session: string;
  type_session: string;
  genres: string[];
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

export interface UserStats {
  count_all: number;
  count_films: number;
  count_games: number;
  count_table_games: number;
  count_another: number;
  count_create_session: number;
  most_pop_day: string;
  most_big_session: number;
  series_session_count: number;
}

export interface UserProfile {
  name: string;
  us: string;
  image: string;
  status: string;
  data_register: string;
  telegram_link: boolean;
  enterprise: boolean;
  tiles: string[];
  user_stats: {
    count_all: number;
    count_films: number;
    count_games: number;
    count_table_games: number;
    count_another: number;
    most_pop_day: string;
    count_create_session: number;
    series_session_count: number;
    spent_time: number;
    most_big_session: number;
  };
  popular_genres: {
    name: string;
    count: number;
  }[];
  recent_sessions: SessionInfo[];
  upcoming_sessions: SessionInfo[];
}

export interface Subscription {
  id: number;
  name: string;
  image: string;
  small_description: string;
  type: string;
  category: string[];
  member_count: number;
}

export interface UpdateProfileRequest {
  name?: string;
  us?: string;
  image?: string;
  status?: string;
}

export interface UpdateProfileResponse {
  message: string;
  user: {
    name: string;
    us: string;
    email: string;
    image: string;
  };
}

export interface TileSettings {
  count_all: boolean;
  count_films: boolean;
  count_games: boolean;
  count_table: boolean;
  count_other: boolean;
  spent_time: boolean;
}

export interface SearchUserItem {
  id: number;
  name: string;
  us: string;
  image: string;
  status: string;
}

export interface SearchUsersResponse {
  users: SearchUserItem[];
  total: number;
  page: number;
  has_more: boolean;
}