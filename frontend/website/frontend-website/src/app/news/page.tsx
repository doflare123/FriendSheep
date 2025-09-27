'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import NewsCard from '@/components/news/NewsCard';
import LoadingIndicator from '@/components/LoadingIndicator';
import styles from '@/styles/news/newsPage.module.css';
import { NewsItem, NewsResponse } from '@/types/news';

// Тестовые данные - создаем один раз и кешируем
const TEST_DATES = [
  "21.01.2008", "15.03.2009", "07.11.2010", "23.06.2011", "12.09.2012",
  "04.02.2013", "18.07.2014", "25.12.2015", "09.05.2016", "31.08.2017",
  "14.01.2018", "28.04.2019", "06.10.2020", "19.03.2021", "11.07.2022",
  "02.12.2023", "26.02.2024", "08.06.2024", "20.09.2024", "13.11.2024",
  "05.01.2025", "17.03.2025", "29.05.2025", "10.08.2025", "22.10.2025",
  "03.12.2025", "16.02.2026", "30.04.2026", "12.07.2026", "24.09.2026",
  "06.11.2026", "18.01.2027", "01.04.2027", "14.06.2027", "27.08.2027",
  "09.10.2027"
];

const generateTestData = (page: number): NewsResponse => {
  const items: NewsItem[] = [];
  const itemsPerPage = 9;
  
  for (let i = 0; i < itemsPerPage; i++) {
    const id = page * itemsPerPage + i + 1;
    items.push({
      id,
      title: "Нереальный апдейт, смотреть всем!!!!",
      description: "Статья о том, что мы подготовили в новом обновлении для вас - любимых пользователь. Проделав огромную работу, готовы представить вам следующую статью по обновлению",
      image: "/default/news.png",
      created_time: TEST_DATES[id - 1] || "21.01.2008",
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      comments: [],
      content: [
        {
          id: id,
          news_id: id,
          text: "Содержание новости..."
        }
      ]
    });
  }

  return {
    hasMore: page < 3, // Симулируем 4 страницы данных
    items,
    page,
    total: 36
  };
};

export default function NewsPage() {
  const [news, setNews] = useState<NewsItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(0);
  const [initialized, setInitialized] = useState(false);
  const router = useRouter();
  
  // Используем ref для предотвращения множественных запросов
  const loadingRef = useRef(false);

  // Загрузка данных
  const loadNews = async (pageNum: number, isInitial: boolean = false) => {
    if (loadingRef.current) return;
    
    loadingRef.current = true;
    setIsLoading(true);
    
    // Симулируем задержку сети
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const response = generateTestData(pageNum);
    
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
    setIsLoading(false);
    loadingRef.current = false;
  };

  // Загрузка первой страницы только один раз
  useEffect(() => {
    if (!initialized) {
      loadNews(0, true);
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

  const handleCardClick = (newsItem: NewsItem) => {
    router.push(`/news/info/${newsItem.id}`);
  };

  return (
    <div className="bgPage">
      <div className={styles.newsContainer}>
        <div className={styles.newsGrid}>
          {news.map((item) => (
            <NewsCard
              key={item.id}
              newsItem={item}
              onClick={() => handleCardClick(item)}
            />
          ))}
        </div>
        
        {isLoading && (
          <LoadingIndicator text="Загружаем больше новостей..." />
        )}
        
        {!hasMore && news.length > 0 && (
          <div className={styles.endMessage}>
            Все новости загружены
          </div>
        )}
      </div>
    </div>
  );
}