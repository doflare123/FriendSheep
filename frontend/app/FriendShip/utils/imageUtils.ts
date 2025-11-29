// eslint-disable-next-line import/no-unresolved
import { LOCAL_IP } from "@env";

const PLACEHOLDER_IMAGE = 'https://via.placeholder.com/150/CCCCCC/FFFFFF?text=User';

export const normalizeImageUrl = (imageUrl: string | null | undefined): string => {
  if (!imageUrl || imageUrl.trim() === '') {
    console.warn('[imageUtils] ⚠️ Пустой URL изображения - используем placeholder');
    return PLACEHOLDER_IMAGE;
  }
 
  if (imageUrl.includes('selcloud.ru') || imageUrl.includes('selstorage.ru')) {
    console.log('[imageUtils] ✅ URL Selectel - используем как есть:', imageUrl);
    return imageUrl;
  }

  if (imageUrl.includes('localhost:8080')) {
    const normalized = imageUrl.replace('http://localhost:8080', `http://${LOCAL_IP}:8080`);
    console.log('[imageUtils] 🔄 Нормализован localhost URL:', normalized);
    return normalized;
  }

  if (imageUrl.includes('127.0.0.1:8080')) {
    const normalized = imageUrl.replace('http://127.0.0.1:8080', `http://${LOCAL_IP}:8080`);
    console.log('[imageUtils] 🔄 Нормализован 127.0.0.1 URL:', normalized);
    return normalized;
  }

  console.log('[imageUtils] ℹ️ URL без изменений:', imageUrl);
  return imageUrl;
};