// src/components/Groups/DeleteConfirmModal.tsx

import React from 'react';
import styles from '@/styles/Groups/DeleteConfirmModal.module.css';

interface DeleteConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  step: 1 | 2;
  groupName: string;
}

const DeleteConfirmModal: React.FC<DeleteConfirmModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  step,
  groupName
}) => {
  if (!isOpen) return null;

  return (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
        <div className={styles.modalHeader}>
          <h3>Удаление группы</h3>
          <button className={styles.closeButton} onClick={onClose}>
            ×
          </button>
        </div>
        
        <div className={styles.modalBody}>
          {step === 1 ? (
            <>
              <p className={styles.warningText}>
                Вы уверены, что хотите удалить группу <strong>"{groupName}"</strong>?
              </p>
              <p className={styles.infoText}>
                Это действие нельзя будет отменить.
              </p>
            </>
          ) : (
            <>
              <p className={styles.warningText}>
                Последнее предупреждение!
              </p>
              <p className={styles.infoText}>
                Группа <strong>"{groupName}"</strong> и все её данные будут безвозвратно удалены.
              </p>
            </>
          )}
        </div>
        
        <div className={styles.modalFooter}>
          <button className={styles.cancelButton} onClick={onClose}>
            Отмена
          </button>
          <button className={styles.deleteButton} onClick={onConfirm}>
            {step === 1 ? 'Продолжить' : 'Удалить окончательно'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default DeleteConfirmModal;