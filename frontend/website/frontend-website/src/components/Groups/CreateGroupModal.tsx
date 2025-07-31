import React from 'react';
import CreateGroupForm from './CreateGroupForm';
import '../../styles/Groups/CreateGroupModal.css';

interface CreateGroupModalProps {
  onClose: () => void;
  onSubmit: (groupData: any) => void;
}

const CreateGroupModal: React.FC<CreateGroupModalProps> = ({ onClose, onSubmit }) => {
  return (
    <div className="modalOverlay" onClick={onClose}>
      <div 
        className="modalContent" 
        onClick={(e) => e.stopPropagation()}
        style={{
          background: 'white',
          borderRadius: '20px',
          boxShadow: '0 10px 30px rgba(0, 0, 0, 0.3)'
        }}
      >
        <div className="modalHeader">
          <h2>Основная информация</h2>
          <button className="closeButton" onClick={onClose}>×</button>
        </div>
        <div className="modalBody">
          <CreateGroupForm onSubmit={onSubmit} />
        </div>
      </div>
    </div>
  );
};

export default CreateGroupModal;