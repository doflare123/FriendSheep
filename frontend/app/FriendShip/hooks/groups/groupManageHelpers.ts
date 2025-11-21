// eslint-disable-next-line import/no-unresolved
import { LOCAL_IP } from '@env';

export type ContactIconType = 'discord' | 'vk' | 'telegram' | 'twitch' | 'youtube' | 'whatsapp' | 'max';

export function convertToRFC3339(dateString: string): string {
  const [datePart, timePart] = dateString.split(' ');
  const [day, month, year] = datePart.split('.');
  const [hour, minute] = timePart.split(':');
  
  const date = new Date(
    parseInt(year),
    parseInt(month) - 1,
    parseInt(day),
    parseInt(hour),
    parseInt(minute)
  );
  
  return date.toISOString();
}

export function normalizeImageUrl(url: string): string {
  if (url && url.includes('localhost')) {
    return url.replace('http://localhost:8080', 'http://' + LOCAL_IP + ':8080');
  }
  return url;
}

export function formatSessionDate(isoDate: string): string {
  const date = new Date(isoDate);
  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const year = date.getFullYear();
  
  return `${day}.${month}.${year}`;
}

export function getContactType(name: string, link: string): ContactIconType {
  const lowerLink = link.toLowerCase();
  const lowerName = name.toLowerCase();
  
  if (lowerLink.includes('discord') || lowerName.includes('discord')) return 'discord';
  if (lowerLink.includes('vk.com') || lowerName.includes('vk')) return 'vk';
  if (lowerLink.includes('t.me') || lowerLink.includes('telegram') || lowerName.includes('telegram')) return 'telegram';
  if (lowerLink.includes('twitch') || lowerName.includes('twitch')) return 'twitch';
  if (lowerLink.includes('youtube') || lowerLink.includes('youtu')) return 'youtube';
  if (lowerLink.includes('wa.me') || lowerLink.includes('whatsapp') || lowerName.includes('whatsapp')) return 'whatsapp';
  
  return 'max';
}

export const categoryToSessionType: { [key: string]: string } = {
  'movie': 'Фильмы',
  'game': 'Игры',
  'table_game': 'Настольные игры',
  'other': 'Другое',
};

export const sessionTypeToCategory: { [key: string]: 'movie' | 'game' | 'table_game' | 'other' } = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};