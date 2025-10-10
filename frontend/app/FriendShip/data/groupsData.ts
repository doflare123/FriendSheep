import { Event } from '@/components/event/EventCard';

export interface GroupData {
  id: string;
  name: string;
  country: string;
  city: string;
  imageUri: string;
  description: string;
  subscribersCount: number;
  categories: ('movie' | 'game' | 'table_game' | 'other')[];
  mode: 'manage' | 'view';
  subscribers: Subscriber[];
  sessions: Event[];
  contacts: Contact[];
}

export interface Subscriber {
  id: string;
  name: string;
  imageUri: string;
}

export interface Contact {
  id: string;
  name: string;
  icon: any;
  description: string;
  link: string;
}

export const groupsData: Record<string, GroupData> = {
  '1': {
    id: '1',
    name: 'Мега крутая группа',
    country: 'Россия',
    city: 'Калининград',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
    description: 'ЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖЖ...',
    subscribersCount: 10000,
    categories: ['movie', 'game', 'table_game'],
    mode: 'manage',
    subscribers: [
      {
        id: '1',
        name: 'Чел',
        imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      },
      {
        id: '2',
        name: 'Flex228666',
        imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      },
      {
        id: '3',
        name: 'Крутой',
        imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      },
      {
        id: '4',
        name: 'Чел',
        imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      },
      {
        id: '5',
        name: 'Чел',
        imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      },
      {
        id: '6',
        name: 'Чел',
        imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      },
    ],
    sessions: [
      {
        id: '1',
        title: 'Крестный отец',
        date: '05.12.2024',
        genres: ['Драма', 'Криминал', 'Триллер'],
        currentParticipants: 5,
        maxParticipants: 5,
        duration: '175 минут',
        imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
        description: 'Легендарный фильм о мафии',
        typeEvent: 'Просмотр фильма',
        typePlace: 'offline',
        eventPlace: 'Калининград',
        publisher: 'Мега крутая группа',
        publicationDate: '01.12.2024',
        ageRating: '18+',
        category: 'movie',
      },
      {
        id: '2',
        title: 'Крестный отец 2',
        date: '06.12.2024',
        genres: ['Драма', 'Криминал'],
        currentParticipants: 3,
        maxParticipants: 5,
        duration: '200 минут',
        imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
        description: 'Продолжение легендарного фильма',
        typeEvent: 'Просмотр фильма',
        typePlace: 'offline',
        eventPlace: 'Калининград',
        publisher: 'Мега крутая группа',
        publicationDate: '01.12.2024',
        ageRating: '18+',
        category: 'movie',
      },
    ],
    contacts: [
      {
        id: 'discord',
        name: 'Discord',
        icon: require('../assets/images/groups/contacts/discord.png'),
        description: 'Группка',
        link: 'https://discord.gg/example',
      },
      {
        id: 'telegram',
        name: 'Telegram',
        icon: require('../assets/images/groups/contacts/telegram.png'),
        description: 'Канальчик',
        link: 'https://t.me/example',
      },
      {
        id: 'vk',
        name: 'VK',
        icon: require('../assets/images/groups/contacts/vk.png'),
        description: 'Страница автора',
        link: 'https://vk.com/example',
      },
    ],
  },
};

export const getGroupData = (id: string): GroupData | null => {
  return groupsData[id] || null;
};