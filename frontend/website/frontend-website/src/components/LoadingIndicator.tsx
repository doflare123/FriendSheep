// src/components/LoadingIndicator.tsx
import styles from '@/styles/LoadingIndicator.module.css';

interface LoadingIndicatorProps {
  text?: string;
}

export default function LoadingIndicator({ text = "Загружаем..." }: LoadingIndicatorProps) {
  return (
    <div className={styles.loadingIndicator}>
      <div className={styles.spinner}></div>
      <span>{text}</span>
    </div>
  );
}