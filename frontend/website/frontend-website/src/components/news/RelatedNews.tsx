import { useState, useEffect } from 'react';
import NewsCard from './NewsCard';
import LoadingIndicator from '@/components/LoadingIndicator';
import styles from '@/styles/news/RelatedNews.module.css';
import { NewsItem } from '@/types/news';
import { getNewsInfo } from '@/api/news/GetNewsInfo';
import { showNotification } from '@/utils';

interface RelatedNewsProps {
  relatedNewsIds: number[];
  currentNewsId: number;
}

export default function RelatedNews({ relatedNewsIds, currentNewsId }: RelatedNewsProps) {
  const [relatedNews, setRelatedNews] = useState<NewsItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (relatedNewsIds.length === 0) {
      setRelatedNews([]);
      return;
    }

    loadRelatedNews();
  }, [relatedNewsIds]);

  const loadRelatedNews = async () => {
    setIsLoading(true);
    const newsPromises = relatedNewsIds.map(id => getNewsInfo(id));
    
    try {
      const newsData = await Promise.all(newsPromises);
      // Преобразуем NewsDetail в NewsItem для совместимости с NewsCard
      const newsItems: NewsItem[] = newsData.map(detail => ({
        ...detail,
        createdAt: '',
        updatedAt: '',
        comments: detail.comments || [],
      }));
      setRelatedNews(newsItems);
    } catch (error: any) {
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Не удалось загрузить связанные новости';
      showNotification(statusCode, errorMessage);
      setRelatedNews([]);
    } finally {
      setIsLoading(false);
    }
  };

  const getNewRelatedIds = (clickedNewsId: number): number[] => {
    // Создаем новый список: убираем ID нажатой новости и добавляем ID текущей новости
    const newIds = relatedNewsIds.filter(id => id !== clickedNewsId);
    
    // Добавляем currentNewsId в начало списка, если его там еще нет
    if (!newIds.includes(currentNewsId)) {
      newIds.unshift(currentNewsId);
    }
    
    // Ограничиваем до 3 новостей
    return newIds.slice(0, 3);
  };

  if (isLoading) {
    return (
      <div className={styles.relatedNewsSection}>
        <h3 className={styles.sectionTitle}>Другие новости:</h3>
        <LoadingIndicator text="Загружаем другие новости..." />
      </div>
    );
  }

  if (relatedNews.length === 0) {
    return null;
  }

  return (
    <div className={styles.relatedNewsSection}>
      <h3 className={styles.sectionTitle}>Другие новости:</h3>
      <div className={styles.relatedNewsGrid}>
        {relatedNews.map((newsItem) => (
          <NewsCard
            key={newsItem.id}
            newsItem={newsItem}
            relatedNewsIds={getNewRelatedIds(newsItem.id)}
          />
        ))}
      </div>
    </div>
  );
}