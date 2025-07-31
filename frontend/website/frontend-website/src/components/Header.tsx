'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import NotificationDropdown from './NotificationDropdown';
import { refreshAccessToken, decodeJWT } from '../api/auth';
import '../styles/Header.css';

interface UserData {
  username: string;
  email: string;
  avatar?: string;
}

export default function Header() {
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [userData, setUserData] = useState<UserData | null>(null);
    const [showNotifications, setShowNotifications] = useState(false);

    useEffect(() => {
        checkAuthStatus();
        
        // Попытка обновить токен при загрузке страницы
        handleInitialTokenRefresh();
        
        // Настройка автоматического обновления токена каждые 15 минут
        const tokenRefreshInterval = setInterval(() => {
            handleTokenRefresh();
        }, 15 * 60 * 1000);

        // Слушаем изменения аутентификации
        const handleAuthChange = () => {
            checkAuthStatus();
        };
        
        window.addEventListener('auth-change', handleAuthChange);

        return () => {
            clearInterval(tokenRefreshInterval);
            window.removeEventListener('auth-change', handleAuthChange);
        };
    }, []);

    const checkAuthStatus = () => {
        const accessToken = localStorage.getItem('access_token');
        if (accessToken) {
            const decoded = decodeJWT(accessToken);
            if (decoded && decoded.exp * 1000 > Date.now()) {
                console.log(decoded)
                setIsLoggedIn(true);
                setUserData({
                    username: decoded.Us || 'Пользователь',
                    email: decoded.Email || '',
                    avatar: decoded.Image
                });
            } else {
                // Токен истек
                localStorage.removeItem('access_token');
                setIsLoggedIn(false);
                setUserData(null);
            }
        }
    };

    const handleInitialTokenRefresh = async () => {
        const accessToken = localStorage.getItem('access_token');
        const refreshToken = getRefreshToken();
        
        if (!accessToken && refreshToken) {
            // Нет access токена, но есть refresh - пытаемся обновить
            await handleTokenRefresh();
        }
    };

    const handleTokenRefresh = async () => {
        const refreshToken = getRefreshToken();
        if (refreshToken) {
            try {
                const tokens = await refreshAccessToken(refreshToken);
                localStorage.setItem('access_token', tokens.access_token);
                saveRefreshToken(tokens.refresh_token);
                checkAuthStatus();
            } catch (error) {
                console.error('Ошибка обновления токена:', error);
                handleLogout();
            }
        }
    };

    const getRefreshToken = (): string | null => {
        // Можно использовать localStorage или cookies
        return localStorage.getItem('refresh_token');
        // Или для cookies: return getCookie('refresh_token');
    };

    const saveRefreshToken = (token: string) => {
        // Сохранение в localStorage
        localStorage.setItem('refresh_token', token);
        
        // Или сохранение в cookies (более безопасно)
        // document.cookie = `refresh_token=${token}; path=/; max-age=${7 * 24 * 60 * 60}; secure; samesite=strict`;
    };

    const handleLogout = () => {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        // document.cookie = 'refresh_token=; path=/; expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        setIsLoggedIn(false);
        setUserData(null);
    };

    return (
        <header className="header">
            {/* Логотип слева */}
            <div className="logo-section">
                <Image src="/logo.png" alt="Логотип" width={64} height={64} className="logo-image"/>
                <span className="logo-text">FriendShip</span>
            </div>

            {/* Центрированные ссылки */}
            <nav className="nav-section">
                <Link href="/" className="nav-link">
                    Главная
                </Link>
                <Link href="#" className="nav-link">
                    Новости
                </Link>
                <Link href="/groups" className="nav-link">
                    Группы
                </Link>
            </nav>

            {/* Правая секция */}
            <div className="auth-section">
                {isLoggedIn ? (
                    <>
                        {/* Кнопка "Создать" */}
                        <button className="create-button">
                            Создать
                        </button>
                        
                        {/* Уведомления */}
                        <div className="notification-container">
                            <button 
                                className="notification-button"
                                onClick={() => setShowNotifications(!showNotifications)}
                            >
                                <Image 
                                    src="/notification-icon.png" 
                                    alt="Уведомления" 
                                    width={24} 
                                    height={24}
                                />
                                {/* Красная точка для новых уведомлений */}
                                <span className="notification-badge"></span>
                            </button>
                            
                            {showNotifications && (
                                <NotificationDropdown 
                                    onClose={() => setShowNotifications(false)}
                                />
                            )}
                        </div>
                        
                        {/* Имя пользователя */}
                        <span className="username">
                            {userData?.username}
                        </span>
                        
                        {/* Аватар */}
                        <div className="avatar-container">
                            <Image 
                                src={userData?.avatar || '/default-avatar.png'} 
                                alt="Аватар" 
                                width={60} 
                                height={60}
                                className="avatar"
                            />
                        </div>
                    </>
                ) : (
                    <>
                        <Link href="/login" className="auth-link">
                            Войти
                        </Link>
                        <span className="auth-separator">/</span>
                        <Link href="/register" className="auth-link">
                            Регистрация
                        </Link>
                    </>
                )}
            </div>
        </header>
    );
}