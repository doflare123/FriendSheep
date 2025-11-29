// eslint-disable-next-line import/no-unresolved
import { LOCAL_IP } from '@env';

export const isValidImageUrl = (url: string | undefined | null): boolean => {
  if (!url || typeof url !== 'string') {
    return false;
  }

  try {
    const parsed = new URL(url);

    if (!['https:', 'http:'].includes(parsed.protocol)) {
      return false;
    }

    const allowedDomains = [
      'selstorage.ru',
      'localhost',
      LOCAL_IP,
    ];
    
    const hostname = parsed.hostname;
    const isAllowed = allowedDomains.some(domain => 
      domain && hostname.includes(domain)
    );
    
    return isAllowed;
  } catch (error) {
    return false;
  }
};

export const getValidImageUrl = (
  url: string | undefined | null, 
  fallback: string = 'https://via.placeholder.com/150'
): string => {
  if (isValidImageUrl(url)) {
    return url as string;
  }
  return fallback;
};