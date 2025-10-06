'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import NewsCard from '@/components/news/NewsCard';
import LoadingIndicator from '@/components/LoadingIndicator';
import styles from '@/styles/news/newsPage.module.css';
import { NewsItem, NewsResponse } from '@/types/news';
import {showNotification} from '@/utils';
import {getNews} from '@/api/news/getNews';

export default function NewsPage() {
  const [news, setNews] = useState<NewsItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(1);
  const [initialized, setInitialized] = useState(false);
  const router = useRouter();
  
  // Используем ref для предотвращения множественных запросов
  const loadingRef = useRef(false);

  // Загрузка данных
  const loadNews = async (pageNum: number, isInitial: boolean = false) => {
    if (loadingRef.current) return;
    
    loadingRef.current = true;
    setIsLoading(true);
    
    try {
      const response: NewsResponse = await getNews(pageNum);
      
      setNews(prev => {
        if (isInitial) {
          return response.items;
        } else {
          // Проверяем, чтобы не добавлять дубликаты
          const existingIds = new Set(prev.map(item => item.id));
          const newItems = response.items.filter(item => !existingIds.has(item.id));
          return [...prev, ...newItems];
        }
      });
      
      setHasMore(response.hasMore);
    } catch (error: any) {
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Не удалось загрузить новости';
      showNotification(statusCode, errorMessage);
    } finally {
      setIsLoading(false);
      loadingRef.current = false;
    }
  };

  // Загрузка первой страницы только один раз
  useEffect(() => {
    if (!initialized) {
      loadNews(1, true);
      setInitialized(true);
    }
  }, [initialized]);

  // Обработка скролла для бесконечной загрузки
  useEffect(() => {
    const handleScroll = () => {
      if (loadingRef.current || !hasMore) return;

      const scrollContainer = document.querySelector(`.${styles.newsContainer}`);
      if (!scrollContainer) return;

      const { scrollTop, scrollHeight, clientHeight } = scrollContainer;
      
      // Загружаем новые данные когда осталось прокрутить 200px до конца
      if (scrollHeight - scrollTop - clientHeight < 200) {
        const nextPage = page + 1;
        setPage(nextPage);
        loadNews(nextPage, false);
      }
    };

    const scrollContainer = document.querySelector(`.${styles.newsContainer}`);
    if (scrollContainer) {
      scrollContainer.addEventListener('scroll', handleScroll);
      return () => scrollContainer.removeEventListener('scroll', handleScroll);
    }
  }, [hasMore, page]);

  const getRelatedNewsIds = (currentNewsId: number): number[] => {
    // Находим индекс текущей новости
    const currentIndex = news.findIndex(item => item.id === currentNewsId);
    
    // Получаем ID первых 3 других новостей, исключая текущую
    const relatedIds: number[] = [];
    let count = 0;
    let index = 0;
    
    while (count < 3 && index < news.length) {
      if (index !== currentIndex) {
        relatedIds.push(news[index].id);
        count++;
      }
      index++;
    }
    
    return relatedIds;
  };

  return (
    <div className="bgPage">
      <div className={styles.newsContainer}>
        <div className={styles.newsGrid}>
          {news.map((item) => (
            <NewsCard
              key={item.id}
              newsItem={item}
              relatedNewsIds={getRelatedNewsIds(item.id)}
            />
          ))}
        </div>
        
        {isLoading && (
          <LoadingIndicator text="Загружаем больше новостей..." />
        )}
      </div>
    </div>
  );
}