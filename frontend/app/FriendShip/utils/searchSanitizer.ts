export const sanitizeSearchQuery = (query: string): string => {
  if (!query || typeof query !== 'string') {
    return '';
  }

  let sanitized = query.trim();

  sanitized = sanitized.substring(0, 100);

  sanitized = sanitized.replace(/[;'"\\<>]/g, '');
  
  sanitized = sanitized.replace(/\s+/g, ' ');
  
  return sanitized;
};

export const sanitizeCityFilter = (city: string): string => {
  if (!city || typeof city !== 'string') {
    return '';
  }

  return city.trim()
    .substring(0, 50)
    .replace(/[^а-яА-ЯёЁa-zA-Z\s-]/g, '');
};

export const sanitizeUsername = (username: string): string => {
  if (!username || typeof username !== 'string') {
    return '';
  }
  
  return username.trim()
    .substring(0, 50)
    .replace(/[<>'"\\;]/g, '');
};