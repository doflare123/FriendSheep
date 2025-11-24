// src/components/Groups/MembersManagement.tsx

'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import Image from 'next/image';
import styles from '@/styles/Groups/admin/MembersManagement.module.css';
import { getAccesToken } from '@/Constants';
import { kickMember } from '@/api/groups/kickMember';
import LoadingIndicator from '@/components/LoadingIndicator';
import { showNotification } from '@/utils';
import { useRouter } from 'next/navigation';
import { GroupData, Subscriber } from '../../types/Groups';

interface MembersManagementProps {
  groupId?: number | string; // Поддерживаем оба типа
  groupData?: GroupData; 
  onGroupDataUpdate?: (updatedData: Partial<GroupData>) => void;
}

const MembersManagement: React.FC<MembersManagementProps> = ({ 
  groupId, 
  groupData,
  onGroupDataUpdate,
}) => {
  // Тестовые данные для проверки скролла

  const [members, setMembers] = useState<Subscriber[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [canScrollUp, setCanScrollUp] = useState(false);
  const [canScrollDown, setCanScrollDown] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isProcessing, setIsProcessing] = useState(false);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const router = useRouter();

  // Загрузка участников при монтировании компонента
  useEffect(() => {
    loadMembers();
  }, [groupData]);

  const loadMembers = () => {
    setIsLoading(true);
    try {
      if (groupData?.subscribers) {
        // Используем реальные данные
        setMembers(groupData.subscribers);
      } else {
        setMembers([]);
      }
    } catch (error) {
      console.error('Ошибка при загрузке участников:', error);
      showNotification(500, 'Ошибка при загрузке участников');
    } finally {
      setIsLoading(false);
    }
  };

  // Фильтрация участников по поиску (только по имени)
  const filteredMembers = members.filter(member =>
    member.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Проверка возможности скролла
  const checkScrollability = useCallback(() => {
    if (scrollContainerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = scrollContainerRef.current;
      const hasVerticalScroll = scrollHeight > clientHeight;
      
      setCanScrollUp(hasVerticalScroll && scrollTop > 5);
      setCanScrollDown(hasVerticalScroll && scrollTop < scrollHeight - clientHeight - 5);
    }
  }, []);

  // Проверяем скролл при изменении данных
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      checkScrollability();
    }, 100);

    return () => clearTimeout(timeoutId);
  }, [filteredMembers, checkScrollability]);

  // Проверяем скролл при монтировании компонента
  useEffect(() => {
    checkScrollability();
  }, [checkScrollability]);

  // Обработчик изменения размера окна
  useEffect(() => {
    const handleResize = () => {
      checkScrollability();
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [checkScrollability]);

  // Скролл вверх
  const scrollUp = () => {
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollBy({
        top: -200,
        behavior: 'smooth'
      });
    }
  };

  // Скролл вниз
  const scrollDown = () => {
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollBy({
        top: 200,
        behavior: 'smooth'
      });
    }
  };

  // Удалить участника
  const removeMember = async (userId: number) => {
    setIsProcessing(true);
    try {
      const accessToken = getAccesToken(router);
      if (!accessToken) {
        showNotification(401, 'Токен доступа не найден');
        return;
      }

      await kickMember(accessToken, groupId, userId);
      
      // Обновляем локальное состояние
      const updatedMembers = members.filter(member => member.userId !== userId);
      setMembers(updatedMembers);
      
      showNotification(200, 'Участник успешно удален');
      
      // Обновляем данные группы в родительском компоненте
      if (onGroupDataUpdate && groupData) {
        onGroupDataUpdate({
          ...groupData,
          subscribers: updatedMembers,
          count_members: updatedMembers.length
        });
      }
    } catch (error: any) {
      console.error('Ошибка при удалении участника:', error);
      const statusCode = error.response?.status || 500;
      const message = error.response?.data?.message || 'Ошибка при удалении участника';
      showNotification(statusCode, message);
    } finally {
      setIsProcessing(false);
    }
  };

  // Показываем индикатор загрузки
  if (isLoading) {
    return (
      <div className={styles.membersContainer}>
        <LoadingIndicator text="Загрузка участников..." />
      </div>
    );
  }

  return (
    <div className={styles.membersContainer}>
      {/* Поиск */}
      <div className={styles.searchSection}>
        <input
          type="text"
          placeholder="Найти участника..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className={styles.searchInput}
          disabled={isProcessing}
        />
      </div>

      {/* Список участников со скроллом */}
      <div className={styles.membersList}>
        {/* Стрелка вверх */}
        {canScrollUp && (
          <button className={`${styles.scrollButton} ${styles.scrollUp}`} onClick={scrollUp}>
            ↑
          </button>
        )}

        {/* Контейнер со скроллом */}
        <div 
          ref={scrollContainerRef}
          className={styles.scrollContainer}
          onScroll={checkScrollability}
        >
          {filteredMembers.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyText}>
                {searchTerm ? 'Участники не найдены' : 'В группе пока нет участников'}
              </div>
            </div>
          ) : (
            filteredMembers.map((member) => (
              <div key={member.userId} className={styles.memberItem}>
                <div className={styles.userInfo}>
                  <Image
                    src={member.image}
                    alt={member.name}
                    width={40}
                    height={40}
                    className={styles.userAvatar}
                  />
                  <div className={styles.userDetails}>
                    <div className={styles.userName}>{member.name}</div>
                  </div>
                </div>
                
                <div className={styles.memberActions}>
                  <button 
                    className={styles.removeBtn}
                    onClick={() => removeMember(member.userId)}
                    disabled={isProcessing}
                  >
                    {isProcessing ? '...' : 'Удалить'}
                  </button>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Стрелка вниз */}
        {canScrollDown && (
          <button className={`${styles.scrollButton} ${styles.scrollDown}`} onClick={scrollDown}>
            ↓
          </button>
        )}
      </div>
    </div>
  );
};

export default MembersManagement;