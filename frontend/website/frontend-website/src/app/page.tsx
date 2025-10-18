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
import { getEventInfo } from '@/api/events/getEventInfo';
import { getAccesToken, convertCategoriesToIds } from '@/Constants';
import { showNotification } from '@/utils';

export default function Home() {
  const [mainSections, setMainSections] = useState<SectionData[]>([]);
  const [additionalSections, setAdditionalSections] = useState<SectionData[]>([]);
  const [loadingStage, setLoadingStage] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      const token = getAccesToken();
      
      if (!token) {
        showNotification(401, 'Необходима авторизация', 'error');
        setIsLoading(false);
        return;
      }

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
        showNotification(500, 'Не удалось загрузить популярные события', 'error');
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
        showNotification(500, 'Не удалось загрузить новые события', 'error');
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
        showNotification(500, 'Не удалось загрузить события из подписок', 'error');
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
          showNotification(500, `Не удалось загрузить категорию "${category.name}"`, 'error');
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
        {/* Главные категории */}
        {mainSections.map((section, sectionIndex) => (
          <div key={sectionIndex} className={styles.section}>
            <CategorySection 
              section={section} 
              title={section.title} 
              clickable={section.title !== "Новые события"}
            />
          </div>
        ))}

        {/* Индикатор загрузки между секциями */}
        {isLoading && mainSections.length > 0 && (
          <div className={styles.section}>
            <LoadingIndicator text={loadingStage} />
          </div>
        )}

        {/* Заголовок "Категории" - показываем только если есть дополнительные секции или они загружаются */}
        {(additionalSections.length > 0 || isLoading) && (
          <div className={styles.categoriesHeader}>
            <h2>Категории</h2>
          </div>
        )}

        {/* Дополнительные категории */}
        {additionalSections.map((section, sectionIndex) => (
          <div key={sectionIndex} className={styles.section}>
            <CategorySection section={section} title={section.title} />
          </div>
        ))}

        {/* Индикатор загрузки для дополнительных секций */}
        {isLoading && mainSections.length === 0 && (
          <div className={styles.section}>
            <LoadingIndicator text={loadingStage} />
          </div>
        )}
      </div>
      <div className={styles.footerWrapper}>
        <Footer />
      </div>
    </div>
  );
}