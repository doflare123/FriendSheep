export interface KinopoiskMovie {
  id: number;
  nameRu?: string;
  nameEn?: string;
  nameOriginal?: string;
  description?: string;
  shortDescription?: string;
  year?: number;
  filmLength?: number;
  countries?: Array<{ country: string }>;
  genres?: Array<{ genre: string }>;
  ratingAgeLimits?: string;
  posterUrl?: string;
  posterUrlPreview?: string;
}

export class KinopoiskAPI {
  private apiKey: string;
  private baseUrl = 'https://kinopoiskapiunofficial.tech/api';

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async searchMovies(query: string): Promise<KinopoiskMovie[]> {
    try {
      const response = await fetch(
        `${this.baseUrl}/v2.1/films/search-by-keyword?keyword=${encodeURIComponent(query)}`,
        {
          headers: {
            'X-API-KEY': this.apiKey,
            'Content-Type': 'application/json',
          },
        }
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return data.films || [];
    } catch (error) {
      console.error('Ошибка поиска фильмов:', error);
      throw error;
    }
  }

  async getMovieDetails(id: number): Promise<KinopoiskMovie | null> {
    try {
      const response = await fetch(`${this.baseUrl}/v2.2/films/${id}`, {
        headers: {
          'X-API-KEY': this.apiKey,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error('Ошибка получения деталей фильма:', error);
      throw error;
    }
  }
}