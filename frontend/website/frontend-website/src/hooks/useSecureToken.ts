import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { isTokenValid } from '@/api/auth';

/**
 * Хук для безопасного получения access токена с автоматической проверкой и обновлением
 * 
 * @returns Валидный access токен или пустую строку (если идёт редирект на /login)
 * 
 * @example
 * const MyComponent = () => {
 *   const token = useSecureToken();
 *   
 *   useEffect(() => {
 *     if (token) {
 *       // делаем запрос с токеном
 *       fetchData(token);
 *     }
 *   }, [token]);
 * }
 */
export const useSecureToken = (): string => {
  const router = useRouter();
  const { forceRefreshToken, logout } = useAuth();
  const [token, setToken] = useState<string>('');

  useEffect(() => {
    const checkToken = async () => {
      const accessToken = localStorage.getItem('access_token') || '';

      // Если токена нет
      if (!accessToken) {
        console.warn('⚠️ Access токен отсутствует, редирект на /login');
        logout();
        console.log("LOGIN7");
        router.push('/login');
        return;
      }

      // Если токен валиден
      if (isTokenValid(accessToken)) {
        setToken(accessToken);
        return;
      }

      // Токен невалиден - пробуем обновить через refresh token
      console.log('🔄 Access токен невалиден, попытка обновления...');
      const refreshSuccess = await forceRefreshToken();

      if (refreshSuccess) {
        const newToken = localStorage.getItem('access_token') || '';
        setToken(newToken);
        console.log('✅ Токен успешно обновлён');
      } else {
        console.warn('❌ Не удалось обновить токен, редирект на /login');
        logout();
        console.log("LOGIN8");
        router.push('/login');
      }
    };

    checkToken();
  }, [router, forceRefreshToken, logout]);

  return token;
};