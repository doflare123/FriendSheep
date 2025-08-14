// src/components/Groups/RequestsManagementComponent.tsx

'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import Image from 'next/image';
import styles from '../../styles/Groups/admin/RequestsManagement.module.css';
import { RequestData } from '../../types/RequestData';
import { getAccesToken } from '../../Constants';
import { getGroupApplication, approveApplication, rejectApplication } from '../../api/group_requests';

interface RequestsManagementComponentProps {
  groupId?: string;
}

const RequestsManagementComponent: React.FC<RequestsManagementComponentProps> = ({ groupId }) => {
  const [requests, setRequests] = useState<RequestData[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [canScrollUp, setCanScrollUp] = useState(false);
  const [canScrollDown, setCanScrollDown] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isProcessing, setIsProcessing] = useState(false);
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  // Загрузка заявок при монтировании компонента
  useEffect(() => {
    loadRequests();
  }, []);

  const loadRequests = async () => {
    setIsLoading(true);
    try {
      const accessToken = getAccesToken();
      if (!accessToken) {
        console.error('Токен доступа не найден');
        return;
      }

      const allRequestsData = await getGroupApplication(accessToken);
      
      // Фильтруем заявки только для текущей группы
      const filteredByGroup = groupId 
        ? allRequestsData.filter(request => request.groupId === parseInt(groupId))
        : allRequestsData;
      
      setRequests(filteredByGroup);
    } catch (error) {
      console.error('Ошибка при загрузке заявок:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // Фильтрация заявок по поиску
  const filteredRequests = requests.filter(request =>
    request.user.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    request.user.email.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Проверка возможности скролла с useCallback для оптимизации
  const checkScrollability = useCallback(() => {
    if (scrollContainerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = scrollContainerRef.current;
      const hasVerticalScroll = scrollHeight > clientHeight;
      
      setCanScrollUp(hasVerticalScroll && scrollTop > 5); // Небольшой отступ для более точного определения
      setCanScrollDown(hasVerticalScroll && scrollTop < scrollHeight - clientHeight - 5);
    }
  }, []);

  // Проверяем скролл при изменении данных
  useEffect(() => {
    // Используем setTimeout чтобы дождаться завершения рендеринга
    const timeoutId = setTimeout(() => {
      checkScrollability();
    }, 100);

    return () => clearTimeout(timeoutId);
  }, [filteredRequests, checkScrollability]);

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

  // Принять заявку
  const acceptRequest = async (requestId: number) => {
    setIsProcessing(true);
    try {
      const accessToken = getAccesToken();
      if (!accessToken) {
        console.error('Токен доступа не найден');
        return;
      }

      await approveApplication(accessToken, requestId);
      setRequests(prev => prev.filter(req => req.id !== requestId));
      console.log('Принята заявка:', requestId);
    } catch (error) {
      console.error('Ошибка при принятии заявки:', error);
    } finally {
      setIsProcessing(false);
    }
  };

  // Отклонить заявку
  const rejectRequest = async (requestId: number) => {
    setIsProcessing(true);
    try {
      const accessToken = getAccesToken();
      if (!accessToken) {
        console.error('Токен доступа не найден');
        return;
      }

      await rejectApplication(accessToken, requestId);
      setRequests(prev => prev.filter(req => req.id !== requestId));
      console.log('Отклонена заявка:', requestId);
    } catch (error) {
      console.error('Ошибка при отклонении заявки:', error);
    } finally {
      setIsProcessing(false);
    }
  };

  // Принять все заявки
  const acceptAllRequests = async () => {
    if (filteredRequests.length === 0) return;
    
    setIsProcessing(true);
    try {
      const accessToken = getAccesToken();
      if (!accessToken) {
        console.error('Токен доступа не найден');
        return;
      }

      // Принимаем все заявки параллельно
      const promises = filteredRequests.map(request => 
        approveApplication(accessToken, request.id)
      );
      
      await Promise.all(promises);
      
      // Удаляем все принятые заявки из состояния
      const acceptedIds = filteredRequests.map(req => req.id);
      setRequests(prev => prev.filter(req => !acceptedIds.includes(req.id)));
      
      console.log('Приняты все заявки');
    } catch (error) {
      console.error('Ошибка при принятии всех заявок:', error);
    } finally {
      setIsProcessing(false);
    }
  };

  // Отклонить все заявки
  const rejectAllRequests = async () => {
    if (filteredRequests.length === 0) return;
    
    setIsProcessing(true);
    try {
      const accessToken = getAccesToken();
      if (!accessToken) {
        console.error('Токен доступа не найден');
        return;
      }

      // Отклоняем все заявки параллельно
      const promises = filteredRequests.map(request => 
        rejectApplication(accessToken, request.id)
      );
      
      await Promise.all(promises);
      
      // Удаляем все отклоненные заявки из состояния
      const rejectedIds = filteredRequests.map(req => req.id);
      setRequests(prev => prev.filter(req => !rejectedIds.includes(req.id)));
      
      console.log('Отклонены все заявки');
    } catch (error) {
      console.error('Ошибка при отклонении всех заявок:', error);
    } finally {
      setIsProcessing(false);
    }
  };

  // Показываем индикатор загрузки
  if (isLoading) {
    return (
      <div className={styles.requestsContainer} style={{ height: 'auto', maxHeight: 'none' }}>
        <div className={styles.loadingState}>
          <div className={styles.loadingSpinner}>⏳</div>
          <div className={styles.loadingText}>Загружаем заявки...</div>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.requestsContainer} style={{ height: 'auto', maxHeight: 'none' }}>
      {/* Статистика и кнопки массовых действий */}
      <div className={styles.headerSection}>
        <div className={styles.requestsCount}>
          <span className={styles.countNumber}>{filteredRequests.length}</span>
          <span className={styles.countText}>нерассмотренных заявок</span>
        </div>
        
        <div className={styles.massActions}>
          <button 
            className={styles.acceptAllBtn}
            onClick={acceptAllRequests}
            disabled={filteredRequests.length === 0 || isProcessing}
          >
            {isProcessing ? 'Обрабатываем...' : 'Принять все'}
          </button>
          <button 
            className={styles.rejectAllBtn}
            onClick={rejectAllRequests}
            disabled={filteredRequests.length === 0 || isProcessing}
          >
            {isProcessing ? 'Обрабатываем...' : 'Отклонить все'}
          </button>
        </div>
      </div>

      {/* Поиск */}
      <div className={styles.searchSection}>
        <input
          type="text"
          placeholder="Найти пользователя..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className={styles.searchInput}
          disabled={isProcessing}
        />
      </div>

      {/* Список заявок со скроллом */}
      <div className={styles.requestsList} style={{ height: '60vh', maxHeight: '60vh' }}>
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
          {filteredRequests.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>📝</div>
              <div className={styles.emptyText}>
                {searchTerm ? 'Пользователи не найдены' : 'Нет новых заявок'}
              </div>
            </div>
          ) : (
            filteredRequests.map((request) => (
              <div key={request.id} className={styles.requestItem}>
                <div className={styles.userInfo}>
                  <Image
                    src={request.user.image}
                    alt={request.user.name}
                    width={40}
                    height={40}
                    className={styles.userAvatar}
                  />
                  <div className={styles.userDetails}>
                    <div className={styles.userName}>{request.user.name}</div>
                    <div className={styles.userEmail}>@{request.user.email.split('@')[0]}</div>
                  </div>
                </div>
                
                <div className={styles.requestActions}>
                  <button 
                    className={styles.acceptBtn}
                    onClick={() => acceptRequest(request.id)}
                    disabled={isProcessing}
                  >
                    {isProcessing ? '...' : 'Принять'}
                  </button>
                  <button 
                    className={styles.rejectBtn}
                    onClick={() => rejectRequest(request.id)}
                    disabled={isProcessing}
                  >
                    {isProcessing ? '...' : 'Отклонить'}
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

export default RequestsManagementComponent;