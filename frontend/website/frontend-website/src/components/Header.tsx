'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useState, useEffect, useRef } from 'react';
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
    const [showSearchMenu, setShowSearchMenu] = useState(false);
    const router = useRouter();
    const searchMenuRef = useRef<HTMLDivElement>(null);
    const searchMenuTimeoutRef = useRef<NodeJS.Timeout>();

    // Проверка наличия непрочитанных уведомлений
    const checkNotifications = async () => {
        if (!isLoggedIn) return;

        try {
            const accessToken = await getAccesToken(router);
            const hasNotifications = await checkNotif(accessToken);
            setHasUnreadNotifications(hasNotifications);
        } catch (error) {
            console.error('Ошибка проверки уведомлений:', error);
        }
    };

    // Проверка уведомлений при монтировании и каждые 10 минут
    useEffect(() => {
        if (isLoggedIn) {
            checkNotifications();
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

    // Закрытие выпадающего меню при клике вне его
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (searchMenuRef.current && !searchMenuRef.current.contains(event.target as Node)) {
                setShowSearchMenu(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, []);

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
    };

    const handleSearchClick = (type: 'events' | 'groups' | 'users') => {
        setShowSearchMenu(false);
        
        if (type === 'events') {
            router.push('/search/events?category=Все');
        } else if (type === 'groups') {
            router.push('/search?type=groups');
        } else {
            router.push('/search?type=users');
        }
    };

    const handleSearchMenuEnter = () => {
        if (searchMenuTimeoutRef.current) {
            clearTimeout(searchMenuTimeoutRef.current);
        }
        setShowSearchMenu(true);
    };

    const handleSearchMenuLeave = () => {
        searchMenuTimeoutRef.current = setTimeout(() => {
            setShowSearchMenu(false);
        }, 200);
    };

    return (
        <>
            <header className={styles.header}>
                <Link href="/" className={styles.logoSection}>
                    <Image src="/logo.png" alt="Логотип" width={64} height={64} className={styles.logoImage}/>
                    <span className={styles.logoText}>FriendShips</span>
                </Link>

                <nav className={styles.navSection}>
                    <Link href="/groups" className={styles.navLink}>Группы</Link>
                    <Link href="/news" className={styles.navLink}>Новости</Link>
                    
                    {isLoggedIn && (
                        <div 
                            className={styles.searchMenuContainer}
                            ref={searchMenuRef}
                            onMouseEnter={handleSearchMenuEnter}
                            onMouseLeave={handleSearchMenuLeave}
                        >
                            <span className={styles.navLink}>Поиск</span>
                            
                            {showSearchMenu && (
                                <div className={styles.searchDropdown}>
                                    <button 
                                        className={styles.searchDropdownItem}
                                        onClick={() => handleSearchClick('events')}
                                    >
                                        События
                                    </button>
                                    <button 
                                        className={styles.searchDropdownItem}
                                        onClick={() => handleSearchClick('groups')}
                                    >
                                        Группы
                                    </button>
                                    <button 
                                        className={styles.searchDropdownItem}
                                        onClick={() => handleSearchClick('users')}
                                    >
                                        Люди
                                    </button>
                                </div>
                            )}
                        </div>
                    )}
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