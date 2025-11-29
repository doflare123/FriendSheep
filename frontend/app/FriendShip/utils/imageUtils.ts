const PLACEHOLDER_IMAGE = 'https://via.placeholder.com/150/CCCCCC/FFFFFF?text=User';

export const normalizeImageUrl = (imageUrl: string | null | undefined): string => {
  if (!imageUrl || imageUrl.trim() === '') {
    console.warn('[imageUtils] ⚠️ Пустой URL изображения - используем placeholder');
    return PLACEHOLDER_IMAGE;
  }

  console.log('[imageUtils] ✅ URL изображения:', imageUrl);
  return imageUrl;
};