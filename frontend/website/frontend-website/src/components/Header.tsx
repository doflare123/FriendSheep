'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import NotificationDropdown from './NotificationDropdown';
import EventModal from './Events/EventModal';
import { useAuth } from '../contexts/AuthContext';
import { useRouter } from 'next/navigation';
import { checkNotif } from '@/api/notification/checkNotif';
import { getAccesToken } from '@/Constants';
import styles from '../styles/Header.module.css';

export default function Header() {
    const { isLoggedIn, userData, logout, isLoading } = useAuth();
    const [showNotifications, setShowNotifications] = useState(false);
    const [hasUnreadNotifications, setHasUnreadNotifications] = useState(false);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
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
            <header className={styles.header}>
                <Link href="/" className={styles.logoSection}>
                    <Image src="/logo.png" alt="Логотип" width={64} height={64} className={styles.logoImage}/>
                    <span className={styles.logoText}>FriendShip</span>
                </Link>
                <nav className={styles.navSection}>
                    <Link href="/news" className={styles.navLink}>Новости</Link>
                    <Link href="/groups" className={styles.navLink}>Группы</Link>
                </nav>
                <div className={styles.authSection}>
                    <div className={styles.loadingSkeleton}>Загрузка...</div>
                </div>
            </header>
        );
    }

    const handleAvatarClick = () => {
        router.push('/profile/' + userData?.us);
    };

    const handleCreateClick = () => {
        setIsCreateModalOpen(true);
    };

    const handleModalClose = () => {
        setIsCreateModalOpen(false);
    };

    const handleEventSave = () => {
        setIsCreateModalOpen(false);
        // Перезагрузка происходит внутри EventModal
    };

    return (
        <>
            <header className={styles.header}>
                <Link href="/" className={styles.logoSection}>
                    <Image src="/logo.png" alt="Логотип" width={64} height={64} className={styles.logoImage}/>
                    <span className={styles.logoText}>FriendShip</span>
                </Link>

                <nav className={styles.navSection}>
                    <Link href="/news" className={styles.navLink}>Новости</Link>
                    <Link href="/groups" className={styles.navLink}>Группы</Link>
                </nav>

                <div className={styles.authSection}>
                    {isLoggedIn ? (
                        <>
                            <button className={styles.createButton} onClick={handleCreateClick}>
                                Создать
                            </button>
                            
                            <div className={styles.notificationContainer}>
                                <button 
                                    className={styles.notificationButton}
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
                            
                            <span className={styles.username}>{userData?.username}</span>
                            
                            <div 
                                className={styles.avatarContainer} 
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
                                    className={styles.avatar}
                                />
                            </div>
                        </>
                    ) : (
                        <>
                            <Link href="/login" className={styles.authLink}>Войти</Link>
                            <span className={styles.authSeparator}>/</span>
                            <Link href="/register" className={styles.authLink}>Регистрация</Link>
                        </>
                    )}
                </div>
            </header>

            <EventModal
                isOpen={isCreateModalOpen}
                onClose={handleModalClose}
                onSave={handleEventSave}
                mode="create"
            />
        </>
    );
}