export interface RequestItem {
  id: string;
  name: string;
  username: string;
  imageUri: string;
}

export interface ConfirmationModalState {
  visible: boolean;
  action: string;
  title: string;
  message: string;
}

export const CATEGORY_MAPPING: { [key: string]: string } = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};

export const CATEGORY_IDS: { [key: string]: number } = {
  'movie': 1,
  'game': 2,
  'table_game': 3,
  'other': 4,
};

export type TabType = 'info' | 'subscribers' | 'requests' | 'events';