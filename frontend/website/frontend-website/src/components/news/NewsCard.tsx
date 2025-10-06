import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import styles from '@/styles/news/NewsCard.module.css';
import { NewsItem } from '@/types/news';
import { formatDate } from '@/Constants';

interface NewsCardProps {
  newsItem: NewsItem;
  relatedNewsIds?: number[];
}

export default function NewsCard({ newsItem, relatedNewsIds = [] }: NewsCardProps) {
  const [imageError, setImageError] = useState(false);
  const router = useRouter();

  const handleImageError = () => {
    setImageError(true);
  };

  const handleClick = () => {
    const queryParams = relatedNewsIds.length > 0 
      ? `?related=${relatedNewsIds.join(',')}` 
      : '';
    router.push(`/news/info/${newsItem.id}${queryParams}`);
  };

  return (
    <div className={styles.newsCard} onClick={handleClick}>
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
        <div className={styles.date}>{formatDate(newsItem.created_time)}</div>
      </div>
    </div>
  );
}