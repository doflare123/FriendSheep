'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useState } from 'react';
import NotificationDropdown from './NotificationDropdown';
import { useAuth } from '../contexts/AuthContext';
import { useRouter } from 'next/navigation';
import '../styles/Header.css';

export default function Header() {
    const { isLoggedIn, userData, logout, isLoading } = useAuth();
    const [showNotifications, setShowNotifications] = useState(false);
    const router = useRouter();

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
                                    src="/notification-icon.png" 
                                    alt="Уведомления" 
                                    width={24} 
                                    height={24}
                                />
                                <span className="notification-badge"></span>
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