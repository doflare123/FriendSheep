import { KinopoiskAPI } from './kinopoisk';
import { YandexMapsAPI } from './yandex_maps';

const KINOPOISK_API_KEY = process.env.NEXT_PUBLIC_KINOPOISK_API_KEY || '';
const YANDEX_MAPS_API_KEY = process.env.NEXT_PUBLIC_YANDEX_MAPS_API_KEY || '';

export const kinopoiskAPI = new KinopoiskAPI(KINOPOISK_API_KEY);
export const yandexMapsAPI = new YandexMapsAPI({ 
  apiKey: YANDEX_MAPS_API_KEY,
  lang: 'ru_RU' 
});

export * from './kinopoisk';
export * from './yandex_maps';