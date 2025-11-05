'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { useSearchParams } from "next/navigation"
import Image from 'next/image';
import styles from '../../../styles/search/eventsSearch.module.css';
import EventCard from '../../../components/Events/EventCard';
import { EventCardProps } from '../../../types/Events';
import CategoryLabel from "../../../components/Events/CategoryLabel"
import { getAccesToken, convertSessionsToEventCards, convertCategoriesToIds, convertCategRuToEng } from '@/Constants';
import { showNotification } from '@/utils';
import LoadingIndicator from '@/components/LoadingIndicator';
import { searchEvents } from '@/api/search/searchEvents';
import { yandexMapsAPI } from '@/lib/api/index';

// Маппинг категорий к паттернам фоновых изображений
const categoryPatterns: Record<string, string> = {
  'Фильмы': '/events/movie_bg.png',
  'Игры': '/events/game_bg.png',
  'Настольные игры': '/events/board_bg.png',
  'Другое': '/events/other_bg.png'
};

export default function EventsSearchPage() {
  const searchParams = useSearchParams()
  const initialCategory = searchParams?.get("category") || '';
  const initialPattern = searchParams?.get("pattern") || '';
  
  const [searchQuery, setSearchQuery] = useState('');
  const [events, setEvents] = useState<EventCardProps[]>([]);
  const [isFilterOpen, setIsFilterOpen] = useState(false);
  const [selectedCategorySort, setSelectedCategorySort] = useState(initialCategory || 'Все');
  const [currentCategory, setCurrentCategory] = useState(initialCategory);
  const [currentPattern, setCurrentPattern] = useState(initialPattern);
  const [selectedDateSort, setSelectedDateSort] = useState('По возрастанию');
  const [selectedParticipantSort, setSelectedParticipantSort] = useState('По возрастанию');
  const [selectedLocation, setSelectedLocation] = useState('Все');
  const [selectedCity, setSelectedCity] = useState('');
  const [cityQuery, setCityQuery] = useState('');
  const [citySuggestions, setCitySuggestions] = useState<string[]>([]);
  const [showCitySuggestions, setShowCitySuggestions] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [canScrollUp, setCanScrollUp] = useState(false);
  const [canScrollDown, setCanScrollDown] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  
  const filterRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const searchTimeoutRef = useRef<NodeJS.Timeout>();
  const citySearchTimeoutRef = useRef<NodeJS.Timeout>();

  // Маппинг категорий для categoryID
  const getCategoryID = (categoryName: string): number | undefined => {
    const typeArray = convertCategRuToEng([categoryName]);
    if (typeArray.length > 0) {
      const ids = convertCategoriesToIds([typeArray[0]]);
      return ids[0];
    }
    return undefined;
  };

  // Обработчик изменения категории
  const handleCategoryChange = (category: string) => {
    setSelectedCategorySort(category);
    
    if (category !== 'Все') {
      setCurrentCategory(category);
      setCurrentPattern(categoryPatterns[category] || '');
    } else {
      setCurrentCategory('');
      setCurrentPattern('');
    }
  };

  // Проверка возможности прокрутки
  const checkScrollButtons = () => {
    if (containerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
      setCanScrollUp(scrollTop > 0);
      setCanScrollDown(scrollTop < scrollHeight - clientHeight - 1);
    }
  };

  // Поиск городов через Яндекс.Карты
  useEffect(() => {
    if (cityQuery.length >= 2) {
      if (citySearchTimeoutRef.current) {
        clearTimeout(citySearchTimeoutRef.current);
      }

      citySearchTimeoutRef.current = setTimeout(async () => {
        try {
          const cities = await yandexMapsAPI.searchCities(cityQuery);
          setCitySuggestions(cities);
          setShowCitySuggestions(true);
        } catch (error) {
          console.error('Ошибка поиска городов:', error);
          setCitySuggestions([]);
        }
      }, 300);
    } else {
      setCitySuggestions([]);
      setShowCitySuggestions(false);
    }

    return () => {
      if (citySearchTimeoutRef.current) {
        clearTimeout(citySearchTimeoutRef.current);
      }
    };
  }, [cityQuery]);

  // Функция загрузки событий
  const loadEvents = useCallback(async (page: number, resetEvents: boolean = false) => {
    if (isLoading || (!hasMore && !resetEvents)) return;

    setIsLoading(true);
    if (resetEvents) {
      setIsInitialLoading(true);
    }

    try {
      const accessToken = getAccesToken();
      
      const params: any = {};

      if (searchQuery.trim()) {
        params.query = searchQuery.trim();
      }

      // Фильтр по категории через categoryID
      if (selectedCategorySort !== 'Все') {
        const categoryID = getCategoryID(selectedCategorySort);
        if (categoryID) {
          params.categoryID = categoryID;
        }
      }

      if (selectedLocation !== 'Все') {
        params.sessionType = selectedLocation === 'Онлайн' ? 'Онлайн' : 'Оффлайн';
      }

      if (selectedCity.trim()) {
        params.city = selectedCity.trim();
      }

      // Сортировка: приоритет участники > дата
      if (selectedParticipantSort !== 'По возрастанию') {
        params.sort_by = 'users';
        params.order = selectedParticipantSort === 'По возрастанию' ? 'asc' : 'desc';
      } else if (selectedDateSort !== 'По возрастанию') {
        params.sort_by = 'date';
        params.order = selectedDateSort === 'По возрастанию' ? 'asc' : 'desc';
      } else {
        params.sort_by = 'date';
        params.order = 'asc';
      }

      const response = await searchEvents(accessToken, page, params);

      if (response && response.sessions) {
        const validSessions = response.sessions.filter((session: any) => 
          session && session.id && session.title
        );
        
        const convertedEvents = convertSessionsToEventCards(validSessions);
        
        if (resetEvents) {
          setEvents(convertedEvents);
        } else {
          setEvents(prev => [...prev, ...convertedEvents]);
        }

        setHasMore(response.has_next || false);
        setCurrentPage(page);
      } else {
        if (resetEvents) {
          setEvents([]);
        }
        setHasMore(false);
      }

    } catch (error: any) {
      console.error('Ошибка при загрузке событий:', error);
      showNotification(
        error.response?.status || 500,
        error.response?.data?.message || 'Не удалось загрузить события'
      );
      
      if (resetEvents) {
        setEvents([]);
      }
      setHasMore(false);
    } finally {
      setIsLoading(false);
      setIsInitialLoading(false);
    }
  }, [searchQuery, selectedCategorySort, selectedDateSort, selectedParticipantSort, selectedLocation, selectedCity, isLoading, hasMore]);

  // Загрузка следующей страницы
  const loadMoreEvents = useCallback(() => {
    if (!isLoading && hasMore) {
      loadEvents(currentPage + 1, false);
    }
  }, [currentPage, isLoading, hasMore, loadEvents]);

  // Начальная загрузка и перезагрузка при изменении фильтров
  useEffect(() => {
    setCurrentPage(1);
    setHasMore(true);
    loadEvents(1, true);
  }, [selectedCategorySort, selectedDateSort, selectedParticipantSort, selectedLocation, selectedCity]);

  // Поиск с задержкой
  useEffect(() => {
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }

    searchTimeoutRef.current = setTimeout(() => {
      setCurrentPage(1);
      setHasMore(true);
      loadEvents(1, true);
    }, 500);

    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, [searchQuery]);

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

  // Бесконечный скролл
  useEffect(() => {
    const handleScroll = () => {
      checkScrollButtons();
      
      if (containerRef.current) {
        const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
        
        if (scrollHeight - scrollTop <= clientHeight + 100 && !isLoading && hasMore) {
          loadMoreEvents();
        }
      }
    };

    const container = containerRef.current;
    if (container) {
      container.addEventListener('scroll', handleScroll);
      checkScrollButtons();
      
      return () => container.removeEventListener('scroll', handleScroll);
    }
  }, [loadMoreEvents, isLoading, hasMore]);

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

  // Обработчик выбора города
  const handleCitySelect = (city: string) => {
    setSelectedCity(city);
    setCityQuery(city);
    setShowCitySuggestions(false);
  };

  // Очистка города
  const handleClearCity = () => {
    setSelectedCity('');
    setCityQuery('');
    setShowCitySuggestions(false);
  };

  return (
    <div className="bgPage">
      <div className={styles.container}>
        <div className={styles.contentWrapper}>
          {selectedCategorySort !== 'Все' && (
            <div className={styles.categoryHeader}>
              <CategoryLabel 
                  title={currentCategory} 
                  patternUrl={currentPattern} 
                  clickable={false}
              />
            </div>
          )}

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
                    <div className={styles.filterSection}>
                      <h4 className={styles.filterTitle}>Категория</h4>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="categories"
                          checked={selectedCategorySort === 'Все'}
                          onChange={() => handleCategoryChange('Все')}
                        />
                        <span>Все</span>
                      </label>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="categories"
                          checked={selectedCategorySort === 'Фильмы'}
                          onChange={() => handleCategoryChange('Фильмы')}
                        />
                        <span>Фильмы</span>
                      </label>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="categories"
                          checked={selectedCategorySort === 'Игры'}
                          onChange={() => handleCategoryChange('Игры')}
                        />
                        <span>Игры</span>
                      </label>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="categories"
                          checked={selectedCategorySort === 'Настольные игры'}
                          onChange={() => handleCategoryChange('Настольные игры')}
                        />
                        <span>Настольные игры</span>
                      </label>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="categories"
                          checked={selectedCategorySort === 'Другое'}
                          onChange={() => handleCategoryChange('Другое')}
                        />
                        <span>Другое</span>
                      </label>
                    </div>
                    
                    <div className={styles.filterSection}>
                      <h4 className={styles.filterTitle}>Тип проведения</h4>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="location"
                          checked={selectedLocation === 'Все'}
                          onChange={() => setSelectedLocation('Все')}
                        />
                        <span>Все</span>
                      </label>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="location"
                          checked={selectedLocation === 'Онлайн'}
                          onChange={() => setSelectedLocation('Онлайн')}
                        />
                        <span>Онлайн</span>
                      </label>
                      <label className={styles.filterOption}>
                        <input
                          type="radio"
                          name="location"
                          checked={selectedLocation === 'Оффлайн'}
                          onChange={() => setSelectedLocation('Оффлайн')}
                        />
                        <span>Оффлайн</span>
                      </label>
                    </div>

                    <div className={styles.filterSection}>
                      <h4 className={styles.filterTitle}>Город</h4>
                      <div style={{ position: 'relative' }}>
                        <input
                          type="text"
                          value={cityQuery}
                          onChange={(e) => setCityQuery(e.target.value)}
                          placeholder="Начните вводить город..."
                          className={styles.cityInput}
                        />
                        {selectedCity && (
                          <button
                            onClick={handleClearCity}
                            className={styles.clearCityButton}
                            title="Очистить"
                          >
                            ×
                          </button>
                        )}
                        {showCitySuggestions && citySuggestions.length > 0 && (
                          <div className={styles.citySuggestions}>
                            {citySuggestions.map((city, index) => (
                              <div
                                key={index}
                                onClick={() => handleCitySelect(city)}
                                className={styles.citySuggestionItem}
                              >
                                {city}
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>
                    
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

          {isInitialLoading ? (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
              <LoadingIndicator text="Загрузка событий..." />
            </div>
          ) : (
            <>
              <div className={styles.resultsContainer} ref={containerRef}>
                {events.length > 0 ? (
                  <>
                    <div className={styles.eventsGrid}>
                      {events.map((event) => (
                        <div key={`${event.id}-${event.groupId}`} className={styles.eventWrapper}>
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
                  </>
                ) : (
                  <div style={{ 
                    textAlign: 'center', 
                    padding: '60px 20px', 
                    color: '#666',
                    fontSize: '18px'
                  }}>
                    Событий не найдено
                  </div>
                )}
              </div>

              {canScrollUp && events.length > 0 && (
                <button 
                  className={`${styles.scrollIndicator} ${styles.scrollUp}`}
                  onClick={scrollToTop}
                  aria-label="Прокрутить вверх"
                >
                  <span style={{ fontSize: '20px', color: '#316BC2' }}>↑</span>
                </button>
              )}
              
              {canScrollDown && events.length > 0 && (
                <button 
                  className={`${styles.scrollIndicator} ${styles.scrollDown}`}
                  onClick={scrollToBottom}
                  aria-label="Прокрутить вниз"
                >
                  <span style={{ fontSize: '20px', color: '#316BC2' }}>↓</span>
                </button>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}