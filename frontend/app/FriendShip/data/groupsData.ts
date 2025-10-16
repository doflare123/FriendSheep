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