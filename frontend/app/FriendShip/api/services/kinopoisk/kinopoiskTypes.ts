export interface KinopoiskSearchResponse {
  docs: KinopoiskMovie[];
  total: number;
  limit: number;
  page: number;
  pages: number;
}

export interface KinopoiskMovie {
  id: number;
  name: string;
  alternativeName?: string;
  enName?: string;
  type?: string;
  year: number;
  description?: string;
  shortDescription?: string;
  rating?: {
    kp?: number;
    imdb?: number;
  };
  movieLength?: number;
  ageRating?: number;
  poster?: {
    url?: string;
    previewUrl?: string;
  };
  genres?: {
    name: string;
  }[];
  countries?: {
    name: string;
  }[];
  persons?: {
    id: number;
    name: string;
    enName?: string;
    profession: string;
  }[];
}

export interface KinopoiskAutoFillData {
  title: string;
  description: string;
  genres: string[];
  publisher: string;
  year: number;
  country: string;
  ageRating: string;
  duration: string;
  imageUrl: string;
}