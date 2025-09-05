// src/components/news/NewsCard.tsx
import { useState } from 'react';
import Image from 'next/image';
import styles from '@/styles/news/NewsCard.module.css';
import { NewsItem } from '@/app/news/page';

interface NewsCardProps {
  newsItem: NewsItem;
  onClick: () => void;
}

export default function NewsCard({ newsItem, onClick }: NewsCardProps) {
  const [imageError, setImageError] = useState(false);

  const handleImageError = () => {
    setImageError(true);
  };

  return (
    <div className={styles.newsCard} onClick={onClick}>
      <div className={styles.imageContainer}>
        <Image
          src={imageError ? '/default/news.png' : newsItem.image}
          alt={newsItem.title}
          fill
          className={styles.newsImage}
          onError={handleImageError}
        />
      </div>
      
      <div className={styles.cardContent}>
        <h3 className={styles.title}>{newsItem.title}</h3>
        <p className={styles.description}>{newsItem.description}</p>
        <div className={styles.date}>{newsItem.created_time}</div>
      </div>
    </div>
  );
}