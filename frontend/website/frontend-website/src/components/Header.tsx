'use client';

import Image from 'next/image';
import Link from 'next/link';
import { useState, useEffect, useRef } from 'react';
import NotificationDropdown from './NotificationDropdown';
import EventModal from './Events/EventModal';
import LoadingIndicator from './LoadingIndicator';
import { useAuth } from '../contexts/AuthContext';
import { useRouter } from 'next/navigation';
import { getNotif } from '@/api/notification/getNotif';
import { getAccesToken, getUserData, MOBILE_APP_URL } from '@/Constants';
import { getOwnGroups } from '@/api/get_owngroups';
import { showNotification } from '@/utils';
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

import { USER_DATA_UPDATED_EVENT } from '@/Constants';

export default function Header() {
    const { isLoggedIn, userData, logout, isLoading } = useAuth();
    const [showNotifications, setShowNotifications] = useState(false);
    const [hasUnreadNotifications, setHasUnreadNotifications] = useState(false);
    const [unreadCount, setUnreadCount] = useState(0);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const [showSearchMenu, setShowSearchMenu] = useState(false);
    const [showProfileMenu, setShowProfileMenu] = useState(false);
    const [showLogoutModal, setShowLogoutModal] = useState(false);
    const [hasTelegramLink, setHasTelegramLink] = useState(true);
    const [isMobile, setIsMobile] = useState(false);
    const [isCheckingGroups, setIsCheckingGroups] = useState(false);
    const router = useRouter();
    const searchMenuRef = useRef<HTMLDivElement>(null);
    const profileMenuRef = useRef<HTMLDivElement>(null);
    const searchMenuTimeoutRef = useRef<NodeJS.Timeout>();
    const audioRef = useRef<HTMLAudioElement | null>(null);
    const originalTitleRef = useRef<string>('');

    useEffect(() => {
        const checkMobile = () => {
            setIsMobile(/Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent));
        };
        
        checkMobile();
        window.addEventListener('resize', checkMobile);
        return () => window.removeEventListener('resize', checkMobile);
    }, []);

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

    useEffect(() => {
        const checkTelegramLink = async () => {
            if (isLoggedIn) {
                const userData = await getUserData();
                setHasTelegramLink(userData?.telegram_link || false);
            }
        };

        checkTelegramLink();

        const handleUserDataUpdate = (event: CustomEvent) => {
            const updatedUserData = event.detail;
            setHasTelegramLink(updatedUserData?.telegram_link || false);
        };

        window.addEventListener(USER_DATA_UPDATED_EVENT, handleUserDataUpdate as EventListener);

        return () => {
            window.removeEventListener(USER_DATA_UPDATED_EVENT, handleUserDataUpdate as EventListener);
        };
    }, [isLoggedIn]);

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

    const playNotificationSound = () => {
        if (audioRef.current) {
            audioRef.current.play().catch(err => {
                console.log('Не удалось воспроизвести звук:', err);
            });
        }
    };

    const checkNotifications = async () => {
        if (!isLoggedIn) return;

        try {
            const accessToken = await getAccesToken(router);
            const data: NotificationResponse = await getNotif(accessToken);
            
            const unreadNotifications = data.notifications?.filter(n => !n.viewed).length || 0;
            const pendingInvites = data.invites?.filter(i => i.status === 'pending').length || 0;
            const totalUnread = unreadNotifications + pendingInvites;
            
            if (totalUnread > unreadCount) {
                playNotificationSound();
                
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

    useEffect(() => {
        if (isLoggedIn && 'Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }
    }, [isLoggedIn]);

    useEffect(() => {
        if (isLoggedIn) {
            checkNotifications();
            const interval = setInterval(checkNotifications, 60000);
            return () => clearInterval(interval);
        } else {
            setUnreadCount(0);
            setHasUnreadNotifications(false);
            updateBrowserTitle(0);
        }
    }, [isLoggedIn]);

    useEffect(() => {
        if (!showNotifications && isLoggedIn) {
            checkNotifications();
        }
    }, [showNotifications]);

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

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (searchMenuRef.current && !searchMenuRef.current.contains(event.target as Node)) {
                setShowSearchMenu(false);
            }
            if (profileMenuRef.current && !profileMenuRef.current.contains(event.target as Node)) {
                setShowProfileMenu(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, []);

    if (isLoading) {
        return (
            <header className={styles.header}>
                <Link href="/" className={styles.logoSection}>
                    <Image src="/logo.png" alt="Логотип" width={64} height={64} className={styles.logoImage}/>
                    <span className={styles.logoText}>FriendShip</span>
                </Link>
                {!isMobile && (
                    <>
                        <nav className={styles.navSection}>
                            <Link href="/news" className={styles.navLink}>Новости</Link>
                            <Link href="/groups" className={styles.navLink}>Группы</Link>
                        </nav>
                        <div className={styles.authSection}>
                            <div className={styles.loadingSkeleton}>Загрузка...</div>
                        </div>
                    </>
                )}
                {isMobile && (
                    <div className={styles.authSection}>
                        <a 
                            href={MOBILE_APP_URL}
                            className={styles.mobileAppLink}
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            Приложение
                        </a>
                    </div>
                )}
            </header>
        );
    }

    const handleAvatarClick = () => {
        setShowProfileMenu(!showProfileMenu);
    };

    const handleProfileClick = () => {
        setShowProfileMenu(false);
        router.push('/profile/' + userData?.us);
    };

    const handleLogoutClick = () => {
        setShowProfileMenu(false);
        setShowLogoutModal(true);
    };

    const handleConfirmLogout = () => {
        logout();
        setShowLogoutModal(false);
        router.push('/login');
    };

    const handleCancelLogout = () => {
        setShowLogoutModal(false);
    };

    const handleCreateClick = async () => {
        setIsCheckingGroups(true);
        
        try {
            const accessToken = await getAccesToken(router) || '';
            const groupsData = await getOwnGroups(accessToken);
            
            if (!groupsData || groupsData.length === 0) {
                showNotification(400, 'Для создания события необходимо иметь собственную группу');
                setIsCheckingGroups(false);
                return;
            }
            
            setIsCheckingGroups(false);
            setIsCreateModalOpen(true);
        } catch (error) {
            console.error('Ошибка загрузки групп:', error);
            showNotification(500, 'Не удалось загрузить список групп');
            setIsCheckingGroups(false);
        }
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

    if (isMobile) {
        return (
            <header className={styles.header}>
                <Link href="/" className={styles.logoSection}>
                    <Image src="/logo.png" alt="Логотип" width={64} height={64} className={styles.logoImage}/>
                    <span className={styles.logoText}>FriendShip</span>
                </Link>
                <div className={styles.authSection}>
                    <a 
                        href={MOBILE_APP_URL}
                        className={styles.mobileAppLink}
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        Приложение
                    </a>
                </div>
            </header>
        );
    }

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
                            <button 
                                className={styles.createButton} 
                                onClick={handleCreateClick}
                                disabled={isCheckingGroups}
                            >
                                {isCheckingGroups ? 'Проверка...' : 'Создать'}
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
                                className={styles.profileMenuContainer}
                                ref={profileMenuRef}
                            >
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

                                {showProfileMenu && (
                                    <div className={styles.profileDropdown}>
                                        <button 
                                            className={styles.profileDropdownItem}
                                            onClick={handleProfileClick}
                                        >
                                            Профиль
                                        </button>
                                        <button 
                                            className={styles.profileDropdownItem}
                                            onClick={handleLogoutClick}
                                        >
                                            Выйти
                                        </button>
                                    </div>
                                )}
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

            {showLogoutModal && (
                <div className={styles.modalOverlay} onClick={handleCancelLogout}>
                    <div className={styles.logoutModal} onClick={(e) => e.stopPropagation()}>
                        <h3 className={styles.logoutModalTitle}>Выход из аккаунта</h3>
                        <p className={styles.logoutModalText}>Вы уверены, что хотите выйти?</p>
                        <div className={styles.logoutModalButtons}>
                            <button 
                                className={styles.logoutModalCancel}
                                onClick={handleCancelLogout}
                            >
                                Отмена
                            </button>
                            <button 
                                className={styles.logoutModalConfirm}
                                onClick={handleConfirmLogout}
                            >
                                Выйти
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isCheckingGroups && (
                <div className={styles.modalOverlay}>
                    <div style={{
                        background: 'var(--color-bg-primary)',
                        borderRadius: '12px',
                        padding: '40px',
                        minWidth: '300px'
                    }}>
                        <LoadingIndicator text="Проверка групп..." />
                    </div>
                </div>
            )}

            <EventModal
                isOpen={isCreateModalOpen}
                onClose={handleModalClose}
                onSave={handleEventSave}
                mode="create"
            />
        </>
    );
}