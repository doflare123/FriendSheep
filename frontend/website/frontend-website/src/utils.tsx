// notificationManager.ts
import { createRoot } from "react-dom/client";
import StatusNotif, { NotifProps } from "./components/statusNotif";
import * as LeoProfanity from 'leo-profanity';

class NotificationManager {
  private container: HTMLDivElement | null = null;
  private root: ReturnType<typeof createRoot> | null = null;
  private initialized = false;
  private isInitializing = false;

  private async init() {
    if (this.initialized) {
      console.log("Manager already initialized");
      return;
    }
    
    if (this.isInitializing) {
      console.log("Manager is initializing, waiting...");
      // Ждем пока инициализация завершится
      while (this.isInitializing) {
        await new Promise(resolve => setTimeout(resolve, 10));
      }
      return;
    }
    
    console.log("Initializing manager");
    this.isInitializing = true;
    
    this.container = document.createElement("div");
    this.container.id = "notif-root";
    document.body.appendChild(this.container);
    this.root = createRoot(this.container);
    this.root.render(<StatusNotif />);
    
    // Даем время React на монтирование компонента и регистрацию слушателя
    await new Promise(resolve => setTimeout(resolve, 50));
    
    this.initialized = true;
    this.isInitializing = false;
    console.log("Manager initialized");
  }

  async show(code: number, description: string, type?: NotifProps["type"]) {
    console.log("showNotification called with:", code, description);
    await this.init();

    let resolvedType: NotifProps["type"];
    if (type) {
      resolvedType = type;
    } else if (code >= 200 && code < 300) {
      resolvedType = "success";
    } else if (code >= 300 && code < 400) {
      resolvedType = "warning";
    } else {
      resolvedType = "error";
    }

    console.log("Dispatching event");
    const event = new CustomEvent("push-notif", {
      detail: { code, description, type: resolvedType },
    });
    window.dispatchEvent(event);
  }
}

const manager = new NotificationManager();

export function showNotification(
  code: number,
  description: string,
  type?: NotifProps["type"]
) {
  console.log("=== showNotification wrapper called ===");
  manager.show(code, description, type);
}

export const getTheme = (): 'light' | 'dark' => {
  if (typeof window === 'undefined') return 'light';
  
  const saved = localStorage.getItem('theme');
  if (saved === 'dark' || saved === 'light') {
    return saved;
  }
  
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
};

export const setTheme = (theme: 'light' | 'dark') => {
  if (typeof window === 'undefined') return;
  
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);
};

export const toggleTheme = () => {
  const current = getTheme();
  const newTheme = current === 'light' ? 'dark' : 'light';
  setTheme(newTheme);
  return newTheme;
};

export const initTheme = () => {
  if (typeof window === 'undefined') return;
  
  const theme = getTheme();
  document.documentElement.setAttribute('data-theme', theme);
};



LeoProfanity.loadDictionary('en');
LeoProfanity.loadDictionary('ru');

const additionalBadWords = [
  'fck', 'fuk', 'fuc', 'fukc', 'fuckc', 'fcuk', 'fucc',
  'sht', 'shyt', 'sh1t', 'shiit',
  'btch', 'biatch', 'bich',
  'dck', 'dik', 'dikk',
  'cnt', 'cnut',
  'хй', 'хy', 'xyй', 'хyй',
  'пзд', 'пздц',
  'еба', 'ебт', 'ебн',
  'бля', 'блд', 'блть',
  'xyй', 'xуй', 'хуи', 'хуя'
];

additionalBadWords.forEach(word => LeoProfanity.add(word));

function normalizeText(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^а-яёa-z0-9]/g, '');
}

function isSimilarToBadWord(word: string): boolean {
  const normalized = normalizeText(word);
  
  const similarityPatterns = [
    /f+[u|y|oo|0]+c*k+/i,
    /s+h+[i|1|y]+t+/i,
    /b+[i|1]+t+c*h+/i,
    /a+s+s+/i,
    /d+[i|1]+c*k+/i,
    /c+[u|y|oo]+n+t+/i,
    /п+[и|1]+з+д+/i,
    /[х|x]+[у|y|u]+[й|и|i|y]/i,
    /е+б+[а|o|л|т|н]+/i,
    /б+л+[я|д|т]+/i
  ];
  
  return similarityPatterns.some(pattern => pattern.test(normalized));
}

function generateBeepString(length: number): string {
  if (length < 2) return 'б';
  return 'б' + 'е'.repeat(length - 1);
}

export function filterProfanity(text: string): string {
  if (!text || typeof text !== 'string') return text;
  
  const words = text.split(/(\s+)/);
  
  return words.map(word => {
    if (/^\s+$/.test(word)) return word;
    
    if (LeoProfanity.check(word) || isSimilarToBadWord(word)) {
      return generateBeepString(word.length);
    }
    
    return word;
  }).join('');
}

export function hasProfanity(text: string): boolean {
  if (!text || typeof text !== 'string') return false;
  
  if (LeoProfanity.check(text)) return true;
  
  const words = text.split(/\s+/);
  return words.some(word => isSimilarToBadWord(word));
}

export function addCustomBadWords(words: string[]): void {
  words.forEach(word => LeoProfanity.add(word));
}

export function removeWord(word: string): void {
  LeoProfanity.remove(word);
}