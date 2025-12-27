import * as filter from 'leo-profanity';

console.log('[ProfanityFilter] Инициализация...');

try {
  filter.clearList();
  
  filter.loadDictionary('ru');
  console.log('[ProfanityFilter] ✅ Русский словарь загружен');
  
  filter.loadDictionary('en');
  console.log('[ProfanityFilter] ✅ Английский словарь загружен');
  
} catch (error) {
  console.error('[ProfanityFilter] ❌ Ошибка загрузки словарей:', error);
}

const russianBadWords = [
  'хуй', 'хуя', 'хуе', 'хую', 'хуем', 'хуи', 'хуёв', 'хер', 'хера',
  'пизда', 'пизде', 'пизду', 'пиздой', 'пизд', 'пиздец', 'пиздюк',
  'ебать', 'ебал', 'ебет', 'ебёт', 'ебут', 'ебала', 'ебали', 'ебаный', 'ебучий', 'ебля',
  'бля', 'блять', 'блядь', 'блядский', 'блядина', 'блядство',
  'мудак', 'мудила', 'мудачье', 'мудачина', 'мудень', 'мудня',
  'сука', 'суки', 'сучка', 'сучий', 'сукин', 'сучье',
  'гавно', 'говно', 'говна', 'говну', 'говном', 'говнюк',
  'жопа', 'жопы', 'жопе', 'жопу', 'жопой', 'жополиз',
  'срать', 'срал', 'срет', 'срёт', 'срала', 'срали', 'сраный',
  'ссать', 'ссал', 'ссышь', 'ссыт', 'ссака',
  'долбоёб', 'долбаёб', 'уёбок', 'ублюдок', 'придурок',
  'дебил', 'дебилка', 'дебильный', 'идиот', 'идиотка', 'тупица',
  'чмо', 'чмошник', 'чмырь', 'чурка', 'шлюха', 'шлюшка',
  'педик', 'педрила', 'пидор', 'пидорас', 'пидорасы',
  'фуй', 'хрен', 'херня', 'хренов', 'херовый',
  'нахуй', 'нахер', 'похуй', 'захуярить', 'хуярить', 'охуеть', 'охренеть',
  'ебнуть', 'ебануть', 'поебать', 'заебать', 'доебаться', 'уебать',
  'пиздануть', 'запиздить', 'попизде', 'впизду',
  'блядун', 'бляха', 'ёб', 'ёбнуть', 'ёбаный', 'ёбнутый',
  'говнище', 'говнюк', 'дерьмо', 'дерьма', 'дермо',
];

filter.add(russianBadWords);
console.log('[ProfanityFilter] ✅ Добавлено вручную:', russianBadWords.length, 'слов');

const testWord = 'хуй';
const isDetected = filter.check(testWord);
console.log(`[ProfanityFilter] Тест "${testWord}":`, isDetected ? '✅ работает' : '❌ НЕ работает');

const checkProfanityPatterns = (text: string): string[] => {
  if (!text) return [];
  
  const foundWords: string[] = [];
  const words = text.toLowerCase().split(/\s+/);
  
  const patterns = [
    { regex: /х+у+[йяеюёивн]/i, description: 'ху-корень' },

    { regex: /п+и+з+д/i, description: 'пизд-корень' },

    { regex: /[знпд]?а?е+б/i, description: 'еб-корень' },
    
    { regex: /б+л+я/i, description: 'бля-корень' },
    
    { regex: /п+о+х+у/i, description: 'поху-корень' },
    { regex: /п+о+х+е+р/i, description: 'похер' },

    { regex: /п+[ие]+д+[аор]/i, description: 'пид-корень' },
    
    { regex: /х+е+р/i, description: 'хер-корень' },
    
    { regex: /с+у+к/i, description: 'сук-корень' },

    { regex: /м+у+д+[ак]/i, description: 'муд-корень' },

    { regex: /г+[ао]+в+н/i, description: 'говн-корень' },

    { regex: /ж+о+п/i, description: 'жоп-корень' },

    { regex: /н+а+х+[уе]/i, description: 'нах-корень' },

    { regex: /з+а+е+б/i, description: 'заеб-корень' },

    { regex: /о+х+у+е/i, description: 'охуе-корень' },

    { regex: /[её]+б/i, description: 'ёб-корень' },

    { regex: /у+[её]+б/i, description: 'уёб-корень' },

    { regex: /д+о+л+б+[ао]/i, description: 'долб*' },

    { regex: /с+р+а/i, description: 'сра-корень' },

    { regex: /с+с+а/i, description: 'сса-корень' },

    { regex: /ш+л+ю+[хш]/i, description: 'шлюх-корень' },
  ];
  
  words.forEach(word => {
    const cleanWord = word.replace(/[^а-яё]/gi, '').trim();

    if (cleanWord.length < 3) return;

    for (const pattern of patterns) {
      if (pattern.regex.test(cleanWord)) {
        foundWords.push(cleanWord);
        console.log(`[Pattern] ✅ Найдено: "${cleanWord}" по паттерну "${pattern.description}"`);
        break;
      }
    }
  });
  
  return [...new Set(foundWords)];
};

const customCensor = (text: string): string => {
  if (!text) return text;

  const badWordsFromLibrary = filter.list(text);
  
  const badWordsFromPatterns = checkProfanityPatterns(text);

  const allBadWords = [...new Set([...badWordsFromLibrary, ...badWordsFromPatterns])];
  
  if (allBadWords.length === 0) return text;

  let cleanedText = text;

  allBadWords.forEach((badWord: string) => {
    const wordLength = badWord.length;
    const replacement = 'б' + 'е'.repeat(Math.max(0, wordLength - 1));

    const escapedWord = badWord.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    
    const regex = new RegExp(`(^|[\\s.,!?;:()\"'])${escapedWord}([\\s.,!?;:()\"']|$)`, 'gi');
    
    cleanedText = cleanedText.replace(regex, (match, prefix, suffix) => {
      return prefix + replacement + suffix;
    });

    const simpleRegex = new RegExp(`^${escapedWord}$`, 'gi');
    cleanedText = cleanedText.replace(simpleRegex, replacement);
  });

  return cleanedText;
};

export const profanityFilter = {

  check: (text: string): boolean => {
    if (!text) return false;
    
    console.log('[ProfanityFilter] check() вход:', text);

    const hasInLibrary = filter.check(text);
    console.log('[ProfanityFilter] библиотека нашла:', hasInLibrary);
    
    if (hasInLibrary) return true;

    const foundByPatterns = checkProfanityPatterns(text);
    console.log('[ProfanityFilter] паттерны нашли:', foundByPatterns);
    
    const result = foundByPatterns.length > 0;
    console.log('[ProfanityFilter] check() результат:', result);
    
    return result;
  },

  clean: (text: string): string => {
    if (!text) return text;
    
    console.log('[ProfanityFilter] clean() вход:', text);
    
    const badWords = profanityFilter.list(text);
    console.log('[ProfanityFilter] найдено слов:', badWords);
    
    const result = customCensor(text);
    console.log('[ProfanityFilter] clean() выход:', result);
    
    return result;
  },

  list: (text: string): string[] => {
    if (!text) return [];
    const fromLibrary = filter.list(text) as string[];
    const fromPatterns = checkProfanityPatterns(text);
    return [...new Set([...fromLibrary, ...fromPatterns])];
  },

  addWords: (...words: string[]): void => {
    filter.add(words);
    console.log('[ProfanityFilter] ➕ Добавлено слов:', words.length);
  },

  removeWords: (...words: string[]): void => {
    filter.remove(words);
  },

  clearList: (): void => {
    filter.clearList();
  },

  validate: (text: string): { isValid: boolean; message?: string; badWords?: string[] } => {
    if (!text) {
      return { isValid: true };
    }

    const hasProfanity = profanityFilter.check(text);
    
    if (hasProfanity) {
      const badWords = profanityFilter.list(text);
      return {
        isValid: false,
        message: 'Текст содержит недопустимые слова',
        badWords,
      };
    }

    return { isValid: true };
  },

  autoClean: (text: string): string => {
    return customCensor(text);
  },
};

export default profanityFilter;