// eslint-disable-next-line import/no-unresolved
import { LOCAL_IP } from "@env";

export const normalizeImageUrl = (imageUrl: string | null | undefined): string => {
  if (!imageUrl) return '';
  
  if (imageUrl.includes('localhost')) {
    return imageUrl.replace('localhost', LOCAL_IP);
  }
  
  if (imageUrl.includes('127.0.0.1')) {
    return imageUrl.replace('127.0.0.1', LOCAL_IP);
  }
  
  return imageUrl;
};