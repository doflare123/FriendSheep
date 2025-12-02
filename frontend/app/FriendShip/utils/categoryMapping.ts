export const categoryDisplayNames: Record<string, string> = {
  'Игры': 'Видеоигры',
  'Фильмы': 'Медиа',
  'Настолки': 'Настолки',
  'Другое': 'Другое',
  'Все': 'Все'
};

export const categoryInternalNames: Record<string, string> = {
  'Видеоигры': 'Игры',
  'Медиа': 'Фильмы',
  'Настолки': 'Настолки',
  'Другое': 'Другое',
  'Все': 'Все'
};

export const getCategoryDisplayName = (internalName: string): string => {
  return categoryDisplayNames[internalName] || internalName;
};

export const getCategoryInternalName = (displayName: string): string => {
  return categoryInternalNames[displayName] || displayName;
};