import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import NewsCard from './NewsCard';
import styles from '@/styles/news/RelatedNews.module.css';
import { NewsItem } from '@/types/news';

interface RelatedNewsProps {
  currentNewsId: number;
}

// Функция для генерации тестовых данных (копия из page.tsx)
const generateTestNewsData = (): NewsItem[] => {
  const TEST_DATES = [
    "21.01.2008", "15.03.2009", "07.11.2010", "23.06.2011", "12.09.2012",
    "04.02.2013", "18.07.2014", "25.12.2015", "09.05.2016", "31.08.2017"
  ];

  return Array.from({ length: 10 }, (_, i) => {
    const id = i + 1;
    return {
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
    };
  });
};

export default function RelatedNews({ currentNewsId }: RelatedNewsProps) {
  const [relatedNews, setRelatedNews] = useState<NewsItem[]>([]);
  const router = useRouter();

  useEffect(() => {
    // Получаем первые 3 новости, исключая текущую
    const allNews = generateTestNewsData();
    const filtered = allNews
      .filter(news => news.id !== currentNewsId)
      .slice(0, 3);
    
    setRelatedNews(filtered);
  }, [currentNewsId]);

  const handleNewsClick = (newsItem: NewsItem) => {
    router.push(`/news/info/${newsItem.id}`);
  };

  return (
    <div className={styles.relatedNewsSection}>
      <h3 className={styles.sectionTitle}>Другие новости:</h3>
      <div className={styles.relatedNewsGrid}>
        {relatedNews.map((newsItem) => (
          <NewsCard
            key={newsItem.id}
            newsItem={newsItem}
            onClick={() => handleNewsClick(newsItem)}
          />
        ))}
      </div>
    </div>
  );
}