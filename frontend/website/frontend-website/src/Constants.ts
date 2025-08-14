export const convertCategoriesToIds = (categories: string[]): number[] => {
    const categoryMap: { [key: string]: number } = {
      'movies': 1,    // Фильмы
      'games': 2,     // Игры
      'board': 3,     // Настольные игры
      'other': 4      // Другое
    };
    
    return categories.map(category => categoryMap[category]).filter(id => id !== undefined);
};

export const convertCategRuToEng = (categoryIds: string[]): ('games' | 'movies' | 'board' | 'other')[] => {
    const idMap: { [key: string]: 'games' | 'movies' | 'board' | 'other' } = {
      'Фильмы': 'movies',    // Фильмы
      'Игры': 'games',     // Игры
      'Настольные игры': 'board',     // Настольные игры
      'Другое': 'other'      // Другое
    };
    
    return categoryIds.map(id => idMap[id]).filter(category => category !== undefined);
};

export const convertIdsToCategories = (categoryIds: number[]): ('games' | 'movies' | 'board' | 'other')[] => {
    const idMap: { [key: number]: 'games' | 'movies' | 'board' | 'other' } = {
      1: 'movies',    // Фильмы
      2: 'games',     // Игры
      3: 'board',     // Настольные игры
      4: 'other'      // Другое
    };
    
    return categoryIds.map(id => idMap[id]).filter(category => category !== undefined);
};

export const convertSocialContactsToString = (socialContacts: { name: string; link: string }[]): string => {
    if (!Array.isArray(socialContacts) || socialContacts.length === 0) {
        return '';
    }

    return socialContacts
        .filter(contact => contact.name && contact.link) // Фильтруем пустые контакты
        .map(contact => `${contact.name.trim()}:${contact.link.trim()}`) // Форматируем каждый контакт
        .join(', '); // Объединяем через запятую и пробел
};

export const getAccesToken = (): string | null => {
  return localStorage.getItem('access_token');
}