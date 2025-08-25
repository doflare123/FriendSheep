import React from 'react';
import { createPortal } from 'react-dom';
import Image from 'next/image';
import styles from '@/styles/Events/EventDetailModal.module.css';

interface EventDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  onJoin?: () => void;
  eventId?: number;
}

// Тестовые данные на основе структуры с сервера
const mockEventData = {
  metadata: {
    ageLimit: "18+",
    country: "Russia",
    genres: [
      "Выживание",
      "Соревновательная", 
      "Батл роял",
      "Приключения",
      "Приключения",
      "Соревновательная",
      "Батл роял",
      "Приключения",
      "Приключения"
    ],
    location: "https://discord.gg/5fbH3cJr",
    notes: "Банан - это съедобный плод, относящийся к ягодам, который вырастает на многолетнем травянистом растении, также называемом банан. Из рода Musa. Растения рода банан, обычно, имеют высокий стебель и крупные листья, и широко распространены в тропических и субтропических регионах",
    year: 2026
  },
  session: {
    count_users_max: 0,
    current_users: 3,
    duration: 120,
    end_time: "2025-02-01T20:55:00",
    group_id: 1,
    id: 1,
    image_url: "/events/banana.jpg", // путь к изображению банана
    session_place: "https://discord.gg/5fbH3cJr",
    session_type: "offline",
    start_time: "2025-02-01T18:55:00",
    title: "БАНАНЗА ФОРЕВА"
  }
};

const EventDetailModal: React.FC<EventDetailModalProps> = ({ 
  isOpen, 
  onClose, 
  onJoin,
  eventId 
}) => {
  if (!isOpen) return null;

  // TODO: Здесь будет запрос к серверу по eventId
  // const eventData = await fetchEventById(eventId);
  console.log('Загружаем данные для события с ID:', eventId);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const day = date.getDate().toString().padStart(2, '0');
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const year = date.getFullYear();
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    
    return `${day}.${month}.${year} ${hours}:${minutes}`;
  };

  const formatDuration = (minutes: number) => {
    return `${minutes} минут`;
  };

  // Функция для проверки, является ли строка ссылкой
  const isUrl = (string: string) => {
    try {
      new URL(string);
      return true;
    } catch {
      return false;
    }
  };

  // Функция для получения ссылки на издателя (поиск в Яндексе)
  const getPublisherLink = () => {
    const publisher = "Japan Nintendo Co., Ltd.";
    return `https://yandex.ru/search/?text=${encodeURIComponent(publisher)}`;
  };

  // Функция для получения ссылки на место проведения
  const getLocationLink = () => {
    const location = mockEventData.session.session_place;
    if (isUrl(location)) {
      return location;
    }
    return `https://yandex.ru/search/?text=${encodeURIComponent(location)}`;
  };

  // Показываем только первые 9 жанров
  const displayGenres = mockEventData.metadata.genres.slice(0, 9);

  const modalContent = (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div className={styles.modalContainer} onClick={(e) => e.stopPropagation()}>
        {/* Заголовок и кнопка закрытия */}
        <div className={styles.modalHeader}>
          <h2 className={styles.modalTitle}>{mockEventData.session.title}</h2>
          <button className={styles.closeButton} onClick={onClose}>
            <Image 
              src="/icons/close.png" 
              alt="Закрыть" 
              width={24} 
              height={24}
            />
          </button>
        </div>

        {/* Изображение события */}
        <div className={styles.eventImageContainer}>
          <Image
            src={mockEventData.session.image_url}
            alt={mockEventData.session.title}
            fill
            className={styles.eventImage}
          />
          <div className={styles.participantsBadge}>
            <Image 
              src={`/events/${mockEventData.session.session_type}.png`} 
              alt={mockEventData.session.session_type} 
              width={20} 
              height={20}
            />
          </div>
        </div>

        {/* Информация о дате и времени */}
        <div className={styles.eventInfo}>
          <div className={styles.dateInfo}>
            <span className={styles.dateValue}>Дата проведения: {formatDate(mockEventData.session.start_time)}</span>
          </div>
          <div className={styles.durationInfo}>
            <span>{formatDuration(mockEventData.session.duration)}</span>
            <Image 
              src="/events/clock.png" 
              alt="Длительность" 
              width={16} 
              height={16}
            />
          </div>
        </div>

        {/* Описание события */}
        <div className={styles.eventDescription}>
          <p>{mockEventData.metadata.notes}</p>
        </div>

        {/* Жанры */}
        <div className={styles.genresSection}>
          <span className={styles.genresLabel}>Жанры:</span>
          <div className={styles.genresContainer}>
            {displayGenres.map((genre, index) => (
              <span key={index} className={styles.genreTag}>
                {genre}
              </span>
            ))}
          </div>
        </div>

        {/* Информация об издателе */}
        <div className={styles.publisherInfo}>
          <div className={styles.publisherRow}>
            <span className={styles.publisherLabel}>Издатель:</span>
            <a 
              href={getPublisherLink()}
              target="_blank" 
              rel="noopener noreferrer"
              className={`${styles.publisherValue} ${styles.underlined}`}
            >
              Japan Nintendo Co., Ltd.
            </a>
          </div>
          <div className={styles.publisherRow}>
            <span className={styles.publisherLabel}>Год издания:</span>
            <span className={styles.publisherValue}>{mockEventData.metadata.year}</span>
          </div>
          <div className={styles.publisherRow}>
            <span className={styles.publisherLabel}>Место проведения:</span>
            <a 
                href={getLocationLink()}
                target="_blank" 
                rel="noopener noreferrer"
                className={`${styles.publisherValue} ${styles.underlined}`}
            >
              {mockEventData.session.session_place}
            </a>
          </div>
          <div className={styles.publisherRow}>
            <span className={styles.publisherLabel}>Возрастное ограничение:</span>
            <span className={styles.publisherValue}>{mockEventData.metadata.ageLimit}</span>
          </div>
        </div>

        {/* Кнопка присоединиться */}
        <button className={styles.joinButton} onClick={onJoin}>
          Присоединиться
        </button>
      </div>
    </div>
  );

  // Используем портал для рендера модального окна в body
  return createPortal(modalContent, document.body);
};

export default EventDetailModal;