'use client';

import Image from 'next/image';
import { useState, useEffect } from 'react';
import styles from '@/styles/NotificationDropdown.module.css';
import { getNotif } from '@/api/notification/getNotif';
import { viewNotif } from '@/api/notification/viewNotif';
import { approveInvite } from '@/api/groups/approveInvite';
import { rejectInvite } from '@/api/groups/rejectInvite';
import { getAccesToken } from '@/Constants';
import { showNotification } from '@/utils';
import LoadingIndicator from '@/components/LoadingIndicator';

interface NewsNotification {
  id: number;
  type: 'news';
  notifType: '24_hours' | '6_hours' | '1_hours';
  text: string;
  image: string;
  isRead: boolean;
  timestamp: Date;
}

interface InviteNotification {
  id: number;
  type: 'invite';
  groupName: string;
  groupId: number;
  image: string;
  timestamp: Date;
}

type Notification = NewsNotification | InviteNotification;

interface NotificationDropdownProps {
  onClose: () => void;
}

export default function NotificationDropdown({ onClose }: NotificationDropdownProps) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [processingIds, setProcessingIds] = useState<Set<number>>(new Set());

  // Получение изображения в зависимости от типа уведомления
  const getNotificationImage = (notifType: string): string => {
    switch (notifType) {
      case '24_hours':
        return '/notification/24_hours.png';
      case '6_hours':
        return '/notification/6_hours.png';
      case '1_hours':
        return '/notification/1_hours.png';
      default:
        return '/notification/24_hours.png';
    }
  };

  // Загрузка уведомлений
  useEffect(() => {
    const loadNotifications = async () => {
      setIsLoading(true);
      try {
        const accessToken = getAccesToken();
        const data = await getNotif(accessToken);

        const allNotifications: Notification[] = [];

        // Обработка notifications
        if (data.notifications && Array.isArray(data.notifications)) {
          data.notifications.forEach((notif: any) => {
            allNotifications.push({
              id: notif.id,
              type: 'news',
              notifType: notif.type,
              text: notif.text || `Событие начнётся через ${notif.type === '24_hours' ? '24 часа' : notif.type === '6_hours' ? '6 часов' : '1 час'}`,
              image: getNotificationImage(notif.type),
              isRead: notif.viewed,
              timestamp: new Date(notif.sendAt),
            });
          });
        }

        // Обработка invites
        if (data.invites && Array.isArray(data.invites)) {
          data.invites.forEach((invite: any) => {
            allNotifications.push({
              id: invite.id,
              type: 'invite',
              groupName: invite.groupName,
              groupId: invite.groupId,
              image: '/notification/group_invite.png',
              timestamp: new Date(invite.createdAt),
            });
          });
        }

        // Сортируем по времени (новые сверху)
        allNotifications.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());

        setNotifications(allNotifications);
      } catch (error) {
        console.error('Ошибка загрузки уведомлений:', error);
        showNotification(500, 'Не удалось загрузить уведомления');
      } finally {
        setIsLoading(false);
      }
    };

    loadNotifications();
  }, []);

  // Функция для расчета времени, прошедшего с момента уведомления
  const getTimeAgo = (timestamp: Date): string => {
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - timestamp.getTime()) / 1000);

    if (diffInSeconds < 60) {
      return `${diffInSeconds} сек назад`;
    } else if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60);
      return `${minutes} мин назад`;
    } else if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600);
      return `${hours} час${hours === 1 ? '' : hours < 5 ? 'а' : 'ов'} назад`;
    } else if (diffInSeconds < 604800) {
      const days = Math.floor(diffInSeconds / 86400);
      return `${days} д${days === 1 ? 'ень' : days < 5 ? 'ня' : 'ней'} назад`;
    } else if (diffInSeconds < 2629800) {
      const weeks = Math.floor(diffInSeconds / 604800);
      return `${weeks} недел${weeks === 1 ? 'ю' : weeks < 5 ? 'и' : 'ь'} назад`;
    } else if (diffInSeconds < 31557600) {
      const months = Math.floor(diffInSeconds / 2629800);
      return `${months} месяц${months === 1 ? '' : months < 5 ? 'а' : 'ев'} назад`;
    } else {
      const years = Math.floor(diffInSeconds / 31557600);
      return `${years} ${years === 1 ? 'год' : years < 5 ? 'года' : 'лет'} назад`;
    }
  };

  const unreadCount = notifications.filter(n =>
    n.type === 'news' ? !n.isRead : true
  ).length;

  const handleNewsHover = async (id: number) => {
    const notification = notifications.find(n => n.id === id && n.type === 'news');
    if (!notification || (notification.type === 'news' && notification.isRead)) {
      return;
    }

    // Оптимистично обновляем UI
    setNotifications(prev =>
      prev.map(notif =>
        notif.id === id && notif.type === 'news'
          ? { ...notif, isRead: true }
          : notif
      )
    );

    try {
      const accessToken = getAccesToken();
      await viewNotif(accessToken, id);
    } catch (error) {
      console.error('Ошибка при отметке уведомления как прочитанного:', error);
      // Откатываем изменения при ошибке
      setNotifications(prev =>
        prev.map(notif =>
          notif.id === id && notif.type === 'news'
            ? { ...notif, isRead: false }
            : notif
        )
      );
    }
  };

  const handleInviteAction = async (id: number, action: 'accept' | 'decline') => {
    const invite = notifications.find(n => n.id === id && n.type === 'invite');
    if (!invite || invite.type !== 'invite') return;

    setProcessingIds(prev => new Set(prev).add(id));

    try {
      const accessToken = getAccesToken();

      if (action === 'accept') {
        await approveInvite(accessToken, id);
        showNotification(200, `Вы успешно присоединились к группе "${invite.groupName}"`);
      } else {
        await rejectInvite(accessToken, id);
        showNotification(200, `Приглашение в группу "${invite.groupName}" отклонено`);
      }

      // Удаляем уведомление из списка
      setNotifications(prev => prev.filter(notif => notif.id !== id));
    } catch (error: any) {
      console.error('Ошибка при обработке приглашения:', error);
      const message = action === 'accept'
        ? 'Не удалось принять приглашение'
        : 'Не удалось отклонить приглашение';
      showNotification(error.response?.status || 500, message);
    } finally {
      setProcessingIds(prev => {
        const newSet = new Set(prev);
        newSet.delete(id);
        return newSet;
      });
    }
  };

  // Обновление времени каждую секунду
  useEffect(() => {
    const interval = setInterval(() => {
      setNotifications(prev => [...prev]);
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  // Закрытие при клике вне компонента
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest(`.${styles.notificationDropdown}`) && !target.closest('.notification-button')) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [onClose]);

  return (
    <div className={styles.notificationDropdown}>
      <div className={styles.notificationHeader}>
        У вас {unreadCount} непрочитанных уведомлений:
      </div>

      <div className={styles.notificationList}>
        {isLoading ? (
          <div style={{ padding: '20px', textAlign: 'center' }}>
            <LoadingIndicator text="Загрузка уведомлений..." />
          </div>
        ) : notifications.length === 0 ? (
          <div style={{ padding: '20px', textAlign: 'center', color: '#888' }}>
            Нет уведомлений
          </div>
        ) : (
          notifications.map((notification) => {
            const isProcessing = processingIds.has(notification.id);

            return (
              <div key={notification.id} className={styles.notificationItem}>
                <div className={styles.notificationImage}>
                  <Image
                    src={notification.image}
                    alt="Уведомление"
                    width={40}
                    height={40}
                  />
                </div>

                <div className={styles.notificationContent}>
                  {notification.type === 'news' ? (
                    <div className={styles.newsWrapper}>
                      <div
                        className={`${styles.newsText} ${notification.isRead ? styles.read : ''}`}
                        onMouseEnter={() => handleNewsHover(notification.id)}
                      >
                        {notification.text}
                      </div>
                      <div className={styles.notificationTime}>
                        {getTimeAgo(notification.timestamp)}
                      </div>
                    </div>
                  ) : (
                    <div className={styles.inviteContent}>
                      <div className={styles.inviteText}>
                        Вас приглашают в группу "{notification.groupName}"
                      </div>
                      {isProcessing ? (
                        <div style={{ padding: '8px 0' }}>
                          <LoadingIndicator text="Обработка..." />
                        </div>
                      ) : (
                        <div className={styles.inviteButtonsWrapper}>
                          <div className={styles.inviteButtons}>
                            <button
                              className={styles.inviteAccept}
                              onClick={() => handleInviteAction(notification.id, 'accept')}
                            >
                              Принять
                            </button>
                            <button
                              className={styles.inviteDecline}
                              onClick={() => handleInviteAction(notification.id, 'decline')}
                            >
                              Отклонить
                            </button>
                          </div>
                          <div className={styles.notificationTime}>
                            {getTimeAgo(notification.timestamp)}
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}