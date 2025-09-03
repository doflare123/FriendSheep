'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { useSearchParams } from "next/navigation"
import Image from 'next/image';
import styles from '../../../styles/search/eventsSearch.module.css';
import EventCard from '../../../components/Events/EventCard';
import { EventCardProps } from '../../../types/Events';
import CategoryLabel from "../../../components/Events/CategoryLabel"

const ITEMS_PER_PAGE = 10;

// Тестовые данные
const mockEvents: EventCardProps[] = Array.from({ length: 10 }, (_, i) => ({
  id: i + 1,
  type: (['games', 'movies', 'board', 'other'] as const)[i % 4],
  image: '/event_card.jpg',
  date: `${15 + (i % 15)} ноября`,
  title: `Крестный отец ${i + 1}`,
  genres: ['драма', 'криминал', 'классика'][Math.floor(i / 3)] ? 
    ['драма', 'криминал', 'классика'].slice(0, Math.floor(i / 3) + 1) : 
    ['драма'],
  participants: 5 + (i % 10),
  maxParticipants: 15 + (i % 5),
  duration: `${120 + (i * 10)} мин`,
  location: i % 2 === 0 ? 'online' : 'offline',
  adress: i % 2 === 0 ? 'Discord' : 'ул. Примерная, д. 123',
  groupId: i + 100
}));

export default function EventsSearchPage() {
  const searchParams = useSearchParams()
  const category = searchParams?.get("category") || 'Популярные';
  const pattern = searchParams?.get("pattern") || '/events/movie_bg.png';
  const [searchQuery, setSearchQuery] = useState('');
  const [events, setEvents] = useState<EventCardProps[]>(mockEvents);
  const [isFilterOpen, setIsFilterOpen] = useState(false);
  const [selectedCategorySort, setSelectedCategorySort] = useState('Все');
  const [selectedDateSort, setSelectedDateSort] = useState('По возрастанию');
  const [selectedParticipantSort, setSelectedParticipantSort] = useState('По возрастанию');
  const [isLoading, setIsLoading] = useState(false);
  const [canScrollUp, setCanScrollUp] = useState(false);
  const [canScrollDown, setCanScrollDown] = useState(true);
  
  const filterRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Проверяем, является ли категория специфическим типом контента
  const isSpecificCategory = ['фильмы', 'игры', 'настольные игры', 'другое'].includes(category.toLowerCase());

  // Проверка возможности прокрутки
  const checkScrollButtons = () => {
    if (containerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
      setCanScrollUp(scrollTop > 0);
      setCanScrollDown(scrollTop < scrollHeight - clientHeight - 1);
    }
  };

  // Фильтрация по поисковому запросу
  const filteredEvents = events.filter(event => 
    event.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Закрытие фильтра при клике вне его
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (filterRef.current && !filterRef.current.contains(event.target as Node)) {
        setIsFilterOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  // Функция для подгрузки новых событий
  const loadMoreEvents = useCallback(() => {
    if (isLoading) return;

    setIsLoading(true);
    
    // Имитируем загрузку
    setTimeout(() => {
      const newEvents = Array.from({ length: 10 }, (_, i) => {
        const baseIndex = events.length + i;
        return {
          id: baseIndex + 1,
          type: (['games', 'movies', 'board', 'other'] as const)[baseIndex % 4],
          image: '/event_card.jpg',
          date: `${15 + (baseIndex % 15)} ноября`,
          title: `Новое событие ${baseIndex + 1}`,
          genres: ['драма', 'криминал', 'классика'][Math.floor(baseIndex / 3)] ? 
            ['драма', 'криминал', 'классика'].slice(0, Math.floor(baseIndex / 3) + 1) : 
            ['драма'],
          participants: 5 + (baseIndex % 10),
          maxParticipants: 15 + (baseIndex % 5),
          duration: `${120 + (baseIndex * 10)} мин`,
          location: baseIndex % 2 === 0 ? 'online' : 'offline',
          adress: baseIndex % 2 === 0 ? 'Discord' : 'ул. Примерная, д. 123',
          groupId: baseIndex + 100
        } as EventCardProps;
      });

      setEvents(prevEvents => [...prevEvents, ...newEvents]);
      setIsLoading(false);
    }, 1000);
  }, [events.length, isLoading]);

  // Бесконечный скролл
  useEffect(() => {
    const handleScroll = () => {
      checkScrollButtons();
      
      if (containerRef.current) {
        const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
        
        // Проверяем, достиг ли пользователь конца контейнера
        if (scrollHeight - scrollTop <= clientHeight + 100 && !isLoading) {
          loadMoreEvents();
        }
      }
    };

    const container = containerRef.current;
    if (container) {
      container.addEventListener('scroll', handleScroll);
      checkScrollButtons(); // Проверяем сразу при монтировании
      
      return () => container.removeEventListener('scroll', handleScroll);
    }
  }, [loadMoreEvents, isLoading]);

  // Функции прокрутки
  const scrollToTop = () => {
    if (containerRef.current) {
      containerRef.current.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const scrollToBottom = () => {
    if (containerRef.current) {
      containerRef.current.scrollTo({ 
        top: containerRef.current.scrollHeight, 
        behavior: 'smooth' 
      });
    }
  };

  return (
    <div className="bgPage">
      <div className={styles.container}>
        <div className={styles.contentWrapper}>
          <div className={styles.categoryHeader}>
            <CategoryLabel 
                title={category} 
                patternUrl={pattern} 
                clickable={false} // делаем некликабельным
            />
          </div>

          <div className={styles.searchHeader}>
            <div className={styles.searchInputWrapper}>
              <input
                type="text"
                placeholder="Найти событие..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className={styles.searchInput}
              />
              <div className={styles.filterWrapper} ref={filterRef}>
                <button
                  className={styles.filterButton}
                  onClick={() => setIsFilterOpen(!isFilterOpen)}
                >
                  <Image 
                    src="/sorting.png" 
                    alt="Filter" 
                    width={20} 
                    height={20} 
                  />
                </button>
                {isFilterOpen && (
                  <div className={styles.filterMenu}>
                    {!isSpecificCategory && (
                      <div className={styles.filterSection}>
                        <h4 className={styles.filterTitle}>Сортировка по категориям</h4>
                        <label className={styles.filterOption}>
                          <input
                            type="radio"
                            name="categories"
                            checked={selectedCategorySort === 'Все'}
                            onChange={() => setSelectedCategorySort('Все')}
                          />
                          <span>Все</span>
                        </label>
                        <label className={styles.filterOption}>
                          <input
                            type="radio"
                            name="categories"
                            checked={selectedCategorySort === 'Кино'}
                            onChange={() => setSelectedCategorySort('Кино')}
                          />
                          <span>Кино</span>
                        </label>
                        <label className={styles.filterOption}>
                          <input
                            type="radio"
                            name="categories"
                            checked={selectedCategorySort === 'Игры'}
                            onChange={() => setSelectedCategorySort('Игры')}
                          />
                          <span>Игры</span>
                        </label>
                        <label className={styles.filterOption}>
                          <input
                            type="radio"
                            name="categories"
                            checked={selectedCategorySort === 'Настольные игры'}
                            onChange={() => setSelectedCategorySort('Настольные игры')}
                          />
                          <span>Настольные игры</span>
                        </label>
                        <label className={styles.filterOption}>
                          <input
                            type="radio"
                            name="categories"
                            checked={selectedCategorySort === 'Другое'}
                            onChange={() => setSelectedCategorySort('Другое')}
                          />
                          <span>Другое</span>
                        </label>
                      </div>
                    )}
                    
                    <div className={styles.filterSection}>
                      <h4 className={styles.filterTitle}>Сортировка по дате</h4>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="date"
                          checked={selectedDateSort === 'По возрастанию'}
                          onChange={() => setSelectedDateSort('По возрастанию')}
                        />
                        <span>По возрастанию</span>
                      </label>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="date"
                          checked={selectedDateSort === 'По убыванию'}
                          onChange={() => setSelectedDateSort('По убыванию')}
                        />
                        <span>По убыванию</span>
                      </label>
                    </div>
                    
                    <div className={styles.filterSection}>
                      <h4 className={styles.filterTitle}>Сортировка по участникам</h4>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="participants"
                          checked={selectedParticipantSort === 'По возрастанию'}
                          onChange={() => setSelectedParticipantSort('По возрастанию')}
                        />
                        <span>По возрастанию</span>
                      </label>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="participants"
                          checked={selectedParticipantSort === 'По убыванию'}
                          onChange={() => setSelectedParticipantSort('По убыванию')}
                        />
                        <span>По убыванию</span>
                      </label>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className={styles.resultsContainer} ref={containerRef}>
            <div className={styles.eventsGrid}>
              {filteredEvents.map((event) => (
                <div key={event.id} className={styles.eventWrapper}>
                  <EventCard {...event} />
                </div>
              ))}
            </div>
            
            {isLoading && (
              <div className={styles.loadingIndicator}>
                <div className={styles.spinner}></div>
                <span>Загружаем больше событий...</span>
              </div>
            )}
          </div>

          {/* Кнопки прокрутки */}
          {canScrollUp && (
            <button 
              className={`${styles.scrollIndicator} ${styles.scrollUp}`}
              onClick={scrollToTop}
              aria-label="Прокрутить вверх"
            >
              <span style={{ fontSize: '20px', color: '#316BC2' }}>↑</span>
            </button>
          )}
          
          {canScrollDown && (
            <button 
              className={`${styles.scrollIndicator} ${styles.scrollDown}`}
              onClick={scrollToBottom}
              aria-label="Прокрутить вниз"
            >
              <span style={{ fontSize: '20px', color: '#316BC2' }}>↓</span>
            </button>
          )}
        </div>
      </div>
    </div>
  );
}