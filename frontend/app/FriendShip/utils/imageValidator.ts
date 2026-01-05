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
      'selcloud.ru',
      'friendsheep.ru',
    ];
    
    const hostname = parsed.hostname;
    const isAllowed = allowedDomains.some(domain => 
      hostname.includes(domain)
    );
    
    return isAllowed;
  } catch (error) {
    return false;
  }
};