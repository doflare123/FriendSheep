'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useState, useEffect, useRef } from 'react';
import NotificationDropdown from './NotificationDropdown';
import EventModal from './Events/EventModal';
import { useAuth } from '../contexts/AuthContext';
import { useRouter } from 'next/navigation';
import { getNotif } from '@/api/notification/getNotif';
import { getAccesToken, getUserData } from '@/Constants';
import styles from '../styles/Header.module.css';

interface Notification {
    id: number;
    sendAt: string;
    sent: boolean;
    text: string;
    type: string;
    viewed: boolean;
}

interface Invite {
    createdAt: string;
    groupId: number;
    groupName: string;
    id: number;
    status: string;
}

interface NotificationResponse {
    invites: Invite[];
    notifications: Notification[];
}

export default function Header() {
    const { isLoggedIn, userData, logout, isLoading } = useAuth();
    const [showNotifications, setShowNotifications] = useState(false);
    const [hasUnreadNotifications, setHasUnreadNotifications] = useState(false);
    const [unreadCount, setUnreadCount] = useState(0);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const [showSearchMenu, setShowSearchMenu] = useState(false);
    const [hasTelegramLink, setHasTelegramLink] = useState(true);
    const router = useRouter();
    const searchMenuRef = useRef<HTMLDivElement>(null);
    const searchMenuTimeoutRef = useRef<NodeJS.Timeout>();
    const audioRef = useRef<HTMLAudioElement | null>(null);
    const originalTitleRef = useRef<string>('');

    // Инициализация звука и сохранение оригинального title
    useEffect(() => {
        audioRef.current = new Audio('/notification-sound.mp3');
        audioRef.current.volume = 0.5;
        originalTitleRef.current = document.title;
        
        return () => {
            if (audioRef.current) {
                audioRef.current.pause();
                audioRef.current = null;
            }
        };
    }, []);

    // Проверка привязки телеграма
    useEffect(() => {
        const checkTelegramLink = async () => {
            if (isLoggedIn) {
                const userData = await getUserData();
                setHasTelegramLink(userData?.telegram_link || false);
            }
        };

        checkTelegramLink();
    }, [isLoggedIn]);

    // Обновление title вкладки
    const updateBrowserTitle = (count: number) => {
        if (count > 0) {
            const countText = count === 1 
                ? '1 непрочитанное уведомление' 
                : `${count} непрочитанных уведомлений`;
            document.title = `${countText} • FriendShip`;
        } else {
            document.title = originalTitleRef.current;
        }
    };

    // Воспроизведение звука уведомления
    const playNotificationSound = () => {
        if (audioRef.current) {
            audioRef.current.play().catch(err => {
                console.log('Не удалось воспроизвести звук:', err);
            });
        }
    };

    // Проверка наличия непрочитанных уведомлений
    const checkNotifications = async () => {
        if (!isLoggedIn) return;

        try {
            const accessToken = await getAccesToken(router);
            const data: NotificationResponse = await getNotif(accessToken);
            
            // Подсчитываем непрочитанные уведомления и непринятые приглашения
            const unreadNotifications = data.notifications?.filter(n => !n.viewed).length || 0;
            const pendingInvites = data.invites?.filter(i => i.status === 'pending').length || 0;
            const totalUnread = unreadNotifications + pendingInvites;
            
            // Если количество непрочитанных увеличилось - новое уведомление
            if (totalUnread > unreadCount) {
                playNotificationSound();
                
                // Показываем системное уведомление (если есть разрешение)
                if ('Notification' in window && Notification.permission === 'granted') {
                    const newCount = totalUnread - unreadCount;
                    const notificationText = newCount === 1 
                        ? 'У вас новое уведомление!' 
                        : `У вас ${newCount} новых уведомлений!`;
                    
                    new Notification('FriendShip', {
                        body: notificationText,
                        icon: '/logo.png',
                        badge: '/logo.png'
                    });
                }
            }
            
            setUnreadCount(totalUnread);
            setHasUnreadNotifications(totalUnread > 0);
            updateBrowserTitle(totalUnread);
        } catch (error) {
            console.error('Ошибка проверки уведомлений:', error);
        }
    };

    // Запрос разрешения на системные уведомления
    useEffect(() => {
        if (isLoggedIn && 'Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }
    }, [isLoggedIn]);

    // Проверка уведомлений при монтировании и каждую минуту
    useEffect(() => {
        if (isLoggedIn) {
            checkNotifications();
            const interval = setInterval(checkNotifications, 60000);
            return () => clearInterval(interval);
        } else {
            // Сбрасываем индикаторы при выходе
            setUnreadCount(0);
            setHasUnreadNotifications(false);
            updateBrowserTitle(0);
        }
    }, [isLoggedIn]);

    // Обновление после открытия/закрытия дропдауна
    useEffect(() => {
        if (!showNotifications && isLoggedIn) {
            checkNotifications();
        }
    }, [showNotifications]);

    // Сброс индикаторов при возвращении на вкладку
    useEffect(() => {
        const handleVisibilityChange = () => {
            if (!document.hidden && isLoggedIn) {
                checkNotifications();
            }
        };

        document.addEventListener('visibilitychange', handleVisibilityChange);
        return () => {
            document.removeEventListener('visibilitychange', handleVisibilityChange);
        };
    }, [isLoggedIn]);

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
                    <span className={styles.logoText}>FriendShip</span>
                </Link>

                <nav className={styles.navSection}>
                    {isLoggedIn && (
                        <Link href="/groups" className={styles.navLink}>Группы</Link>
                    )}
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
                                    {!hasTelegramLink && (
                                        <div className={styles.telegramBadge}>
                                            <Image 
                                                src="/social/tg_grey.png"
                                                alt="Telegram не привязан"
                                                width={12}
                                                height={12}
                                            />
                                        </div>
                                    )}
                                </button>
                                
                                {showNotifications && (
                                    <NotificationDropdown 
                                        onClose={() => setShowNotifications(false)}
                                        hasTelegramLink={hasTelegramLink}
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