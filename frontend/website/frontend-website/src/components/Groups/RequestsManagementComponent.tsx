// src/components/Groups/RequestsManagementComponent.tsx

'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import Image from 'next/image';
import styles from '../../styles/Groups/admin/RequestsManagement.module.css';

// Тип для данных заявки от сервера
interface RequestData {
  id: number;
  userId: number;
  groupId: number;
  status: 'pending' | 'accepted' | 'rejected';
  user: {
    id: number;
    name: string;
    email: string;
    image: string;
  };
  group: {
    id: number;
    name: string;
    image: string;
  };
}

// Тестовые данные
const mockRequests: RequestData[] = [
  {
    id: 1,
    userId: 42,
    groupId: 5,
    status: 'pending',
    user: {
      id: 42,
      name: 'Ассемблер',
      email: 'assembler@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 2,
    userId: 43,
    groupId: 5,
    status: 'pending',
    user: {
      id: 43,
      name: 'Дмитрий Кодов',
      email: 'dmitry.kodov@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 3,
    userId: 44,
    groupId: 5,
    status: 'pending',
    user: {
      id: 44,
      name: 'Алексей Программист',
      email: 'alex.programmer@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 4,
    userId: 45,
    groupId: 5,
    status: 'pending',
    user: {
      id: 45,
      name: 'Мария Дизайнер',
      email: 'maria.designer@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 5,
    userId: 46,
    groupId: 5,
    status: 'pending',
    user: {
      id: 46,
      name: 'Иван Тестировщик',
      email: 'ivan.tester@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 6,
    userId: 47,
    groupId: 5,
    status: 'pending',
    user: {
      id: 47,
      name: 'Елена Менеджер',
      email: 'elena.manager@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 7,
    userId: 48,
    groupId: 5,
    status: 'pending',
    user: {
      id: 48,
      name: 'Павел Архитектор',
      email: 'pavel.architect@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 8,
    userId: 49,
    groupId: 5,
    status: 'pending',
    user: {
      id: 49,
      name: 'Анна Аналитик',
      email: 'anna.analyst@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 9,
    userId: 50,
    groupId: 5,
    status: 'pending',
    user: {
      id: 50,
      name: 'Сергей Фронтенд',
      email: 'sergey.frontend@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  },
  {
    id: 10,
    userId: 51,
    groupId: 5,
    status: 'pending',
    user: {
      id: 51,
      name: 'Ольга Бэкенд',
      email: 'olga.backend@example.com',
      image: '/default-avatar.png'
    },
    group: {
      id: 5,
      name: 'Закрытая группа',
      image: 'https://img.com/group.png'
    }
  }
];

interface RequestsManagementComponentProps {
  groupId?: string;
  // requests?: RequestData[]; // Будет использоваться когда подключим API
}

const RequestsManagementComponent: React.FC<RequestsManagementComponentProps> = ({ groupId }) => {
  const [requests, setRequests] = useState<RequestData[]>(mockRequests);
  const [searchTerm, setSearchTerm] = useState('');
  const [canScrollUp, setCanScrollUp] = useState(false);
  const [canScrollDown, setCanScrollDown] = useState(false);
  const scrollContainerRef = useRef<HTMLDivElement>(null);

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
  const acceptRequest = (requestId: number) => {
    setRequests(prev => prev.filter(req => req.id !== requestId));
    // Здесь будет API вызов для принятия заявки
    console.log('Принята заявка:', requestId);
  };

  // Отклонить заявку
  const rejectRequest = (requestId: number) => {
    setRequests(prev => prev.filter(req => req.id !== requestId));
    // Здесь будет API вызов для отклонения заявки
    console.log('Отклонена заявка:', requestId);
  };

  // Принять все заявки
  const acceptAllRequests = () => {
    setRequests([]);
    // Здесь будет API вызов для принятия всех заявок
    console.log('Приняты все заявки');
  };

  // Отклонить все заявки
  const rejectAllRequests = () => {
    setRequests([]);
    // Здесь будет API вызов для отклонения всех заявок
    console.log('Отклонены все заявки');
  };

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
            disabled={filteredRequests.length === 0}
          >
            Принять все
          </button>
          <button 
            className={styles.rejectAllBtn}
            onClick={rejectAllRequests}
            disabled={filteredRequests.length === 0}
          >
            Отклонить все
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
                  >
                    Принять
                  </button>
                  <button 
                    className={styles.rejectBtn}
                    onClick={() => rejectRequest(request.id)}
                  >
                    Отклонить
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