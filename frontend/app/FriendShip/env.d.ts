declare module '@env' {
  export const API_BASE_URL: string;
  export const LOCAL_IP: string;
  export const KINOPOISK_API_KEY: string;
  export const YANDEX_JS_API_KEY: string;
  export const RAWG_API_KEY: string;
}

declare namespace NodeJS {
  interface ProcessEnv {
    EXPO_PUBLIC_API_URL: string;
    EXPO_PUBLIC_KINOPOISK_API_KEY: string;
    EXPO_PUBLIC_RAWG_API_KEY: string;
    EXPO_PUBLIC_YANDEX_JS_API_KEY: string;
  }
}