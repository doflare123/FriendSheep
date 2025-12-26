// src/types/UserData.ts

import { EventCardProps } from "./Events";

export interface PopularGenre {
  count: number;
  name: string;
}

export interface UserStats {
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

export interface UserDataResponse {
  id: number;
  data_register: string;
  enterprise: boolean;
  image: string;
  name: string;
  us: string;
  popular_genres: PopularGenre[];
  recent_sessions: EventCardProps[]; 
  status: string;
  telegram_link: boolean;
  tiles: string[];
  upcoming_sessions: EventCardProps[];
  user_stats: UserStats;
}
