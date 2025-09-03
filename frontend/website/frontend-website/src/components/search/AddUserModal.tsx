'use client';

import { useState, useEffect, useRef } from 'react';
import Image from 'next/image';
import styles from '../../styles/search/addUserModal.module.css';
import { getCategoryIcon } from '../../Constants';

interface Group {
  id: number;
  name: string;
  image: string;
  category: string[];
  count: number;
}

interface AddUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  userId: number | null;
}

// Тестовые данные групп пользователя
const mockUserGroups: Group[] = Array.from({ length: 15 }, (_, i) => ({
  id: i + 1,
  name: `Группа крутых пацанят ${i + 1}`,
  image: '/default/group.jpg',
  category: ['games', 'movies', 'board', 'other'][i % 4] ? [['games', 'movies', 'board', 'other'][i % 4]] : ['games'],
  count: Math.floor(Math.random() * 100) + 10
}));

export default function AddUserModal({ isOpen, onClose, userId }: AddUserModalProps) {
  const [selectedGroups, setSelectedGroups] = useState<Set<number>>(new Set());
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  // Закрытие модалки по Escape
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  // Сброс выбранных групп при открытии модалки
  useEffect(() => {
    if (isOpen) {
      setSelectedGroups(new Set());
    }
  }, [isOpen]);

  const handleGroupToggle = (groupId: number) => {
    setSelectedGroups(prev => {
      const newSet = new Set(prev);
      if (newSet.has(groupId)) {
        newSet.delete(groupId);
      } else {
        newSet.add(groupId);
      }
      return newSet;
    });
  };

  const handleSubmit = () => {
    if (selectedGroups.size > 0) {
      // Здесь будет отправка приглашений
      console.log('Inviting user', userId, 'to groups', Array.from(selectedGroups));
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Выберите группу</h2>
          <button className={styles.closeButton} onClick={onClose}>
            <Image src="/icons/close.png" alt="Close" width={20} height={20} />
          </button>
        </div>

        <div 
          className={styles.groupsList}
          ref={scrollContainerRef}
        >
          {mockUserGroups.map((group, index) => (
            <div
              key={group.id}
              className={`${styles.groupItem} ${
                selectedGroups.has(group.id) ? styles.selectedGroup : ''
              }`}
              onClick={() => handleGroupToggle(group.id)}
              style={{
                '--item-index': index
              } as React.CSSProperties}
            >
              <div className={styles.groupContent}>
                <div className={styles.groupImageWrapper}>
                  <Image
                    src={group.image}
                    alt={group.name}
                    width={50}
                    height={50}
                    className={styles.groupImage}
                  />
                </div>
                
                <div className={styles.groupInfo}>
                  <span className={styles.groupName}>{group.name}</span>
                  <span className={styles.memberCount}>{group.count} участников</span>
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className={styles.footer}>
          <button
            className={styles.cancelButton}
            onClick={onClose}
          >
            Отмена
          </button>
          <button
            className={`${styles.submitButton} ${
              selectedGroups.size === 0 ? styles.disabledButton : ''
            }`}
            onClick={handleSubmit}
            disabled={selectedGroups.size === 0}
          >
            Пригласить ({selectedGroups.size})
          </button>
        </div>
      </div>
    </div>
  );
}