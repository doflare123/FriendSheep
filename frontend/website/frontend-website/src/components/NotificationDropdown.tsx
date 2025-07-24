'use client';

import Image from 'next/image';
import { useState, useEffect } from 'react';
import '../styles/NotificationDropdown.css';

interface NewsNotification {
    id: string;
    type: 'news';
    text: string;
    image: string;
    isRead: boolean;
    timestamp: Date;
}

interface InviteNotification {
    id: string;
    type: 'invite';
    groupName: string;
    image: string;
    timestamp: Date;
}

type Notification = NewsNotification | InviteNotification;

interface NotificationDropdownProps {
    onClose: () => void;
}

export default function NotificationDropdown({ onClose }: NotificationDropdownProps) {
  const [notifications, setNotifications] = useState<Notification[]>([
    {
      id: '1',
      type: 'news',
      text: 'Событие "Какое-то событие" начнётся через 30 мин',
      image: '/notification-news.png',
      isRead: false,
      timestamp: new Date(Date.now() - 3600000) // 1 час назад
    },
    {
      id: '2',
      type: 'news',
      text: 'Событие "Какое-то событие" начнётся через 30 мин',
      image: '/notification-news.png',
      isRead: false,
      timestamp: new Date(Date.now() - 1800000) // 30 минут назад
    },
    {
      id: '3',
      type: 'invite',
      groupName: 'МЕГАКРУТЫЕ ПАЦАНЯТА',
      image: '/group-invite.png',
      timestamp: new Date(Date.now() - 300000) // 5 минут назад
    },
    {
      id: '4',
      type: 'invite',
      groupName: 'МЕГАКРУТЫЕ ПАЦАНЯТА',
      image: '/group-invite.png',
      timestamp: new Date(Date.now() - 30000) // 30 секунд назад
    }
  ]);

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

  const handleNewsHover = (id: string) => {
    setNotifications(prev => prev.map(notif => 
      notif.id === id && notif.type === 'news' 
        ? { ...notif, isRead: true }
        : notif
    ));
  };

  const handleInviteAction = (id: string, action: 'accept' | 'decline') => {
    setNotifications(prev => prev.filter(notif => notif.id !== id));
  };

  // Обновление времени каждую секунду
  useEffect(() => {
    const interval = setInterval(() => {
      // Принудительно обновляем компонент для пересчета времени
      setNotifications(prev => [...prev]);
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  // Закрытие при клике вне компонента
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest('.notification-dropdown') && !target.closest('.notification-button')) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [onClose]);

  return (
    <div className="notification-dropdown">
      <div className="notification-header">
        У вас {unreadCount} непрочитанных уведомлений:
      </div>
      
      <div className="notification-list">
        {notifications.map((notification) => (
          <div key={notification.id} className="notification-item">
            <div className="notification-image">
              <Image 
                src={notification.image} 
                alt="Уведомление" 
                width={40} 
                height={40}
              />
            </div>
            
            <div className="notification-content">
              {notification.type === 'news' ? (
                <div className="news-wrapper">
                  <div 
                    className={`news-text ${notification.isRead ? 'read' : ''}`}
                    onMouseEnter={() => handleNewsHover(notification.id)}
                  >
                    {notification.text}
                  </div>
                  <div className="notification-time">
                    {getTimeAgo(notification.timestamp)}
                  </div>
                </div>
              ) : (
                <div className="invite-content">
                  <div className="invite-text">
                    Вас приглашают в группу "{notification.groupName}"
                  </div>
                  <div className="invite-buttons-wrapper">
                    <div className="invite-buttons">
                      <button 
                        className="invite-accept"
                        onClick={() => handleInviteAction(notification.id, 'accept')}
                      >
                        Принять
                      </button>
                      <button 
                        className="invite-decline"
                        onClick={() => handleInviteAction(notification.id, 'decline')}
                      >
                        Отклонить
                      </button>
                    </div>
                    <div className="notification-time">
                      {getTimeAgo(notification.timestamp)}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}