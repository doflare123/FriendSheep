export const sanitizeUrl = (url: string): string => {
  const trimmed = url.trim();
  
  if (trimmed.toLowerCase().startsWith('javascript:')) {
    throw new Error('Недопустимый URL');
  }
  
  const allowedProtocols = ['http://', 'https://', 'mailto:', 'tel:'];
  const hasAllowedProtocol = allowedProtocols.some(protocol => 
    trimmed.toLowerCase().startsWith(protocol)
  );
  
  if (!hasAllowedProtocol && !trimmed.includes('://')) {
    return `https://${trimmed}`;
  }
  
  return trimmed;
};

export const sanitizeEmail = (email: string): string => {
  return email.trim().toLowerCase();
};

export const sanitizeText = (text: string, maxLength: number): string => {
  return text.trim().substring(0, maxLength);
};