'use client';

import { useState, useEffect, useRef } from 'react';
import Image from 'next/image';
import styles from '../../styles/search/AddUserModal.module.css';
import { getAccesToken } from '@/Constants';
import { inviteGroup } from '@/api/groups/inviteGroup';
import { getOwnGroups } from '@/api/get_owngroups';
import { showNotification } from '@/utils';
import LoadingIndicator from '@/components/LoadingIndicator';
import { useRouter } from 'next/navigation';

interface OwnGroup {
  id: number;
  name: string;
  image: string;
  category: string[];
  member_count: number;
  small_description: string;
  type: string;
}

interface AddUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  userId: number | null;
}

export default function AddUserModal({ isOpen, onClose, userId }: AddUserModalProps) {
  const [selectedGroups, setSelectedGroups] = useState<Set<number>>(new Set());
  const [userGroups, setUserGroups] = useState<OwnGroup[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const router = useRouter();

  // Загрузка групп пользователя при открытии модалки
  useEffect(() => {
    if (isOpen) {
      loadUserGroups();
      setSelectedGroups(new Set());
    }
  }, [isOpen]);

  const loadUserGroups = async () => {
    setIsLoading(true);
    try {
      const accessToken = await getAccesToken(router);
      const groups = await getOwnGroups(accessToken);
      
      if (Array.isArray(groups)) {
        setUserGroups(groups);
      } else {
        setUserGroups([]);
        showNotification(500, 'Неверный формат данных групп');
      }
    } catch (error: any) {
      console.error('Ошибка при загрузке групп:', error);
      showNotification(
        error.response?.status || 500,
        error.response?.data?.message || 'Не удалось загрузить список групп'
      );
      setUserGroups([]);
    } finally {
      setIsLoading(false);
    }
  };

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

  const handleSubmit = async () => {
    if (selectedGroups.size === 0 || !userId || isSubmitting) return;

    setIsSubmitting(true);
    try {
      const accessToken = await getAccesToken(router);

      await Promise.all(
        Array.from(selectedGroups).map(groupId =>
          inviteGroup(accessToken, groupId, userId)
        )
      );

      showNotification(
        200, 
        `Пользователь приглашён в ${selectedGroups.size} ${
          selectedGroups.size === 1 ? 'группу' : 
          selectedGroups.size < 5 ? 'группы' : 'групп'
        }`
      );
      onClose();
    } catch (error: any) {
      console.error('Ошибка при приглашении в группы:', error);
      showNotification(
        error.response?.status || 500, 
        error.response?.data?.message || 'Не удалось отправить приглашения'
      );
    } finally {
      setIsSubmitting(false);
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

        {isLoading ? (
          <div style={{ 
            display: 'flex', 
            justifyContent: 'center', 
            alignItems: 'center', 
            minHeight: '300px' 
          }}>
            <LoadingIndicator text="Загрузка групп..." />
          </div>
        ) : userGroups.length === 0 ? (
          <div style={{ 
            textAlign: 'center', 
            padding: '60px 20px', 
            color: '#666',
            fontSize: '16px'
          }}>
            У вас пока нет групп для приглашения
          </div>
        ) : (
          <div 
            className={styles.groupsList}
            ref={scrollContainerRef}
          >
            {userGroups.map((group, index) => (
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
                      src={group.image || '/default/group.jpg'}
                      alt={group.name}
                      width={50}
                      height={50}
                      className={styles.groupImage}
                    />
                  </div>
                  
                  <div className={styles.groupInfo}>
                    <span className={styles.groupName}>{group.name}</span>
                    <span className={styles.memberCount}>
                      {group.member_count} {
                        group.member_count === 1 ? 'участник' :
                        group.member_count < 5 ? 'участника' : 'участников'
                      }
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        <div className={styles.footer}>
          <button
            className={styles.cancelButton}
            onClick={onClose}
            disabled={isSubmitting}
          >
            Отмена
          </button>
          <button
            className={`${styles.submitButton} ${
              selectedGroups.size === 0 || isSubmitting ? styles.disabledButton : ''
            }`}
            onClick={handleSubmit}
            disabled={selectedGroups.size === 0 || isSubmitting}
          >
            {isSubmitting ? 'Отправка...' : `Пригласить (${selectedGroups.size})`}
          </button>
        </div>
      </div>
    </div>
  );
}