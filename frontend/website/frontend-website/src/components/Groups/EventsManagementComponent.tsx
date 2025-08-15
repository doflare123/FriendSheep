'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import Image from 'next/image';
import EventCard from '../Events/EventCard';
import styles from '../../styles/Groups/admin/EventsManagement.module.css';
import { EventCardProps } from '../../types/Events';

interface EventsManagementComponentProps {
  groupId?: string;
}

// Тестовые данные событий
const mockEvents: EventCardProps[] = [
  {
    id: 1,
    type: 'movies',
    image: '/events/minecraft_movie.jpg',
    date: '03.12.2024',
    title: 'Minecraft в кино',
    genres: ['Комедия', 'Фентези', 'Приключения'],
    participants: 666,
    maxParticipants: 666,
    duration: '101 минуты',
    location: 'offline'
  },
  {
    id: 2,
    type: 'games',
    image: '/events/dota2_tournament.jpg',
    date: '05.12.2024',
    title: 'Dota 2 Tournament',
    genres: ['MOBA', 'Командная'],
    participants: 32,
    maxParticipants: 64,
    duration: '4 часа',
    location: 'online'
  },
  {
    id: 3,
    type: 'board',
    image: '/events/board_games.jpg',
    date: '07.12.2024',
    title: 'Настольные игры',
    genres: ['Стратегия', 'Карточные'],
    participants: 8,
    maxParticipants: 12,
    duration: '3 часа',
    location: 'offline'
  },
  {
    id: 4,
    type: 'other',
    image: '/events/meetup.jpg',
    date: '10.12.2024',
    title: 'IT Meetup',
    genres: ['Программирование', 'Сети'],
    participants: 25,
    maxParticipants: 50,
    duration: '2 часа',
    location: 'online'
  }
];

interface SortOptions {
  category: 'all' | 'games' | 'movies' | 'other';
  date: 'age_asc' | 'age_desc';
  participants: 'participants_asc' | 'participants_desc';
}

// Функция для преобразования даты из формата DD.MM.YYYY в Date
const parseDate = (dateString: string): Date => {
  const [day, month, year] = dateString.split('.');
  return new Date(parseInt(year), parseInt(month) - 1, parseInt(day));
};

const EventsManagementComponent: React.FC<EventsManagementComponentProps> = ({ groupId }) => {
  const [events, setEvents] = useState<EventCardProps[]>(mockEvents);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortOptions, setSortOptions] = useState<SortOptions>({
    category: 'all',
    date: 'age_asc',
    participants: 'participants_asc'
  });
  const [showSortMenu, setShowSortMenu] = useState(false);
  const [canScrollUp, setCanScrollUp] = useState(false);
  const [canScrollDown, setCanScrollDown] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const sortMenuRef = useRef<HTMLDivElement>(null);

  // Закрытие меню сортировки при клике вне его
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (sortMenuRef.current && !sortMenuRef.current.contains(event.target as Node)) {
        setShowSortMenu(false);
      }
    };

    if (showSortMenu) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showSortMenu]);

  // Фильтрация и сортировка событий
  const filteredAndSortedEvents = React.useMemo(() => {
    let filtered = events.filter(event =>
      event.title.toLowerCase().includes(searchTerm.toLowerCase())
    );

    // Применяем фильтрацию по категориям
    if (sortOptions.category !== 'all') {
      if (sortOptions.category === 'other') {
        filtered = filtered.filter(event => event.type === 'other' || event.type === 'board');
      } else {
        filtered = filtered.filter(event => event.type === sortOptions.category);
      }
    }

    filtered = filtered.sort((a, b) => {
      // Сначала сортируем по дате
      const dateA = parseDate(a.date).getTime();
      const dateB = parseDate(b.date).getTime();
      
      let dateComparison;
      if (sortOptions.date === 'age_asc') {
        dateComparison = dateA - dateB;
      } else {
        dateComparison = dateB - dateA;
      }
      
      // Если даты одинаковые, сортируем по участникам
      if (dateComparison === 0) {
        if (sortOptions.participants === 'participants_asc') {
          return a.participants - b.participants;
        } else {
          return b.participants - a.participants;
        }
      }
      
      return dateComparison;
    });

    return filtered;
  }, [events, searchTerm, sortOptions]);

  // Проверка возможности скролла
  const checkScrollability = useCallback(() => {
    if (scrollContainerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = scrollContainerRef.current;
      const hasVerticalScroll = scrollHeight > clientHeight;
      
      setCanScrollUp(hasVerticalScroll && scrollTop > 5);
      setCanScrollDown(hasVerticalScroll && scrollTop < scrollHeight - clientHeight - 5);
    }
  }, []);

  // Проверяем скролл при изменении данных
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      checkScrollability();
    }, 100);

    return () => clearTimeout(timeoutId);
  }, [filteredAndSortedEvents, checkScrollability]);

  // Проверяем скролл при монтировании компонента
  useEffect(() => {
    checkScrollability();
  }, [checkScrollability]);

  // Обработчик изменения размера окна
  useEffect(() => {
    const handleResize = () => {
      checkScrollability();
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [checkScrollability]);

  // Скролл вверх
  const scrollUp = () => {
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollBy({
        top: -300,
        behavior: 'smooth'
      });
    }
  };

  // Скролл вниз
  const scrollDown = () => {
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollBy({
        top: 300,
        behavior: 'smooth'
      });
    }
  };

  // Обработчик редактирования события
  const handleEditEvent = (eventId: number) => {
    console.log('Редактирование события:', eventId);
    // Здесь будет логика перехода на страницу редактирования
  };

  // Обработчик создания нового события
  const handleCreateEvent = () => {
    console.log('Создание нового события');
    // Здесь будет логика перехода на страницу создания события
  };

  // Обработчики для сортировки
  const handleCategoryChange = (category: 'all' | 'games' | 'movies' | 'other') => {
    setSortOptions(prev => ({ ...prev, category }));
    setShowSortMenu(false); // Закрываем меню после выбора
  };

  const handleDateChange = (date: 'age_asc' | 'age_desc') => {
    setSortOptions(prev => ({ ...prev, date }));
    setShowSortMenu(false); // Закрываем меню после выбора
  };

  const handleParticipantsChange = (participants: 'participants_asc' | 'participants_desc') => {
    setSortOptions(prev => ({ ...prev, participants }));
    setShowSortMenu(false); // Закрываем меню после выбора
  };

  return (
    <div className={styles.eventsContainer}>
      {/* Поиск с кнопкой сортировки и кнопкой создания */}
      <div className={styles.controlsSection}>
        <div className={styles.searchWrapper}>
          <input
            type="text"
            placeholder="Найти событие..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className={styles.searchInput}
          />
          <button 
            className={styles.sortButton}
            onClick={() => setShowSortMenu(!showSortMenu)}
          >
            <Image src="/sorting.png" alt="sorting" width={20} height={20} />
          </button>
          
          {/* Выпадающее меню сортировки */}
          {showSortMenu && (
            <div ref={sortMenuRef} className={styles.sortMenu}>
              <div className={styles.sortGroup}>
                <h4 className={styles.sortGroupTitle}>Сортировка по категориям</h4>
                <label className={styles.sortOption}>
                  <input
                    type="radio"
                    name="category"
                    checked={sortOptions.category === 'all'}
                    onChange={() => handleCategoryChange('all')}
                  />
                  <span>Все</span>
                </label>
                <label className={styles.sortOption}>
                  <input
                    type="radio"
                    name="category"
                    checked={sortOptions.category === 'games'}
                    onChange={() => handleCategoryChange('games')}
                  />
                  <span>Игры</span>
                </label>
                <label className={styles.sortOption}>
                  <input
                    type="radio"
                    name="category"
                    checked={sortOptions.category === 'movies'}
                    onChange={() => handleCategoryChange('movies')}
                  />
                  <span>Фильмы</span>
                </label>
                <label className={styles.sortOption}>
                  <input
                    type="radio"
                    name="category"
                    checked={sortOptions.category === 'other'}
                    onChange={() => handleCategoryChange('other')}
                  />
                  <span>Другое</span>
                </label>
              </div>

              <div className={styles.sortGroup}>
                <h4 className={styles.sortGroupTitle}>Сортировка по дате</h4>
                <label className={styles.sortOption}>
                  <input
                    type="radio"
                    name="date"
                    checked={sortOptions.date === 'age_asc'}
                    onChange={() => handleDateChange('age_asc')}
                  />
                  <span>По возрастанию</span>
                </label>
                <label className={styles.sortOption}>
                  <input
                    type="radio"
                    name="date"
                    checked={sortOptions.date === 'age_desc'}
                    onChange={() => handleDateChange('age_desc')}
                  />
                  <span>По убыванию</span>
                </label>
              </div>

              <div className={styles.sortGroup}>
                <h4 className={styles.sortGroupTitle}>Сортировка по участникам</h4>
                <label className={styles.sortOption}>
                  <input
                    type="radio"
                    name="participants"
                    checked={sortOptions.participants === 'participants_asc'}
                    onChange={() => handleParticipantsChange('participants_asc')}
                  />
                  <span>По возрастанию</span>
                </label>
                <label className={styles.sortOption}>
                  <input
                    type="radio"
                    name="participants"
                    checked={sortOptions.participants === 'participants_desc'}
                    onChange={() => handleParticipantsChange('participants_desc')}
                  />
                  <span>По убыванию</span>
                </label>
              </div>
            </div>
          )}
        </div>

        {/* Кнопка создания события */}
        <button 
          className={styles.createButton}
          onClick={handleCreateEvent}
        >
          Создать
        </button>
      </div>

      {/* Список событий с вертикальным скроллом */}
      <div className={styles.eventsList}>
        {/* Стрелка вверх */}
        {canScrollUp && (
          <button className={`${styles.scrollButton} ${styles.scrollUp}`} onClick={scrollUp}>
            ↑
          </button>
        )}

        {/* Контейнер со скроллом */}
        <div 
          ref={scrollContainerRef}
          className={styles.scrollContainer}
          onScroll={checkScrollability}
        >
          {filteredAndSortedEvents.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>📅</div>
              <div className={styles.emptyText}>
                {searchTerm ? 'События не найдены' : 'Нет событий'}
              </div>
            </div>
          ) : (
            <div className={styles.eventsGrid}>
              {filteredAndSortedEvents.map((event) => (
                <div key={event.id} className={styles.eventCardWrapper}>
                  <div className={styles.eventCardScaler}>
                    <EventCard
                      {...event}
                      isEditMode={true}
                      onEdit={() => handleEditEvent(event.id)}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Стрелка вниз */}
        {canScrollDown && (
          <button className={`${styles.scrollButton} ${styles.scrollDown}`} onClick={scrollDown}>
            ↓
          </button>
        )}
      </div>
    </div>
  );
};

export default EventsManagementComponent;