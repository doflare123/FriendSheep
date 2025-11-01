'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import NotificationDropdown from './NotificationDropdown';
import { useAuth } from '../contexts/AuthContext';
import { useRouter } from 'next/navigation';
import { checkNotif } from '@/api/notification/checkNotif';
import { getAccesToken } from '@/Constants';
import '../styles/Header.css';

export default function Header() {
    const { isLoggedIn, userData, logout, isLoading } = useAuth();
    const [showNotifications, setShowNotifications] = useState(false);
    const [hasUnreadNotifications, setHasUnreadNotifications] = useState(false);
    const router = useRouter();

    // Проверка наличия непрочитанных уведомлений
    const checkNotifications = async () => {
        if (!isLoggedIn) return;

        try {
            const accessToken = getAccesToken();
            const hasNotifications = await checkNotif(accessToken);
            setHasUnreadNotifications(hasNotifications);
        } catch (error) {
            console.error('Ошибка проверки уведомлений:', error);
        }
    };

    // Проверка уведомлений при монтировании и каждые 10 минут
    useEffect(() => {
        if (isLoggedIn) {
            checkNotifications(); // Сразу при загрузке

            // Интервал каждые 10 минут (600000 мс)
            const interval = setInterval(checkNotifications, 60000);

            return () => clearInterval(interval);
        }
    }, [isLoggedIn]);

    // Обновление после открытия/закрытия дропдауна
    useEffect(() => {
        if (!showNotifications && isLoggedIn) {
            checkNotifications();
        }
    }, [showNotifications]);

    // Показываем скелетон во время загрузки
    if (isLoading) {
        return (
            <header className="header">
                <div className="logo-section">
                    <Image src="/logo.png" alt="Логотип" width={64} height={64} className="logo-image"/>
                    <span className="logo-text">FriendShip</span>
                </div>
                <nav className="nav-section">
                    <Link href="/" className="nav-link">Главная</Link>
                    <Link href="/news" className="nav-link">Новости</Link>
                    <Link href="/groups" className="nav-link">Группы</Link>
                </nav>
                <div className="auth-section">
                    <div className="loading-skeleton">Загрузка...</div>
                </div>
            </header>
        );
    }

    const handleAvatarClick = () => {
        router.push('/profile/' + userData?.us);
    };

    return (
        <header className="header">
            <div className="logo-section">
                <Image src="/logo.png" alt="Логотип" width={64} height={64} className="logo-image"/>
                <span className="logo-text">FriendShip</span>
            </div>

            <nav className="nav-section">
                <Link href="/" className="nav-link">Главная</Link>
                <Link href="/news" className="nav-link">Новости</Link>
                <Link href="/groups" className="nav-link">Группы</Link>
            </nav>

            <div className="auth-section">
                {isLoggedIn ? (
                    <>
                        <button className="create-button">Создать</button>
                        
                        <div className="notification-container">
                            <button 
                                className="notification-button"
                                onClick={() => setShowNotifications(!showNotifications)}
                            >
                                <Image 
                                    src={hasUnreadNotifications ? "/notification_icon_active.png" : "/notification_icon.png"}
                                    alt="Уведомления" 
                                    width={24} 
                                    height={24}
                                />
                            </button>
                            
                            {showNotifications && (
                                <NotificationDropdown 
                                    onClose={() => setShowNotifications(false)}
                                />
                            )}
                        </div>
                        
                        <span className="username">{userData?.username}</span>
                        
                        <div 
                            className="avatar-container" 
                            onClick={handleAvatarClick}
                            role="button"
                            tabIndex={0}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                    handleAvatarClick();
                                }
                            }}
                        >
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
                        <Link href="/login" className="auth-link">Войти</Link>
                        <span className="auth-separator">/</span>
                        <Link href="/register" className="auth-link">Регистрация</Link>
                    </>
                )}
            </div>
        </header>
    );
}