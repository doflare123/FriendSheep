import { useState } from 'react';
import Image from 'next/image';
import styles from '@/styles/news/CommentsList.module.css';
import {formatTime} from '@/Constants'
import { Comment } from '@/types/news';

interface CommentsListProps {
  comments: Comment[];
}

interface CommentItemProps {
  comment: Comment;
}

function CommentItem({ comment }: CommentItemProps) {
  const [imageError, setImageError] = useState(false);

  const handleImageError = () => {
    setImageError(true);
  };

  return (
    <div className={styles.commentItem}>
      <div className={styles.avatarContainer}>
        <Image
          src={imageError ? '/default-avatar.png' : comment.image}
          alt={comment.name}
          width={40}
          height={40}
          className={styles.avatar}
          onError={handleImageError}
        />
      </div>
      
      <div className={styles.commentContent}>
        <div className={styles.commentHeader}>
          <span className={styles.userName}>{comment.name}</span>
          <span className={styles.commentDate}>4 мая</span>
        </div>
        <div className={styles.commentText}>
          {comment.text}
        </div>
        <div className={styles.commentTime}>
          {formatTime(comment.created_time) || "14:56"}
        </div>
      </div>
    </div>
  );
}

export default function CommentsList({ comments }: CommentsListProps) {
  return (
    <div className={styles.commentsList}>
      {comments.map((comment) => (
        <CommentItem key={comment.id} comment={comment} />
      ))}
    </div>
  );
}