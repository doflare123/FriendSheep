'use client';

import { useState, useEffect } from 'react';
import { useParams, useSearchParams } from 'next/navigation';
import Image from 'next/image';
import NewsDetailCard from '@/components/news/NewsDetailCard';
import CommentsList from '@/components/news/CommentsList';
import CommentForm from '@/components/news/CommentForm';
import RelatedNews from '@/components/news/RelatedNews';
import LoadingIndicator from '@/components/LoadingIndicator';
import styles from '@/styles/news/newsInfo.module.css';
import { NewsDetail, Comment } from '@/types/news';
import { getNewsInfo } from '@/api/news/GetNewsInfo';
import { addNewsComment } from '@/api/news/addNewsComment';
import { showNotification } from '@/utils';
import { getAccesToken } from '@/Constants';

export default function NewsInfoPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const id = params?.id as string;
  
  const [newsDetail, setNewsDetail] = useState<NewsDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showCommentForm, setShowCommentForm] = useState(false);
  const [visibleComments, setVisibleComments] = useState(5);
  const [isSubmittingComment, setIsSubmittingComment] = useState(false);
  const [relatedNewsIds, setRelatedNewsIds] = useState<number[]>([]);

  useEffect(() => {
    if (id) {
      loadNewsDetail();
      
      // Получаем ID связанных новостей из query параметров
      const relatedParam = searchParams.get('related');
      if (relatedParam) {
        const ids = relatedParam.split(',').map(id => parseInt(id)).filter(id => !isNaN(id));
        setRelatedNewsIds(ids);
      }
    }
  }, [id, searchParams]);

  const loadNewsDetail = async () => {
    try {
      setIsLoading(true);
      const data = await getNewsInfo(parseInt(id));
      setNewsDetail(data);
    } catch (error: any) {
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Не удалось загрузить новость';
      showNotification(statusCode, errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCommentSubmit = async (commentText: string) => {
    if (!newsDetail || !commentText.trim()) return;

    // Проверка авторизации
    const token = getAccesToken();
    if (!token) {
      showNotification(401, 'Необходимо авторизоваться');
      setShowCommentForm(false);
      return;
    }

    try {
      setIsSubmittingComment(true);
      await addNewsComment(token, newsDetail.id, commentText);
      
      // Перезагружаем новость чтобы получить обновленные комментарии
      await loadNewsDetail();
      setShowCommentForm(false);
      showNotification(200, 'Комментарий успешно добавлен');
    } catch (error: any) {
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Не удалось добавить комментарий';
      showNotification(statusCode, errorMessage);
    } finally {
      setIsSubmittingComment(false);
    }
  };

  const loadMoreComments = () => {
    setVisibleComments(prev => Math.min(prev + 5, newsDetail?.comments?.length || 0));
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

  const hasMoreComments = visibleComments < (newsDetail.comments?.length || 0);

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
              disabled={isSubmittingComment}
            >
              Прокомментировать
            </button>
          </div>

          {showCommentForm && (
            <CommentForm 
              onSubmit={handleCommentSubmit}
              onCancel={() => setShowCommentForm(false)}
              disabled={isSubmittingComment}
            />
          )}

          <CommentsList 
            comments={newsDetail.comments?.slice(0, visibleComments) || []}
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

        <RelatedNews relatedNewsIds={relatedNewsIds} currentNewsId={parseInt(id)} />
      </div>
    </div>
  );
}