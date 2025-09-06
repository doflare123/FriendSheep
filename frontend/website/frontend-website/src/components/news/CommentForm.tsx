import { useState } from 'react';
import styles from '@/styles/news/CommentForm.module.css';

interface CommentFormProps {
  onSubmit: (text: string) => void;
  onCancel: () => void;
}

export default function CommentForm({ onSubmit, onCancel }: CommentFormProps) {
  const [commentText, setCommentText] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (commentText.trim()) {
      onSubmit(commentText.trim());
      setCommentText('');
    }
  };

  const handleCancel = () => {
    setCommentText('');
    onCancel();
  };

  return (
    <div className={styles.commentForm}>
      <form onSubmit={handleSubmit}>
        <textarea
          className={styles.commentInput}
          placeholder="Введите свой комментарий здесь..."
          value={commentText}
          onChange={(e) => setCommentText(e.target.value)}
          rows={3}
          maxLength={500}
        />
        
        <div className={styles.formButtons}>
          <button 
            type="button" 
            className={styles.cancelButton}
            onClick={handleCancel}
          >
            Отменить
          </button>
          <button 
            type="submit" 
            className={styles.submitButton}
            disabled={!commentText.trim()}
          >
            Отправить
          </button>
        </div>
      </form>
    </div>
  );
}