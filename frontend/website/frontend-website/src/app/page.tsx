// app/page.tsx
'use client';

import { useState, useEffect } from 'react';
import Footer from "../components/Footer";
import styles from '../styles/MainPage.module.css';
import { SectionData } from "../types/Events";
import CategorySection from '../components/Events/CategorySection';
import LoadingIndicator from '@/components/LoadingIndicator';
import { getGenreEvents } from '@/api/mainEvents/getGenreEvents';
import { getGroupEvents } from '@/api/mainEvents/getGroupEvents';
import { getNewEvents } from '@/api/mainEvents/getNewEvents';
import { getPopular } from '@/api/mainEvents/getPopular';
import { getAccesToken, convertCategoriesToIds } from '@/Constants';

export default function Home() {
  const [mainSections, setMainSections] = useState<SectionData[]>([]);
  const [additionalSections, setAdditionalSections] = useState<SectionData[]>([]);
  const [loadingStage, setLoadingStage] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  // Список категорий, которые должны быть кликабельными
  const clickableCategories = ['Фильмы', 'Игры', 'Настольные игры', 'Другое'];

  useEffect(() => {
    const loadData = async () => {
      const token = await getAccesToken();
      
      // Если нет токена - просто не загружаем данные
      if (!token) {
        setIsAuthenticated(false);
        setIsLoading(false);
        return;
      }

      setIsAuthenticated(true);

      const newMainSections: SectionData[] = [];
      const newAdditionalSections: SectionData[] = [];

      // Функция для фильтрации событий с пустыми изображениями
      const filterValidEvents = (events: any[]) => {
        return events.filter(event => event.image && event.image.trim() !== '');
      };

      // 1. Загружаем популярные события
      setLoadingStage('Загружаем популярные события...');
      try {
        const popularEvents = await getPopular();
        const validPopularEvents = filterValidEvents(popularEvents || []);
        if (validPopularEvents.length > 0) {
          newMainSections.push({
            title: "Популярные события",
            pattern: "/events/popular_bg.png",
            categories: validPopularEvents
          });
        }
      } catch (error) {
        console.error('Ошибка загрузки популярных событий:', error);
      }

      // 2. Загружаем новые события
      setLoadingStage('Загружаем новые события...');
      try {
        const newEvents = await getNewEvents(token);
        const validNewEvents = filterValidEvents(newEvents || []);
        if (validNewEvents.length > 0) {
          newMainSections.push({
            title: "Новые события",
            pattern: "/events/new_bg.png",
            categories: validNewEvents
          });
        }
      } catch (error) {
        console.error('Ошибка загрузки новых событий:', error);
      }

      // 3. Загружаем новые события из подписанных групп
      setLoadingStage('Загружаем события из ваших подписок...');
      try {
        const groupEvents = await getGroupEvents(token, 1);
        const validGroupEvents = filterValidEvents(groupEvents || []);
        if (validGroupEvents.length > 0) {
          newMainSections.push({
            title: "Новые события в ваших подписках",
            pattern: "/events/new_bg.png",
            categories: validGroupEvents
          });
        }
      } catch (error) {
        console.error('Ошибка загрузки событий из групп:', error);
      }

      setMainSections(newMainSections);

      // 4. Загружаем события по категориям
      const categories = [
        { name: 'Фильмы', key: 'movies', pattern: '/events/new_bg.png' },
        { name: 'Игры', key: 'games', pattern: '/events/game_bg.png' },
        { name: 'Настольные игры', key: 'board', pattern: '/events/board_bg.png' },
        { name: 'Другое', key: 'other', pattern: '/events/other_bg.png' }
      ];

      for (const category of categories) {
        setLoadingStage(`Загружаем категорию "${category.name}"...`);
        try {
          const categoryIds = convertCategoriesToIds([category.key]);
          if (categoryIds.length > 0) {
            const categoryEvents = await getGenreEvents(token, categoryIds[0]);
            const validCategoryEvents = filterValidEvents(categoryEvents || []);
            if (validCategoryEvents.length > 0) {
              newAdditionalSections.push({
                title: category.name,
                pattern: category.pattern,
                categories: validCategoryEvents
              });
            }
          }
        } catch (error) {
          console.error(`Ошибка загрузки категории ${category.name}:`, error);
        }
      }

      setAdditionalSections(newAdditionalSections);
      setIsLoading(false);
      setLoadingStage('');
    };

    loadData();
  }, []);

  return (
    <div className={styles.pageWrapper}>
      <div className='bgPage'>
        <div className={styles.contentWrapper}>
          {/* Сообщение для неавторизованных пользователей */}
          {!isAuthenticated && !isLoading && (
            <div className={styles.emptyMessage}>
              <p>Для просмотра всех событий нужна авторизация</p>
            </div>
          )}

          {/* Главные категории */}
          {mainSections.map((section, sectionIndex) => (
            <div key={sectionIndex} className={styles.section}>
              <CategorySection 
                section={section} 
                title={section.title}
                clickable={false}
              />
            </div>
          ))}

          {/* Индикатор загрузки между секциями */}
          {isLoading && mainSections.length > 0 && (
            <div className={styles.section}>
              <LoadingIndicator text={loadingStage} />
            </div>
          )}

          {/* Заголовок "Категории" */}
          {additionalSections.length > 0 && (
            <div className={styles.categoriesHeader}>
              <h2>Категории</h2>
            </div>
          )}

          {/* Дополнительные категории */}
          {additionalSections.map((section, sectionIndex) => (
            <div key={sectionIndex} className={styles.section}>
              <CategorySection 
                section={section} 
                title={section.title}
                clickable={clickableCategories.includes(section.title)}
              />
            </div>
          ))}

          {/* Индикатор загрузки в начале */}
          {isLoading && mainSections.length === 0 && isAuthenticated && (
            <div className={styles.section}>
              <LoadingIndicator text={loadingStage} />
            </div>
          )}
        </div>
      </div>
      <div className={styles.footerWrapper}>
        <Footer />
      </div>
    </div>
  );
}