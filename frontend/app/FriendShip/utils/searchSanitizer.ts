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