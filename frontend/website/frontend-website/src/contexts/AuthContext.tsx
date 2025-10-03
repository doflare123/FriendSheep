'use client';

import React, { createContext, useContext, useState, useEffect, useRef } from 'react';
import { refreshAccessToken, isTokenValid, getCookie, setCookie, deleteCookie } from '../api/auth';
import {decodeJWT} from '@/Constants'

interface UserData {
  us: string;
  username: string;
  email: string;
  avatar?: string;
}

interface AuthContextType {
  isLoggedIn: boolean;
  userData: UserData | null;
  login: (accessToken: string, refreshToken: string) => void;
  logout: () => void;
  isLoading: boolean;
  forceRefreshToken: () => Promise<boolean>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [userData, setUserData] = useState<UserData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const refreshIntervalRef = useRef<NodeJS.Timeout | null>(null);
  
  // Обновление пользовательских данных из токена
  const updateUserData = (token: string) => {
    const decoded = decodeJWT(token);
    if (decoded) {
      setUserData({
        us: decoded.Us,
        username: decoded.Username || 'Пользователь',
        email: decoded.Email || '',
        avatar: decoded.Image
      });
      setIsLoggedIn(true);
    }
  };

  // Проверка и обновление токена
  const checkAndRefreshToken = async (): Promise<boolean> => {
    try {
      const accessToken = localStorage.getItem('access_token');
      
      // Если есть валидный access токен
      if (accessToken && isTokenValid(accessToken)) {
        updateUserData(accessToken);
        return true;
      }

      // Если access токен отсутствует или невалиден, попытаемся обновить
      const refreshToken = getCookie('refresh_token');
      
      if (refreshToken) {
        const tokens = await refreshAccessToken(refreshToken);
        localStorage.setItem('access_token', tokens.access_token);
        setCookie('refresh_token', tokens.refresh_token, 7); // 7 дней
        updateUserData(tokens.access_token);
        return true;
      } else {
        // Refresh токен отсутствует
        logout();
        return false;
      }
    } catch (error) {
      console.error('Ошибка при проверке токена:', error);
      logout();
      return false;
    }
  };

  // Настройка автоматического обновления
  const setupTokenRefresh = () => {
    const scheduleNextRefresh = () => {
      if (refreshIntervalRef.current) {
        clearTimeout(refreshIntervalRef.current);
      }
      
      const accessToken = localStorage.getItem('access_token');
      if (!accessToken) return;

      const decoded = decodeJWT(accessToken);
      if (!decoded || !decoded.exp) return;

      const expiryTime = decoded.exp * 1000;
      const currentTime = Date.now();
      const timeUntilExpiry = expiryTime - currentTime;
      
      // Обновляем за 2 минуты до истечения (токен живет 10 минут)
      const refreshTime = Math.max(timeUntilExpiry - 2 * 60 * 1000, 30 * 1000);

      refreshIntervalRef.current = setTimeout(async () => {
        console.log('Автоматическое обновление токена');
        const success = await checkAndRefreshToken();
        if (success) {
          scheduleNextRefresh(); // Планируем следующее обновление
        }
      }, refreshTime);

      console.log(`Следующее обновление токена через ${Math.round(refreshTime / 1000)} секунд`);
    };

    scheduleNextRefresh();
  };

  // Инициализация при загрузке
  useEffect(() => {
    const initAuth = async () => {
      setIsLoading(true);
      const success = await checkAndRefreshToken();
      if (success) {
        setupTokenRefresh();
      }
      setIsLoading(false);
    };

    initAuth();

    return () => {
      if (refreshIntervalRef.current) {
        clearTimeout(refreshIntervalRef.current);
      }
    };
  }, []);

  const login = (accessToken: string, refreshToken: string) => {
    localStorage.setItem('access_token', accessToken);
    setCookie('refresh_token', refreshToken, 7); // 7 дней
    updateUserData(accessToken);
    setupTokenRefresh();
  };

  const logout = () => {
    localStorage.removeItem('access_token');
    deleteCookie('refresh_token');
    setIsLoggedIn(false);
    setUserData(null);
    if (refreshIntervalRef.current) {
      clearTimeout(refreshIntervalRef.current);
    }
  };

  const forceRefreshToken = async (): Promise<boolean> => {
    try {
      const refreshToken = getCookie("refresh_token");

      if (!refreshToken) {
        logout();
        return false;
      }

      const tokens = await refreshAccessToken(refreshToken);
      localStorage.setItem("access_token", tokens.access_token);
      setCookie("refresh_token", tokens.refresh_token, 7); // 7 дней
      updateUserData(tokens.access_token);

      // обновляем расписание автообновления
      setupTokenRefresh();

      console.log("🔄 Токен обновлён принудительно");
      return true;
    } catch (error) {
      console.error("❌ Ошибка при принудительном обновлении токена:", error);
      logout();
      return false;
    }
  };


  const value = {
    isLoggedIn,
    userData,
    login,
    logout,
    isLoading,
    forceRefreshToken,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}
