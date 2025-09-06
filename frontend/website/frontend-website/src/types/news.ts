// src/types/news.ts

export interface User {
  createdAt: string;
  data_register: string;
  email: string;
  enterprise: boolean;
  image: string;
  name: string;
  password: string;
  updatedAt: string;
  us: string;
}

export interface Comment {
  id: number;
  news_id?: number;
  text: string;
  user?: User;
  user_id: number;
  // Поля от сервера для отдельной новости
  image?: string;
  name?: string;
}

export interface Content {
  id: number;
  news_id: number;
  text: string;
}

export interface NewsItem {
  comments: Comment[];
  content: Content[] | string; // Может быть массивом или строкой
  createdAt: string;
  created_time: string;
  description: string;
  id: number;
  image: string;
  title: string;
  updatedAt: string;
}

export interface NewsResponse {
  hasMore: boolean;
  items: NewsItem[];
  page: number;
  total: number;
}

// Отдельный интерфейс для детальной страницы новости
export interface NewsDetail {
  comments: NewsComment[];
  content: string;
  created_time: string;
  description: string;
  id: number;
  image: string;
  title: string;
}