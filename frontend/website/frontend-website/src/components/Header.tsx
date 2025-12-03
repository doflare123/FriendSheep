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
    const audioRef = useRef<HTMLAudioElement | null>(null);
    const originalTitleRef = useRef<string>('');
    const faviconRef = useRef<HTMLLinkElement | null>(null);

    // Инициализация звука и сохранение оригинального title
    useEffect(() => {
        audioRef.current = new Audio('/notification-sound.mp3'); // Добавьте звуковой файл в public
        audioRef.current.volume = 0.5;
        originalTitleRef.current = document.title;
        
        // Находим favicon
        faviconRef.current = document.querySelector("link[rel*='icon']");
        
        return () => {
            if (audioRef.current) {
                audioRef.current.pause();
                audioRef.current = null;
            }
        };
    }, []);

    // Обновление favicon и title
    const updateBrowserIndicators = (hasNotifications: boolean, notificationCount: number = 1) => {
        if (hasNotifications) {
            // Обновляем title с текстом и числом уведомлений
            const countText = notificationCount > 1 ? `${notificationCount} новых сообщения` : '1 новое сообщение';
            document.title = `${countText} • FriendShip`;
            
            // Создаём favicon с красным badge
            if (faviconRef.current) {
                const canvas = document.createElement('canvas');
                canvas.width = 32;
                canvas.height = 32;
                const ctx = canvas.getContext('2d');
                
                if (ctx) {
                    const img = new Image();
                    img.onload = () => {
                        // Рисуем оригинальную иконку
                        ctx.drawImage(img, 0, 0, 32, 32);
                        
                        // Рисуем красный кружок с белой обводкой
                        ctx.strokeStyle = '#ffffff';
                        ctx.lineWidth = 2;
                        ctx.fillStyle = '#ff3b30';
                        ctx.beginPath();
                        ctx.arc(24, 8, 7, 0, 2 * Math.PI);
                        ctx.fill();
                        ctx.stroke();
                        
                        // Добавляем число внутри кружка (если <= 9)
                        if (notificationCount <= 9) {
                            ctx.fillStyle = '#ffffff';
                            ctx.font = 'bold 12px Arial';
                            ctx.textAlign = 'center';
                            ctx.textBaseline = 'middle';
                            ctx.fillText(notificationCount.toString(), 24, 8);
                        }
                        
                        faviconRef.current!.href = canvas.toDataURL('image/png');
                    };
                    img.src = '/favicon.ico';
                }
            }
        } else {
            // Возвращаем оригинальный title
            document.title = originalTitleRef.current;
            
            // Возвращаем оригинальный favicon
            if (faviconRef.current) {
                faviconRef.current.href = '/favicon.ico';
            }
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
            const hasNotifications = await checkNotif(accessToken);
            
            // Если появилось новое уведомление
            if (hasNotifications && !hasUnreadNotifications) {
                playNotificationSound();
                
                // Показываем системное уведомление (если есть разрешение)
                if ('Notification' in window && Notification.permission === 'granted') {
                    new Notification('FriendShip', {
                        body: 'У вас новое уведомление!',
                        icon: '/favicon.ico',
                        badge: '/favicon.ico'
                    });
                }
            }
            
            setHasUnreadNotifications(hasNotifications);
            // TODO: Если API возвращает количество уведомлений, передайте его вторым параметром
            // Например: updateBrowserIndicators(hasNotifications, notificationCount);
            updateBrowserIndicators(hasNotifications, 1);
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
            updateBrowserIndicators(false);
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