import {RawSession, RawUserDataResponse} from './types/RawEvents';
import {EventCardProps} from './types/Events'
import {UserDataResponse} from './types/UserData'
import { getUserInfo } from './api/profile/getOwnProfile';

export const convertCategoriesToIds = (categories: string[]): number[] => {
    const categoryMap: { [key: string]: number } = {
      'movies': 1,    // Фильмы
      'games': 2,     // Игры
      'board': 3,     // Настольные игры
      'other': 4      // Другое
    };
    
    return categories.map(category => categoryMap[category]).filter(id => id !== undefined);
};

export const convertCategRuToEng = (categoryIds: string[]): ('games' | 'movies' | 'board' | 'other')[] => {
    const idMap: { [key: string]: 'games' | 'movies' | 'board' | 'other' } = {
      'Фильмы': 'movies',    // Фильмы
      'Игры': 'games',     // Игры
      'Настольные игры': 'board',     // Настольные игры
      'Другое': 'other'      // Другое
    };
    
    return categoryIds.map(id => idMap[id]).filter(category => category !== undefined);
};

export const convertIdsToCategories = (categoryIds: number[]): ('games' | 'movies' | 'board' | 'other')[] => {
    const idMap: { [key: number]: 'games' | 'movies' | 'board' | 'other' } = {
      1: 'movies',    // Фильмы
      2: 'games',     // Игры
      3: 'board',     // Настольные игры
      4: 'other'      // Другое
    };
    
    return categoryIds.map(id => idMap[id]).filter(category => category !== undefined);
};

export const convertSocialContactsToString = (socialContacts: { name: string; link: string }[]): string => {
    if (!Array.isArray(socialContacts) || socialContacts.length === 0) {
        return '';
    }

    return socialContacts
        .filter(contact => contact.name && contact.link) // Фильтруем пустые контакты
        .map(contact => `${contact.name.trim()}:${contact.link.trim()}`) // Форматируем каждый контакт
        .join(', '); // Объединяем через запятую и пробел
};

export const getAccesToken = (): string => {
  return localStorage.getItem('access_token') || '';
}

export function decodeJWT(token: string) {
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return decoded;
  } catch (error) {
    console.error('Ошибка декодирования JWT:', error);
    return null;
  }
}

export function debugJWT(token: string): void {
  try {
    // Разбиваем токен на части
    const parts = token.split('.');
    
    if (parts.length !== 3) {
      console.error('❌ Неверный формат JWT токена');
      console.log('Токен должен состоять из 3 частей, разделённых точками');
      return;
    }

    // Декодируем header (первая часть)
    const headerBase64 = parts[0].replace(/-/g, '+').replace(/_/g, '/');
    const header = JSON.parse(atob(headerBase64));

    // Декодируем payload (вторая часть)
    const payloadBase64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const payload = JSON.parse(
      decodeURIComponent(
        atob(payloadBase64)
          .split('')
          .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      )
    );

    // Красиво выводим в консоль
    console.log('🔐 JWT TOKEN DECODED');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    console.log('\n📋 HEADER:');
    console.log(header);
    console.log('\n📦 PAYLOAD:');
    console.log(payload);
    
    // Проверяем срок действия
    if (payload.exp) {
      const expirationDate = new Date(payload.exp * 1000);
      const isExpired = Date.now() >= payload.exp * 1000;
      console.log('\n⏰ EXPIRATION:');
      console.log(`  Истекает: ${expirationDate.toLocaleString()}`);
      console.log(`  Статус: ${isExpired ? '❌ Истёк' : '✅ Действителен'}`);
    }
    
    // Показываем время создания
    if (payload.iat) {
      const issuedDate = new Date(payload.iat * 1000);
      console.log('\n📅 ISSUED AT:');
      console.log(`  Создан: ${issuedDate.toLocaleString()}`);
    }

    console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    
  } catch (error) {
    console.error('❌ Ошибка при декодировании JWT:', error);
    console.log('Убедись, что передана корректная строка JWT токена');
  }
}

export const getUserData = async (): UserDataResponse | null => {
  try {
    const data = localStorage.getItem('userData');
    return data ? JSON.parse(data) : await updateUserData();
  } catch (error) {
    console.error('Ошибка чтения из localStorage:', error);
    return null;
  }
}

export const updateUserData = async (): UserDataResponse | null => {
  const accessToken: string = getAccesToken();
  
  let UserInfo;

  try {
    UserInfo = getUserInfo(accessToken);
    localStorage.setItem('userData', JSON.stringify(UserInfo));
  } catch (error) {
    console.error('Ошибка сохранения в localStorage:', error);
  }

  return UserInfo || null;
}

export const getCategoryIcon = (category: string): string => {
  const lowerCategory = category.toLowerCase();
  
  if (lowerCategory.includes('games') || lowerCategory =='игры') {
    return '/events/games.png';
  } else if (lowerCategory.includes('movies') || lowerCategory.includes('фильмы')) {
    return '/events/movies.png';
  } else if (lowerCategory.includes('boards') || lowerCategory.includes('настолки') || lowerCategory.includes('настольные игры')) {
    return '/events/board.png';
  } else {
    return '/events/other.png'; // default
  }
};

export const getSocialIcon = (name: string, link?: string, ): string => {
  const lowerLink = (link|| " ").toLowerCase();
  const lowerName = name.toLowerCase();
  
  if (lowerLink.includes('discord') || lowerName.includes('discord') || lowerName.includes('ds')) {
    return '/social/ds.png';
  } else if (lowerLink.includes('t.me') || lowerLink.includes('telegram') || lowerName.includes('telegram') || lowerName.includes('tg')) {
    return '/social/tg.png';
  } else if (lowerLink.includes('vk.com') || lowerName.includes('вконтакте') || lowerName.includes('vk')) {
    return '/social/vk.png';
  } else if (lowerLink.includes('wa.me') || lowerLink.includes('whatsapp') || lowerName.includes('whatsapp') || lowerName.includes('wa')) {
    return '/social/wa.png';
  } else if (lowerLink.includes('snapchat') || lowerName.includes('snapchat') || lowerName.includes('snap')) {
    return '/social/snap.png';
  } else {
    return '/default/soc_net.png';
  }
};

export function mapServerSessionToEvent(session: RawSession): EventCardProps {
  return {
    id: session.id,
    type: 'other',
    image: session.image_url ?? '',
    date: session.start_time ?? '',
    end_time: session.end_time ?? undefined,
    title: session.title ?? '',
    genres: session.genres ?? [],
    participants: session.current_users ?? 0,
    maxParticipants: session.max_users ?? 0,
    // остальные фронтовые поля — оставляем пустыми/дефолтными
    duration: undefined,
    location: typeof session.location === 'string' && session.location.toLowerCase() === 'online' ? 'online' : 'offline',
    adress: session.location ?? '',
    city: session.city ?? '',
    scale: undefined,
    isEditMode: undefined,
    onEdit: undefined,
    groupId: undefined,
  };
}

export function mapRawUserDataToUserData(raw: RawUserDataResponse): UserDataResponse {
  const recent_sessions = (raw.recent_sessions || []).map(mapServerSessionToEvent);
  const upcoming_sessions = (raw.upcoming_sessions || []).map(mapServerSessionToEvent);

  // Прямо возвращаем объект, приведя тип к UserDataResponse — поля совпадают по семантике.
  return {
    data_register: formatDate(raw.data_register),
    enterprise: raw.enterprise,
    image: raw.image,
    name: raw.name,
    us: raw.us,
    popular_genres: raw.popular_genres,
    recent_sessions,
    status: raw.status,
    telegram_link: raw.telegram_link,
    tiles: raw.tiles,
    upcoming_sessions,
    user_stats: raw.user_stats,
  } as UserDataResponse;
}

export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  const day = date.getDate().toString().padStart(2, '0');
  const month = (date.getMonth() + 1).toString().padStart(2, '0'); // месяцы с 0
  const year = date.getFullYear();
  return `${day}.${month}.${year}`;
}

export function formatTime(dateString: string): string {
  const date = new Date(dateString);
  const hours = date.getHours().toString().padStart(2, '0');
  const minutes = date.getMinutes().toString().padStart(2, '0');
  return `${hours}:${minutes}`;
}