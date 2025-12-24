'use client';

import { useState, useEffect } from 'react';
import styles from '../styles/MainPage.module.css';
import { SectionData } from "../types/Events";
import CategorySection from '../components/Events/CategorySection';
import LoadingIndicator from '@/components/LoadingIndicator';
import { getGenreEvents } from '@/api/mainEvents/getGenreEvents';
import { getGroupEvents } from '@/api/mainEvents/getGroupEvents';
import { getNewEvents } from '@/api/mainEvents/getNewEvents';
import { getPopular } from '@/api/mainEvents/getPopular';
import { convertCategoriesToIds } from '@/Constants';

interface MainHomeProps {
  token: string;
}

export default function MainHome({ token }: MainHomeProps) {
  const [mainSections, setMainSections] = useState<SectionData[]>([]);
  const [additionalSections, setAdditionalSections] = useState<SectionData[]>([]);
  const [loadingStage, setLoadingStage] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);

  const clickableCategories = ['Медиа', 'Видеоигры', 'Настольные игры', 'Другое'];

  useEffect(() => {
    const loadData = async () => {
      const newMainSections: SectionData[] = [];
      const newAdditionalSections: SectionData[] = [];

      const filterValidEvents = (events: any[]) => {
        return events.filter(event => event.image && event.image.trim() !== '');
      };

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

      const categories = [
        { name: 'Медиа', key: 'movies', pattern: '/events/new_bg.png' },
        { name: 'Видеоигры', key: 'games', pattern: '/events/game_bg.png' },
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
  }, [token]);

  const hasAnyEvents = mainSections.length > 0 || additionalSections.length > 0;

  return (
    <div className={styles.contentWrapper}>
      {!isLoading && !hasAnyEvents && (
        <div className={styles.emptyMessage}>
          <p>Пока нет никаких событий</p>
        </div>
      )}

      {mainSections.map((section, sectionIndex) => (
        <div key={sectionIndex} className={styles.section}>
          <CategorySection 
            section={section} 
            title={section.title}
            clickable={false}
          />
        </div>
      ))}

      {isLoading && mainSections.length > 0 && (
        <div className={styles.section}>
          <LoadingIndicator text={loadingStage} />
        </div>
      )}

      {additionalSections.length > 0 && (
        <div className={styles.categoriesHeader}>
          <h2>Категории</h2>
        </div>
      )}

      {additionalSections.map((section, sectionIndex) => (
        <div key={sectionIndex} className={styles.section}>
          <CategorySection 
            section={section} 
            title={section.title}
            clickable={clickableCategories.includes(section.title)}
          />
        </div>
      ))}

      {isLoading && mainSections.length === 0 && (
        <div className={styles.section}>
          <LoadingIndicator text={loadingStage} />
        </div>
      )}
    </div>
  );
}