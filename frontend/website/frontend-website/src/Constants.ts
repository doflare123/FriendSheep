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

export const getCategoryIcon = (category: string): string => {
  const lowerCategory = category.toLowerCase();
  
  if (lowerCategory.includes('games') || lowerCategory =='игры') {
    return '/events/games.png';
  } else if (lowerCategory.includes('movies') || lowerCategory.includes('фильмы')) {
    return '/events/movies.png';
  } else if (lowerCategory.includes('boards') || lowerCategory.includes('настолки') || lowerCategory.includes('настольные игры')) {
    return '/events/board.png';
  } else {
    return '/events/other.png'; // default
  }
};

export const getSocialIcon = (name: string, link?: string, ): string => {
  const lowerLink = (link|| " ").toLowerCase();
  const lowerName = name.toLowerCase();
  
  if (lowerLink.includes('discord') || lowerName.includes('discord') || lowerName.includes('ds')) {
    return '/social/ds.png';
  } else if (lowerLink.includes('t.me') || lowerLink.includes('telegram') || lowerName.includes('telegram') || lowerName.includes('tg')) {
    return '/social/tg.png';
  } else if (lowerLink.includes('vk.com') || lowerName.includes('вконтакте') || lowerName.includes('vk')) {
    return '/social/vk.png';
  } else if (lowerLink.includes('wa.me') || lowerLink.includes('whatsapp') || lowerName.includes('whatsapp') || lowerName.includes('wa')) {
    return '/social/wa.png';
  } else if (lowerLink.includes('snapchat') || lowerName.includes('snapchat') || lowerName.includes('snap')) {
    return '/social/snap.png';
  } else {
    return '/default/soc_net.png';
  }
};