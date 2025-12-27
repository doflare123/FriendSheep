declare module 'leo-profanity' {
  export function loadDictionary(lang?: string): void;
  export function check(text: string): boolean;
  export function clean(text: string, placeholder?: string): string;
  export function list(text: string): string[];
  export function add(words: string[]): void;
  export function remove(words: string[]): void;
  export function clearList(): void;
  export function getDictionary(lang?: string): string[];
}