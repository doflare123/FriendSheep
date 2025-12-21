'use client';

import { useState, useRef, useEffect } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import styles from '../styles/GuestHomeStyles.module.css';
import EventCard from './Events/EventCard';
import QRCode from './QRCode';
import { EventCardProps } from '../types/Events';
import { MOBILE_APP_URL } from '@/Constants';

const getFutureDate = (daysFromNow: number): string => {
  const date = new Date();
  date.setDate(date.getDate() + daysFromNow);
  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const year = date.getFullYear();
  return `${day}.${month}.${year}`;
};

const demoEvents: EventCardProps[] = [
  {
    id: 1,
    type: 'games',
    image: '/preview/dota2.png',
    date: getFutureDate(3),
    title: 'Dota 2',
    genres: ['Экшены', 'Ммо'],
    participants: 4,
    maxParticipants: 5,
    duration: "300",
    location: 'online',
    start_time: '20:00',
    adress: 'Discord'
  },
  {
    id: 2,
    type: 'movies',
    image: '/preview/k_otec.png',
    date: getFutureDate(5),
    title: 'Крестный отец',
    genres: ['Драма', 'Криминал'],
    participants: 10,
    maxParticipants: 13,
    duration: "175",
    location: 'online',
    start_time: '19:00',
    adress: 'Zoom'
  },
  {
    id: 3,
    type: 'board',
    image: '/preview/monopoly.png',
    date: getFutureDate(7),
    title: 'Monopoly',
    genres: ['Экономическая стратегия', 'Семейная'],
    participants: 6,
    maxParticipants: 6,
    duration: "120",
    location: 'offline',
    city: 'Москва',
    start_time: '18:00',
    adress: 'ул. Арбат, 10'
  }
];

export default function GuestHome() {
  const [currentCard, setCurrentCard] = useState(1);
  const [isMobile, setIsMobile] = useState(false);
  const touchStartX = useRef(0);
  const touchEndX = useRef(0);
  const mouseStartX = useRef(0);
  const isDragging = useRef(false);

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(/Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent));
    };
    
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  const handlePrevCard = () => {
    setCurrentCard((prev) => (prev === 0 ? demoEvents.length - 1 : prev - 1));
  };

  const handleNextCard = () => {
    setCurrentCard((prev) => (prev === demoEvents.length - 1 ? 0 : prev + 1));
  };

  const handleTouchStart = (e: React.TouchEvent) => {
    touchStartX.current = e.touches[0].clientX;
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    touchEndX.current = e.touches[0].clientX;
  };

  const handleTouchEnd = () => {
    const swipeDistance = touchStartX.current - touchEndX.current;
    if (Math.abs(swipeDistance) > 50) {
      if (swipeDistance > 0) {
        handleNextCard();
      } else {
        handlePrevCard();
      }
    }
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    isDragging.current = true;
    mouseStartX.current = e.clientX;
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDragging.current) return;
    e.preventDefault();
  };

  const handleMouseUp = (e: React.MouseEvent) => {
    if (!isDragging.current) return;
    isDragging.current = false;
    
    const swipeDistance = mouseStartX.current - e.clientX;
    if (Math.abs(swipeDistance) > 50) {
      if (swipeDistance > 0) {
        handleNextCard();
      } else {
        handlePrevCard();
      }
    }
  };

  const handleMouseLeave = () => {
    isDragging.current = false;
  };

  const handleWheel = (e: React.WheelEvent) => {
    if (Math.abs(e.deltaX) > Math.abs(e.deltaY)) {
      e.preventDefault();
      if (e.deltaX > 30) {
        handleNextCard();
      } else if (e.deltaX < -30) {
        handlePrevCard();
      }
    } else if (Math.abs(e.deltaY) > 30) {
      e.preventDefault();
      if (e.deltaY > 0) {
        handleNextCard();
      } else {
        handlePrevCard();
      }
    }
  };

  return (
    <div className={styles.container}>
        <div className={styles.hero}>
            <div className={styles.heroContent}>
            <h1 className={styles.heroTitle}>
                Ваш досуг с друзьями - проще,<br />чем когда-либо
            </h1>
            <p className={styles.heroSubtitle}>
                FriendShip помогает организовать встречи, планировать мероприятия и никогда<br />
                не забывать о важных событиях с друзьями
            </p>
            <div className={styles.heroButtons}>
                {isMobile ? (
                  <a 
                    href={MOBILE_APP_URL}
                    className={styles.createAccountButton}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Скачать мобильное приложение
                  </a>
                ) : (
                  <Link href="/register" className={styles.createAccountButton}>
                    Создать аккаунт
                  </Link>
                )}
                <Link href="/info/about" className={styles.learnMoreButton}>
                Узнать больше
                </Link>
            </div>
            </div>

            <div 
              className={styles.cardsPreview}
              onTouchStart={handleTouchStart}
              onTouchMove={handleTouchMove}
              onTouchEnd={handleTouchEnd}
              onMouseDown={handleMouseDown}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onMouseLeave={handleMouseLeave}
              onWheel={handleWheel}
            >
              {demoEvents.map((event, index) => {
                const position = (index - currentCard + demoEvents.length) % demoEvents.length;
                let transform = '';
                let zIndex = 1;
                let opacity = 0;

                if (position === 0) {
                  transform = 'translateX(0) scale(1) rotate(0deg)';
                  zIndex = 10;
                  opacity = 1;
                } else if (position === 1) {
                  transform = 'translateX(100px) scale(0.85) rotate(8deg)';
                  zIndex = 5;
                  opacity = 0.7;
                } else if (position === demoEvents.length - 1) {
                  transform = 'translateX(-100px) scale(0.85) rotate(-8deg)';
                  zIndex = 5;
                  opacity = 0.7;
                } else {
                  transform = position < demoEvents.length / 2 
                    ? 'translateX(150px) scale(0.7) rotate(12deg)' 
                    : 'translateX(-150px) scale(0.7) rotate(-12deg)';
                  zIndex = 1;
                  opacity = 0;
                }

                return (
                  <div 
                    key={event.id} 
                    className={styles.cardWrapper}
                    style={{ 
                      transform,
                      zIndex,
                      opacity,
                    }}
                  >
                    <EventCard {...event} isDemo={true} />
                  </div>
                );
              })}
            </div>
        </div>

        <div className={styles.features}>
            <h2 className={styles.featuresTitle}>Что умеет FriendShip?</h2>
            <p className={styles.featuresSubtitle}>
            Всё необходимое для организации досуга в одном месте
            </p>

            <div className={styles.featuresList}>
            <div className={styles.featureCard}>
                <div className={styles.featureIcon}>
                <Image src="/icons/skill1.png" alt="Планирование" width={64} height={64} />
                </div>
                <h3 className={styles.featureTitle}>Планирование мероприятий</h3>
                <p className={styles.featureDescription}>
                Создавайте события с датой, местом и временем. Друзья сразу получат 
                приглашение и смогут откликнуться на него в пару кликов
                </p>
            </div>

            <div className={styles.featureCard}>
                <div className={styles.featureIcon}>
                <Image src="/icons/skill2.png" alt="Группы" width={64} height={64} />
                </div>
                <h3 className={styles.featureTitle}>Группы друзей</h3>
                <p className={styles.featureDescription}>
                Создавайте закрытые и открытые группы для разных 
                компаний друзей: отдельно работа, учёба, семья и любимые друзья
                </p>
            </div>

            <div className={styles.featureCard}>
                <div className={styles.featureIcon}>
                <Image src="/icons/skill3.png" alt="Telegram-бот" width={64} height={64} />
                </div>
                <h3 className={styles.featureTitle}>Telegram-бот</h3>
                <p className={styles.featureDescription}>
                Получайте напоминания и оповещения о новых 
                событиях, сообщениях в группах и приглашениях на мероприятия
                </p>
            </div>
            </div>
        </div>

        <div className={styles.howItWorks}>
            <h2 className={styles.howItWorksTitle}>Как это работает?</h2>
            <p className={styles.howItWorksSubtitle}>
            Всего три простых шага до организации вашего досуга
            </p>

            <div className={styles.stepsList}>
            <div className={styles.stepCard}>
                <div className={styles.stepNumber}>1</div>
                <h3 className={styles.stepTitle}>Регистрация</h3>
                <p className={styles.stepDescription}>
                Создайте аккаунт и добавьте друзей. Это займет меньше минуты
                </p>
            </div>

            <div className={styles.stepCard}>
                <div className={styles.stepNumber}>2</div>
                <h3 className={styles.stepTitle}>Создание события</h3>
                <p className={styles.stepDescription}>
                Опишите мероприятие, укажите время и место. Ваши друзья сразу 
                получат уведомление
                </p>
            </div>

            <div className={styles.stepCard}>
                <div className={styles.stepNumber}>3</div>
                <h3 className={styles.stepTitle}>Наслаждайтесь</h3>
                <p className={styles.stepDescription}>
                Получайте напоминания и проводите время со своей компанией
                </p>
            </div>
            </div>
        </div>

        <div className={styles.cta}>
            <h2 className={styles.ctaTitle}>Готовы начать?</h2>
            <p className={styles.ctaSubtitle}>
            Присоединяйтесь к FriendShip и сделайте досуг с друзьями незабываемым
            </p>
            {isMobile ? (
              <QRCode size={200} showText={true} />
            ) : (
              <div className={styles.ctaButtons}>
                <Link href="/register" className={styles.ctaButton}>
                  Создать аккаунт
                </Link>
                <Link href="/login" className={styles.ctaButtonSecondary}>
                  Войти в аккаунт
                </Link>
              </div>
            )}
            <p className={styles.ctaFooter}>
            Никаких сложностей. Начните использовать прямо сейчас
            </p>
        </div>
    </div>
  );
}