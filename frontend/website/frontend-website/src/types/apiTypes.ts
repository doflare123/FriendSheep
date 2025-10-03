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