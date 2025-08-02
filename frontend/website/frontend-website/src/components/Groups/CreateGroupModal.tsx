import React from 'react';
import CreateGroupForm from './CreateGroupForm';
import styles from '../../styles/Groups/CreateGroupModal.module.css';

interface CreateGroupModalProps {
  onClose: () => void;
  onSubmit: (groupData: any) => void;
}

const CreateGroupModal: React.FC<CreateGroupModalProps> = ({ onClose, onSubmit }) => {
  return (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div 
        className={styles.modalContent} 
        onClick={(e) => e.stopPropagation()}
        style={{
          background: 'white',
          borderRadius: '20px',
          boxShadow: '0 10px 30px rgba(0, 0, 0, 0.3)'
        }}
      >
        <div className={styles.modalHeader}>
          <h2>Основная информация</h2>
          <button className={styles.closeButton} onClick={onClose}>×</button>
        </div>
        <div className={styles.modalBody}>
          <CreateGroupForm onSubmit={onSubmit} />
        </div>
      </div>
    </div>
  );
};

export default CreateGroupModal;