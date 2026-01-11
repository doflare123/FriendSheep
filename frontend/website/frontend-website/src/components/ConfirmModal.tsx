'use client';

import React from 'react';
import styles from '@/styles/ConfirmModal.module.css';

interface ConfirmModalProps {
  isOpen: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  url?: string;
  searchQuery?: string;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({ 
  isOpen, 
  onConfirm, 
  onCancel, 
  url,
  searchQuery 
}) => {
  if (!isOpen) return null;
  
  const isSearchMode = !!searchQuery;
  
  return (
    <div className={styles.overlay} onClick={onCancel}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <h3 className={styles.title}>
          {isSearchMode ? 'Переход в поисковую систему' : 'Переход по внешней ссылке'}
        </h3>
        
        {isSearchMode ? (
          <>
            <p className={styles.message}>
              Вы уверены, что хотите перейти в поисковую систему для поиска информации?
            </p>
            <p className={styles.searchQuery}>
              Поисковый запрос: <strong>{searchQuery}</strong>
            </p>
          </>
        ) : (
          <>
            <p className={styles.message}>
              Вы уверены, что хотите перейти по внешней ссылке?
            </p>
            <p className={styles.url}>{url}</p>
          </>
        )}
        
        <div className={styles.buttons}>
          <button 
            className={styles.confirmButton}
            onClick={onConfirm}
          >
            Да, перейти
          </button>
          <button 
            className={styles.cancelButton}
            onClick={onCancel}
          >
            Отмена
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;