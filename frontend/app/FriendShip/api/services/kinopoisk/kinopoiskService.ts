// eslint-disable-next-line import/no-unresolved
import { KINOPOISK_API_KEY } from '@env';
import {
  KinopoiskAutoFillData,
  KinopoiskMovie,
  KinopoiskSearchResponse
} from './kinopoiskTypes';

const KINOPOISK_API_URL = 'https://api.kinopoisk.dev/v1.4';

const AVAILABLE_GENRES = [
  'Драма', 'Комедия', 'Боевик', 'Триллер', 'Ужасы', 'Фантастика',
  'Детектив', 'Приключения', 'Романтика', 'Криминал', 'Военный',
  'Исторический', 'Биография', 'Документальный', 'Анимация', 'Семейный',
  'Мюзикл', 'Вестерн', 'Спорт', 'Фэнтези'
];

const GENRE_MAPPING: Record<string, string> = {
  'драма': 'Драма',
  'комедия': 'Комедия',
  'боевик': 'Боевик',
  'триллер': 'Триллер',
  'ужасы': 'Ужасы',
  'фантастика': 'Фантастика',
  'детектив': 'Детектив',
  'приключения': 'Приключения',
  'мелодрама': 'Романтика',
  'романтика': 'Романтика',
  'криминал': 'Криминал',
  'военный': 'Военный',
  'история': 'Исторический',
  'исторический': 'Исторический',
  'биография': 'Биография',
  'документальный': 'Документальный',
  'мультфильм': 'Анимация',
  'аниме': 'Анимация',
  'анимация': 'Анимация',
  'семейный': 'Семейный',
  'мюзикл': 'Мюзикл',
  'вестерн': 'Вестерн',
  'спорт': 'Спорт',
  'фэнтези': 'Фэнтези',
};

class KinopoiskService {
  private apiKey: string;

  constructor() {
    this.apiKey = KINOPOISK_API_KEY;
  }

  async searchMovieByTitle(title: string): Promise<KinopoiskMovie | null> {
    try {
      console.log('[KinopoiskService] 🔍 Поиск фильма:', title);

      if (!this.apiKey) {
        throw new Error('API ключ Кинопоиска не найден');
      }

      const response = await fetch(
        `${KINOPOISK_API_URL}/movie/search?page=1&limit=1&query=${encodeURIComponent(title)}`,
        {
          method: 'GET',
          headers: {
            'X-API-KEY': this.apiKey,
            'accept': 'application/json',
          },
        }
      );

      if (!response.ok) {
        const errorText = await response.text();
        console.error('[KinopoiskService] ❌ Ошибка запроса:', response.status, errorText);
        throw new Error(`Ошибка API Кинопоиска: ${response.status}`);
      }

      const data: KinopoiskSearchResponse = await response.json();
      
      console.log('[KinopoiskService] 📦 Получено результатов:', data.total);

      if (data.docs && data.docs.length > 0) {
        const movie = data.docs[0];
        console.log('[KinopoiskService] ✅ Найден фильм:', movie.name);

        if (!movie.persons || movie.persons.length === 0) {
          console.log('[KinopoiskService] 🔄 Загружаем полную информацию о фильме...');
          const fullMovie = await this.getMovieById(movie.id);
          if (fullMovie) {
            return fullMovie;
          }
        }
        
        return movie;
      }

      console.log('[KinopoiskService] ⚠️ Фильм не найден');
      return null;
    } catch (error: any) {
      console.error('[KinopoiskService] ❌ Ошибка поиска:', error);
      throw error;
    }
  }

  async getMovieById(movieId: number): Promise<KinopoiskMovie | null> {
    try {
      console.log('[KinopoiskService] 🔍 Запрос полной информации для ID:', movieId);

      if (!this.apiKey) {
        throw new Error('API ключ Кинопоиска не найден');
      }

      const response = await fetch(
        `${KINOPOISK_API_URL}/movie/${movieId}`,
        {
          method: 'GET',
          headers: {
            'X-API-KEY': this.apiKey,
            'accept': 'application/json',
          },
        }
      );

      if (!response.ok) {
        console.error('[KinopoiskService] ⚠️ Не удалось получить полную информацию');
        return null;
      }

      const movie: KinopoiskMovie = await response.json();
      console.log('[KinopoiskService] ✅ Получена полная информация о фильме');
      
      return movie;
    } catch (error: any) {
      console.error('[KinopoiskService] ❌ Ошибка получения полной информации:', error);
      return null;
    }
  }

  private mapGenres(kinopoiskGenres: { name: string }[]): string[] {
    const mappedGenres: string[] = [];

    for (const genre of kinopoiskGenres) {
      const genreLower = genre.name.toLowerCase();
      const mappedGenre = GENRE_MAPPING[genreLower];

      if (mappedGenre && AVAILABLE_GENRES.includes(mappedGenre)) {
        if (!mappedGenres.includes(mappedGenre)) {
          mappedGenres.push(mappedGenre);
        }
      }
    }

    return mappedGenres;
  }

  private getDirector(movie: KinopoiskMovie): string {
    if (!movie.persons || movie.persons.length === 0) {
      console.log('[KinopoiskService] ⚠️ Нет данных о персонах фильма');
      return '';
    }

    const director = movie.persons.find(
      person => {
        const profession = person.profession?.toLowerCase() || '';
        return profession === 'режиссеры' || 
               profession === 'режиссёры' ||
               profession === 'режиссер' ||
               profession === 'режиссёр' ||
               profession === 'director';
      }
    );

    if (director) {
      console.log('[KinopoiskService] ✅ Найден режиссёр:', director.name);
      return director.name;
    }

    console.log('[KinopoiskService] ⚠️ Режиссёр не найден. Доступные профессии:', 
      movie.persons.map(p => p.profession).join(', '));
    return '';
  }

  convertMovieToAutoFillData(movie: KinopoiskMovie): KinopoiskAutoFillData {
    const genres = movie.genres ? this.mapGenres(movie.genres) : [];
    const director = this.getDirector(movie);
    const country = movie.countries && movie.countries.length > 0 
      ? movie.countries[0].name 
      : '';

    let ageRating = '';
    if (movie.ageRating && movie.ageRating > 0) {
      ageRating = `${movie.ageRating}+`;
      console.log('[KinopoiskService] ✅ Возрастной рейтинг:', ageRating);
    } else {
      console.log('[KinopoiskService] ⚠️ Возрастной рейтинг не указан');
    }

    let description = movie.description || movie.shortDescription || '';

    if (description.length > 300) {
      description = description.substring(0, 297) + '...';
      console.log('[KinopoiskService] ⚠️ Описание обрезано до 300 символов');
    }

    const duration = movie.movieLength ? movie.movieLength.toString() : '';

    const imageUrl = movie.poster?.url || movie.poster?.previewUrl || '';

    console.log('[KinopoiskService] 📋 Данные для автозаполнения:');
    console.log('  - Название:', movie.name);
    console.log('  - Жанры:', genres);
    console.log('  - Режиссёр:', director || '(не найден)');
    console.log('  - Год:', movie.year);
    console.log('  - Страна:', country || '(не указана)');
    console.log('  - Возраст:', ageRating || '(не указан)');
    console.log('  - Длительность:', duration || '(не указана)');
    console.log('  - Постер:', imageUrl ? 'Есть' : '(не найден)');

    return {
      title: movie.name,
      description,
      genres,
      publisher: director,
      year: movie.year,
      country,
      ageRating,
      duration,
      imageUrl,
    };
  }

  async getAutoFillData(title: string): Promise<KinopoiskAutoFillData | null> {
    try {
      const movie = await this.searchMovieByTitle(title);
      
      if (!movie) {
        return null;
      }

      return this.convertMovieToAutoFillData(movie);
    } catch (error: any) {
      console.error('[KinopoiskService] ❌ Ошибка получения данных:', error);
      throw error;
    }
  }
}

export default new KinopoiskService();