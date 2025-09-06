import { useState } from 'react';
import Image from 'next/image';
import styles from '@/styles/news/NewsDetailCard.module.css';
import { NewsDetail } from '@/types/news';

interface NewsDetailCardProps {
  newsDetail: NewsDetail;
}

export default function NewsDetailCard({ newsDetail }: NewsDetailCardProps) {
  const [imageError, setImageError] = useState(false);

  const handleImageError = () => {
    setImageError(true);
  };

  // Форматируем контент для отображения
  const formatContent = (content: string) => {
    return content.split('\n').map((line, index) => (
      <p key={index} className={styles.contentLine}>
        {line}
      </p>
    ));
  };

  return (
    <div className={styles.newsDetailCard}>
      <div className={styles.imageContainer}>
        <Image
          src={imageError ? '/default/news.png' : newsDetail.image}
          alt={newsDetail.title}
          fill
          className={styles.newsImage}
          onError={handleImageError}
          priority
        />
      </div>
      
      <div className={styles.cardContent}>
        <h1 className={styles.title}>{newsDetail.title}</h1>
        
        <div className={styles.contentSection}>
          {formatContent(newsDetail.content)}
        </div>
      </div>
    </div>
  );
}