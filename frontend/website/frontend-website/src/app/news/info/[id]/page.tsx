'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Image from 'next/image';
import NewsDetailCard from '@/components/news/NewsDetailCard';
import CommentsList from '@/components/news/CommentsList';
import CommentForm from '@/components/news/CommentForm';
import RelatedNews from '@/components/news/RelatedNews';
import LoadingIndicator from '@/components/LoadingIndicator';
import styles from '@/styles/news/newsInfo.module.css';
import { NewsDetail, Comment } from '@/types/news';

// Тестовые данные
const generateTestNewsDetail = (id: string): NewsDetail => {
  const testComments: Comment[] = Array.from({ length: 12 }, (_, index) => ({
    id: index + 1,
    user_id: index + 1,
    name: "Хлеб хлебушек",
    image: "/default-avatar.png",
    text: "Это просто жесть!!!!!!!! Вы ПРОСТО НЕРЕАЛЬНЫЕ МОЛОДЦЫ, ЛЮБЛЮ ВАС ❤️❤️❤️"
  }));

  return {
    id: parseInt(id),
    title: "Нереальный апдейт, смотреть всем!!!!",
    description: "Просто нереальное обновление. В нем были добавлены:",
    image: "/default/news.png",
    content: `Просто нереальное обновление. В нем были добавлены:
• Новые возможности для админов
• Новое звуковое сопровождение на сайте
• Улучшили обмен информации между серверами, поэтому ожидание стало еще меньше!!!

и т.д. я устал писать`,
    created_time: "21.01.2008",
    comments: testComments
  };
};

export default function NewsInfoPage() {
  const params = useParams();
  const id = params?.id as string;
  
  const [newsDetail, setNewsDetail] = useState<NewsDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showCommentForm, setShowCommentForm] = useState(false);
  const [visibleComments, setVisibleComments] = useState(5);

  useEffect(() => {
    if (id) {
      // Симулируем загрузку данных
      setTimeout(() => {
        const testData = generateTestNewsDetail(id);
        setNewsDetail(testData);
        setIsLoading(false);
      }, 500);
    }
  }, [id]);

  const handleCommentSubmit = (commentText: string) => {
    if (!newsDetail) return;

    const newComment: NewsComment = {
      id: newsDetail.comments.length + 1,
      user_id: 1,
      name: "Текущий пользователь", // В реальном приложении будет из контекста
      image: "/default-avatar.png",
      text: commentText,
      created_time: new Date().toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    };

    setNewsDetail(prev => ({
      ...prev!,
      comments: [newComment, ...prev!.comments]
    }));
    setShowCommentForm(false);
  };

  const loadMoreComments = () => {
    setVisibleComments(prev => Math.min(prev + 5, newsDetail?.comments.length || 0));
  };

  if (isLoading) {
    return (
      <div className="bgPage">
        <LoadingIndicator text="Загружаем новость..." />
      </div>
    );
  }

  if (!newsDetail) {
    return (
      <div className="bgPage">
        <div className={styles.errorContainer}>
          <h2>Новость не найдена</h2>
        </div>
      </div>
    );
  }

  const hasMoreComments = visibleComments < newsDetail.comments.length;

  return (
    <div className="bgPage">
      <div className={styles.newsInfoContainer}>
        <NewsDetailCard newsDetail={newsDetail} />
        
        <div className={styles.commentsSection}>
          <div className={styles.commentsHeader}>
            <h3>Комментарии:</h3>
            <button 
              className={styles.commentButton}
              onClick={() => setShowCommentForm(!showCommentForm)}
            >
              Прокомментировать
            </button>
          </div>

          {showCommentForm && (
            <CommentForm 
              onSubmit={handleCommentSubmit}
              onCancel={() => setShowCommentForm(false)}
            />
          )}

          <CommentsList 
            comments={newsDetail.comments.slice(0, visibleComments)}
          />

          {hasMoreComments && (
            <div className={styles.loadMoreContainer}>
              <button 
                className={styles.loadMoreButton}
                onClick={loadMoreComments}
              >
                Читать далее
              </button>
            </div>
          )}
        </div>

        <RelatedNews currentNewsId={newsDetail.id} />
      </div>
    </div>
  );
}