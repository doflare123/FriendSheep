import axios from 'axios';
// eslint-disable-next-line import/no-unresolved
import { RAWG_API_KEY } from '@env';

const RAWG_BASE_URL = 'https://api.rawg.io/api';

interface RAWGGame {
  id: number;
  name: string;
  description_raw: string;
  released: string;
  background_image: string;
  rating: number;
  genres: { id: number; name: string }[];
  publishers: { id: number; name: string }[];
  esrb_rating?: { id: number; name: string };
  playtime: number;
  tags: { id: number; name: string }[];
}

interface RAWGSearchResult {
  id: number;
  name: string;
  background_image: string;
  released: string;
  rating: number;
}

export interface GameAutoFillData {
  title: string;
  description: string;
  genres: string[];
  publisher: string;
  year: number;
  ageRating: string;
  duration: number;
  imageUrl: string;
}

class RawgService {
  private apiKey: string;

  constructor() {
    this.apiKey = RAWG_API_KEY || '';
    if (!this.apiKey) {
      console.warn('[RawgService] ⚠️ RAWG_API_KEY не найден в .env');
    }
  }

  async searchGames(query: string): Promise<RAWGSearchResult[]> {
    try {
      console.log('[RawgService] 🔍 Поиск игр:', query);

      const response = await axios.get(`${RAWG_BASE_URL}/games`, {
        params: {
          key: this.apiKey,
          search: query,
          page_size: 5,
          locale: 'ru',
        },
      });

      console.log('[RawgService] ✅ Найдено игр:', response.data.results.length);
      return response.data.results;
    } catch (error: any) {
      console.error('[RawgService] ❌ Ошибка поиска:', error);
      throw new Error('Не удалось найти игры');
    }
  }

  async getGameDetails(gameId: number): Promise<RAWGGame> {
    try {
      console.log('[RawgService] 📋 Загрузка деталей игры:', gameId);

      const response = await axios.get(`${RAWG_BASE_URL}/games/${gameId}`, {
        params: {
          key: this.apiKey,
          locale: 'ru',
        },
      });

      console.log('[RawgService] ✅ Детали загружены:', response.data.name);
      return response.data;
    } catch (error: any) {
      console.error('[RawgService] ❌ Ошибка загрузки деталей:', error);
      throw new Error('Не удалось загрузить информацию об игре');
    }
  }

  async getAutoFillData(gameName: string): Promise<GameAutoFillData | null> {
    try {
      console.log('[RawgService] 🎮 Автозаполнение для:', gameName);

      const searchResults = await this.searchGames(gameName);

      if (searchResults.length === 0) {
        console.log('[RawgService] ❌ Игра не найдена');
        return null;
      }

      const firstResult = searchResults[0];

      const gameDetails = await this.getGameDetails(firstResult.id);

      const rawDescription = gameDetails.description_raw || '';
      let translatedDescription = await this.translateText(rawDescription);

      if (translatedDescription.length > 500) {
        translatedDescription = translatedDescription.substring(0, 497) + '...';
      }

      const autoFillData: GameAutoFillData = {
        title: gameDetails.name,
        description: translatedDescription,
        genres: gameDetails.genres
          .map(g => this.translateGenre(g.name))
          .slice(0, 5),
        publisher: gameDetails.publishers.length > 0 
          ? gameDetails.publishers[0].name 
          : '',
        year: gameDetails.released 
          ? new Date(gameDetails.released).getFullYear() 
          : new Date().getFullYear(),
        ageRating: gameDetails.esrb_rating 
          ? this.formatEsrbRating(gameDetails.esrb_rating.name)
          : this.guessAgeRatingFromTags(gameDetails.tags || []),
        duration: (gameDetails.playtime || 2) * 60,
        imageUrl: gameDetails.background_image || '',
      };

      console.log('[RawgService] ✅ Автозаполнение завершено:', autoFillData.title);
      return autoFillData;
    } catch (error: any) {
      console.error('[RawgService] ❌ Ошибка автозаполнения:', error);
      throw error;
    }
  }

  private async translateText(text: string): Promise<string> {
    try {
      if (text.length < 10) {
        return text;
      }

      const textToTranslate = text.length > 400 ? text.substring(0, 397) + '...' : text;

      console.log('[RawgService] 🔄 Перевод описания...');

      try {
        const response = await axios.post(
          'https://libretranslate.com/translate',
          {
            q: textToTranslate,
            source: 'en',
            target: 'ru',
            format: 'text',
          },
          {
            timeout: 10000,
          }
        );

        console.log('[RawgService] ✅ Описание переведено (LibreTranslate)');
        return response.data.translatedText;
      } catch (libreError) {
        console.log('[RawgService] ⚠️ LibreTranslate недоступен, пробуем MyMemory...');

        const myMemoryResponse = await axios.get('https://api.mymemory.translated.net/get', {
          params: {
            q: textToTranslate,
            langpair: 'en|ru',
          },
          timeout: 10000,
        });

        if (myMemoryResponse.data.responseStatus === 200) {
          console.log('[RawgService] ✅ Описание переведено (MyMemory)');
          return myMemoryResponse.data.responseData.translatedText;
        }

        throw new Error('MyMemory также не сработал');
      }
    } catch (error: any) {
      console.error('[RawgService] ⚠️ Все сервисы перевода недоступны, используем оригинал:', error.message);
      return text.length > 400 ? text.substring(0, 397) + '...' : text;
    }
  }

  private translateGenre(genre: string): string {
    const genreTranslations: Record<string, string> = {
      'Action': 'Экшен',
      'Adventure': 'Приключения',
      'RPG': 'РПГ',
      'Strategy': 'Стратегия',
      'Shooter': 'Шутер',
      'Puzzle': 'Головоломка',
      'Racing': 'Гонки',
      'Sports': 'Спорт',
      'Fighting': 'Файтинг',
      'Platformer': 'Платформер',
      'Simulation': 'Симулятор',
      'Arcade': 'Аркада',
      'Casual': 'Казуальная',
      'Family': 'Семейная',
      'Educational': 'Образовательная',

      'Massively Multiplayer': 'Массовая многопользовательская',
      'MMO': 'ММО',
      'MMORPG': 'ММОРПГ',
      'Multiplayer': 'Мультиплеер',

      'Indie': 'Инди',

      'Board Games': 'Настольные игры',
      'Card': 'Карточная',
    };

    return genreTranslations[genre] || genre;
  }

  private truncateDescription(description: string): string {
    if (description.length <= 500) {
      return description;
    }
    return description.substring(0, 497) + '...';
  }

  private formatEsrbRating(rating?: string): string {
    if (!rating) return '';

    const ratingMap: Record<string, string> = {
      'Everyone': '0+',
      'Everyone 10+': '10+',
      'Teen': '13+',
      'Mature': '17+',
      'Adults Only': '18+',
    };

    return ratingMap[rating] || rating;
  }

  private guessAgeRatingFromTags(tags: { name: string }[]): string {
    const tagNames = tags.map(t => t.name.toLowerCase());

    const matureKeywords = [
      'gore', 'violent', 'nudity', 'sexual content', 
      'blood', 'nsfw', 'mature', 'horror'
    ];
    
    const teenKeywords = [
      'action', 'shooter', 'pvp', 'multiplayer',
      'competitive', '战斗'
    ];

    const everyoneKeywords = [
      'casual', 'family friendly', 'cute', 'relaxing',
      'puzzle', 'educational'
    ];

    if (tagNames.some(tag => matureKeywords.some(keyword => tag.includes(keyword)))) {
      return '18+';
    }

    if (tagNames.some(tag => everyoneKeywords.some(keyword => tag.includes(keyword)))) {
      return '6+';
    }

    if (tagNames.some(tag => teenKeywords.some(keyword => tag.includes(keyword)))) {
      return '12+';
    }

    return '';
  }

  async getPopularGames(limit: number = 10): Promise<RAWGSearchResult[]> {
    try {
      console.log('[RawgService] 🔥 Загрузка популярных игр');

      const response = await axios.get(`${RAWG_BASE_URL}/games`, {
        params: {
          key: this.apiKey,
          ordering: '-rating',
          page_size: limit,
        },
      });

      console.log('[RawgService] ✅ Популярные игры загружены:', response.data.results.length);
      return response.data.results;
    } catch (error: any) {
      console.error('[RawgService] ❌ Ошибка загрузки популярных игр:', error);
      throw new Error('Не удалось загрузить популярные игры');
    }
  }

  async getNewGames(limit: number = 10): Promise<RAWGSearchResult[]> {
    try {
      console.log('[RawgService] 🆕 Загрузка новых игр');

      const response = await axios.get(`${RAWG_BASE_URL}/games`, {
        params: {
          key: this.apiKey,
          ordering: '-released',
          page_size: limit,
        },
      });

      console.log('[RawgService] ✅ Новые игры загружены:', response.data.results.length);
      return response.data.results;
    } catch (error: any) {
      console.error('[RawgService] ❌ Ошибка загрузки новых игр:', error);
      throw new Error('Не удалось загрузить новые игры');
    }
  }
}

export default new RawgService();