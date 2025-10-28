import { Event } from '@/components/event/EventCard';
import { StatisticsDataItem } from '@/components/profile/StatisticsChart';
import { TileType } from '@/components/profile/TileSelectionModal';
import { Colors } from '@/constants/Colors';

export interface UserProfile {
  id: string;
  avatar: any;
  name: string;
  username: string;
  description: string;
  registrationDate: string;
  telegramLink?: string;
  stats: {
    media: number;
    games: number;
    table_games: number;
    other: number;
    hours: number;
    sessions: number;
  };
  selectedTiles: TileType[];
  favoriteGenres: { name: string; count: number }[];
  subscriptions: { id: number; image: any }[];
  completedSessions: Event[];
  upcomingSessions: Event[];
  statisticsData: StatisticsDataItem[];
}

export const mockUsers: Record<string, UserProfile> = {
  currentUser: {
    id: 'me',
    avatar: require('@/assets/images/profile/profile_avatar.jpg'),
    name: 'Та самая Игфи',
    username: '@lgfi_22',
    description: 'Всем привет! Я новый участник сего проекта 🫶',
    registrationDate: '21.11.2025',
    telegramLink: 'https://t.me/your_bot',
    stats: {
      media: 20,
      games: 20,
      table_games: 20,
      other: 20,
      hours: 20,
      sessions: 20,
    },
    selectedTiles: ['media', 'games', 'hours', 'sessions'],
    favoriteGenres: [
      { name: 'Боевики', count: 21 },
      { name: 'Приколы', count: 18 },
      { name: 'Страшилки', count: 18 },
      { name: 'Романтика', count: 18 },
      { name: 'РПГ', count: 18 },
    ],
    subscriptions: [
      { id: 1, image: require('@/assets/images/profile/profile_avatar.jpg') },
      { id: 2, image: require('@/assets/images/profile/profile_avatar.jpg') },
      { id: 3, image: require('@/assets/images/profile/profile_avatar.jpg') },
      { id: 4, image: require('@/assets/images/profile/profile_avatar.jpg') },
    ],
    completedSessions: [
      {
        id: '2',
        title: 'Матрица',
        date: '15.03.2004',
        imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
        description: "мяу",
        genres: ['Фантастика'],
        group: 'Мега крутая группа',
        currentParticipants: 48,
        maxParticipants: 50,
        duration: '136 минут',
        typeEvent: 'Фильм',
        typePlace: 'offline',
        eventPlace: 'Кинотеатр «Октябрь»',
        publisher: 'Warner Bros',
        publicationDate: '1999',
        ageRating: '16+',
        category: 'movie',
      },
      {
        id: '3',
        title: 'The Elder Scrolls V: Skyrim',
        date: '10.01.2004',
        imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
        description: "Киберспортивный турнир",
        genres: ['Шутер'],
        group: 'Мега крутая группа',
        currentParticipants: 32,
        maxParticipants: 64,
        duration: '240 минут',
        typeEvent: 'Турнир',
        typePlace: 'online',
        eventPlace: 'Steam',
        publisher: 'Valve',
        publicationDate: '2012',
        ageRating: '16+',
        category: 'game',
      },
    ],
    upcomingSessions: [
      {
        id: '1',
        title: 'Крестный отец',
        date: '12.02.2004',
        imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
        description: "ЭЩКЕРЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕ",
        genres: ['Драма', 'Криминал'],
        group: 'Мега крутая группа',
        currentParticipants: 52,
        maxParticipants: 52,
        duration: '175 минут',
        typeEvent: 'Фильм',
        typePlace: 'online',
        eventPlace: 'https://cinema.com',
        publisher: 'Paramount Pictures',
        publicationDate: '1972',
        ageRating: '18+',
        category: 'movie',
      },
    ],
    statisticsData: [
      { name: 'Боевики', percentage: 25, color: '#4A90E2', legendFontColor: Colors.black },
      { name: 'РПГ', percentage: 22, color: '#7B68EE', legendFontColor: Colors.black },
      { name: 'Приколы', percentage: 18, color: '#50C878', legendFontColor: Colors.black },
      { name: 'Страшилки', percentage: 15, color: '#FFB6C1', legendFontColor: Colors.black },
      { name: 'Романтика', percentage: 12, color: '#FFA500', legendFontColor: Colors.black },
      { name: 'Стратегии', percentage: 8, color: '#FF6B6B', legendFontColor: Colors.black },
    ],
  },
  '1': {
    id: '1',
    avatar: require('@/assets/images/profile/profile_avatar.jpg'),
    name: 'Лейс с крабом',
    username: '@laysKRAB',
    description: 'Вкуснее, чем Pringles 😎',
    registrationDate: '20.10.2025',
    telegramLink: 'https://t.me/lays_krab',
    stats: {
      media: 15,
      games: 25,
      table_games: 10,
      other: 5,
      hours: 30,
      sessions: 15,
    },
    selectedTiles: ['games', 'media', 'sessions', 'hours'],
    favoriteGenres: [
      { name: 'Шутеры', count: 30 },
      { name: 'РПГ', count: 25 },
      { name: 'Стратегии', count: 20 },
      { name: 'Хорроры', count: 15 },
      { name: 'Инди', count: 10 },
    ],
    subscriptions: [
      { id: 1, image: require('@/assets/images/profile/profile_avatar.jpg') },
      { id: 2, image: require('@/assets/images/profile/profile_avatar.jpg') },
      { id: 3, image: require('@/assets/images/profile/profile_avatar.jpg') },
    ],
    completedSessions: [
      {
        id: '4',
        title: 'Counter-Strike 2',
        date: '01.11.2025',
        imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
        description: "Соревновательная игра",
        genres: ['Шутер', 'Тактика'],
        group: 'Геймеры',
        currentParticipants: 10,
        maxParticipants: 10,
        duration: '180 минут',
        typeEvent: 'Турнир',
        typePlace: 'online',
        eventPlace: 'Steam',
        publisher: 'Valve',
        publicationDate: '2023',
        ageRating: '16+',
        category: 'game',
      },
    ],
    upcomingSessions: [
      {
        id: '5',
        title: 'Dota 2',
        date: '15.11.2025',
        imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
        description: "Командная игра",
        genres: ['MOBA', 'Стратегия'],
        group: 'Геймеры',
        currentParticipants: 8,
        maxParticipants: 10,
        duration: '120 минут',
        typeEvent: 'Игровая сессия',
        typePlace: 'online',
        eventPlace: 'Steam',
        publisher: 'Valve',
        publicationDate: '2013',
        ageRating: '12+',
        category: 'game',
      },
    ],
    statisticsData: [
      { name: 'Шутеры', percentage: 35, color: '#4A90E2', legendFontColor: Colors.black },
      { name: 'РПГ', percentage: 25, color: '#7B68EE', legendFontColor: Colors.black },
      { name: 'Стратегии', percentage: 20, color: '#50C878', legendFontColor: Colors.black },
      { name: 'Хорроры', percentage: 12, color: '#FFB6C1', legendFontColor: Colors.black },
      { name: 'Инди', percentage: 8, color: '#FFA500', legendFontColor: Colors.black },
    ],
  },
};

export const getUserById = (userId: string): UserProfile => {
  return mockUsers[userId] || mockUsers.currentUser;
};

export const getCurrentUser = (): UserProfile => {
  return mockUsers.currentUser;
};