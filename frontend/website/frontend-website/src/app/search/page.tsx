'use client';

export const dynamic = 'force-dynamic';

import { useState, useEffect, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Image from 'next/image';
import styles from '../../styles/search/search.module.css';
import { getCategoryIcon, getAccesToken, convertCategRuToEng } from '../../Constants';
import AddUserModal from '../../components/search/AddUserModal';
import { searchGroup } from '@/api/search/searchGroup';
import { searchUsers } from '@/api/search/searchUsers';
import { joinGroup } from '@/api/groups/joinGroup';
import { showNotification } from '@/utils';
import LoadingIndicator from '@/components/LoadingIndicator';

interface User {
  id: number;
  image: string;
  name: string;
  status: string;
  us: string;
}

interface Group {
  category: string[] | null;
  count: number;
  description: string;
  id: number;
  image: string;
  isPrivate?: boolean;
  createdAt?: string;
  name: string;
}

interface UsersResponse {
  has_more: boolean;
  page: number;
  total: number;
  users: User[];
}

interface GroupsResponse {
  groups: Group[];
  has_more: boolean;
  page: number;
  total: number;
}

const ITEMS_PER_PAGE = 5;

export default function SearchPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialType = searchParams.get('type') as 'groups' | 'users' | null;
  
  const [searchType, setSearchType] = useState<'groups' | 'users'>(initialType === 'users' ? 'users' : 'groups');
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [joinedGroups, setJoinedGroups] = useState<Set<number>>(new Set());
  const [requestedGroups, setRequestedGroups] = useState<Set<number>>(new Set());
  const [loadingGroups, setLoadingGroups] = useState<Set<number>>(new Set());
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
  const [isFilterOpen, setIsFilterOpen] = useState(false);
  const [selectedParticipantSort, setSelectedParticipantSort] = useState('По возрастанию');
  const [selectedCategory, setSelectedCategory] = useState('Все');
  const [selectedRegistrationSort, setSelectedRegistrationSort] = useState('По возрастанию');
  const [isLoading, setIsLoading] = useState(false);
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [groups, setGroups] = useState<Group[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [hasAccess, setHasAccess] = useState(false);
  
  const filterRef = useRef<HTMLDivElement>(null);
  const searchTimeoutRef = useRef<NodeJS.Timeout>();
  const isInitialMount = useRef(true);

  // Проверка токена при монтировании
  useEffect(() => {
  const checkAuth = async () => {
    const accessToken = await getAccesToken(router);
    if (!accessToken) {
      console.log("LOGIN2");
      router.push('/login');
    } else {
      setHasAccess(true);
    }
  };
  
  checkAuth();
}, [router]);

  // Обновление searchType при изменении URL параметра
  useEffect(() => {
    const typeFromUrl = searchParams.get('type') as 'groups' | 'users' | null;
    if (typeFromUrl && (typeFromUrl === 'groups' || typeFromUrl === 'users')) {
      setSearchType(typeFromUrl);
    }
  }, [searchParams]);

  // Debounce для поискового запроса
  useEffect(() => {
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }

    searchTimeoutRef.current = setTimeout(() => {
      setDebouncedQuery(searchQuery);
    }, 500);

    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, [searchQuery]);

  // Сброс страницы при изменении поискового запроса
  useEffect(() => {
    if (isInitialMount.current) {
      return;
    }
    setCurrentPage(1);
  }, [debouncedQuery]);

  // Основной эффект загрузки данных
  useEffect(() => {
    if (!hasAccess) return;

    const loadData = async () => {
      setIsLoading(true);
      
      try {
        const accessToken = await getAccesToken(router);
        
        if (searchType === 'groups') {
          const params: any = {};

          // Поиск по названию
          if (debouncedQuery.trim()) {
            params.name = debouncedQuery.trim();
          }

          // Фильтр по категории
          if (selectedCategory !== 'Все') {
            params.category = selectedCategory;
          }

          // Сортировка по участникам или регистрации
          if (selectedParticipantSort !== 'По возрастанию') {
            params.sort_by = 'members';
            params.order = selectedParticipantSort === 'По возрастанию' ? 'asc' : 'desc';
          } else if (selectedRegistrationSort !== 'По возрастанию') {
            params.sort_by = 'date';
            params.order = selectedRegistrationSort === 'По возрастанию' ? 'asc' : 'desc';
          }

          console.log('Загрузка групп, страница:', currentPage, 'params:', params);
          const response: GroupsResponse = await searchGroup(accessToken, currentPage, params);
          console.log('Ответ групп:', response);

          if (response && response.groups) {
            setGroups(response.groups);
            const totalPagesCalc = Math.ceil(response.total / ITEMS_PER_PAGE);
            setTotalPages(totalPagesCalc || 1);
          } else {
            setGroups([]);
            setTotalPages(1);
          }
        } else {
          // Загрузка пользователей
          const params: any = {};

          if (debouncedQuery.trim()) {
            params.name = debouncedQuery.trim();
          }

          console.log('Загрузка пользователей, страница:', currentPage, 'params:', params);
          const response: UsersResponse = await searchUsers(accessToken, currentPage, params);
          console.log('Ответ пользователей:', response);

          if (response && response.users) {
            setUsers(response.users);
            const totalPagesCalc = Math.ceil(response.total / ITEMS_PER_PAGE);
            setTotalPages(totalPagesCalc || 1);
          } else {
            setUsers([]);
            setTotalPages(1);
          }
        }
      } catch (error: any) {
        console.error('Ошибка при загрузке данных:', error);
        showNotification(
          error.response?.status || 500,
          error.response?.data?.message || `Не удалось загрузить ${searchType === 'groups' ? 'группы' : 'пользователей'}`
        );
        
        if (searchType === 'groups') {
          setGroups([]);
        } else {
          setUsers([]);
        }
        setTotalPages(1);
      } finally {
        setIsLoading(false);
        setIsInitialLoading(false);
      }
    };

    loadData();
  }, [hasAccess, searchType, currentPage, debouncedQuery, selectedCategory, selectedParticipantSort, selectedRegistrationSort]);

  // Сброс страницы и данных при изменении фильтров или типа поиска
  useEffect(() => {
    if (isInitialMount.current) {
      isInitialMount.current = false;
      return;
    }
    setCurrentPage(1);
    setGroups([]);
    setUsers([]);
  }, [searchType, selectedCategory, selectedParticipantSort, selectedRegistrationSort]);

  // Закрытие фильтра при клике вне его
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (filterRef.current && !filterRef.current.contains(event.target as Node)) {
        setIsFilterOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const handleJoinGroup = async (groupId: number, isPrivate: boolean, e: React.MouseEvent) => {
    e.stopPropagation(); // Предотвращаем переход на профиль при клике на кнопку
    setLoadingGroups(prev => new Set([...prev, groupId]));

    try {
      const accessToken = await getAccesToken(router);
      await joinGroup(accessToken, groupId);

      if (isPrivate) {
        setRequestedGroups(prev => new Set([...prev, groupId]));
        showNotification(200, 'Заявка на вступление отправлена');
      } else {
        setJoinedGroups(prev => new Set([...prev, groupId]));
        showNotification(200, 'Вы успешно присоединились к группе');
      }
    } catch (error: any) {
      console.error('Ошибка при присоединении к группе:', error);
      
      if (error.response?.status === 400) {
        if (isPrivate) {
          setRequestedGroups(prev => new Set([...prev, groupId]));
          showNotification(400, 'Вы уже отправили заявку в эту группу');
        } else {
          setJoinedGroups(prev => new Set([...prev, groupId]));
          showNotification(400, 'Вы уже присоединены к этой группе');
        }
      } else {
        showNotification(
          error.response?.status || 500,
          error.response?.data?.message || 'Не удалось присоединиться к группе'
        );
      }
    } finally {
      setLoadingGroups(prev => {
        const newSet = new Set(prev);
        newSet.delete(groupId);
        return newSet;
      });
    }
  };

  const handleAddUser = (userId: number, e: React.MouseEvent) => {
    e.stopPropagation(); // Предотвращаем переход на профиль при клике на кнопку
    setSelectedUserId(userId);
    setIsModalOpen(true);
  };

  const handleGroupClick = (groupId: number) => {
    router.push(`/groups/profile/${groupId}`);
  };

  const handleUserClick = (userUs: string) => {
    router.push(`/profile/${userUs}`);
  };

  const handleToggleSearchType = () => {
    const newType = searchType === 'groups' ? 'users' : 'groups';
    setSearchType(newType);
    router.push(`/search?type=${newType}`);
  };

  const renderPagination = () => {
    if (totalPages <= 1) return null;

    return (
      <div className={styles.pagination}>
        <button
          className={styles.arrowButton}
          onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
          disabled={currentPage === 1}
        >
          ←
        </button>
        
        {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
          <button
            key={page}
            className={`${styles.pageNumber} ${currentPage === page ? styles.activePage : ''}`}
            onClick={() => setCurrentPage(page)}
          >
            {page}
          </button>
        ))}
        
        <button
          className={styles.arrowButton}
          onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
          disabled={currentPage === totalPages}
        >
          →
        </button>
      </div>
    );
  };

  return (
    <div className="bgPage">
      {!hasAccess ? (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
          <LoadingIndicator text="Перенаправление..." />
        </div>
      ) : (
        <div className={styles.container}>
          <div className={styles.contentWrapper}>
            <div className={styles.searchHeader}>
              <div className={styles.searchInputWrapper}>
                <input
                  type="text"
                  placeholder={searchType === 'groups' ? 'Найти группу...' : 'Найти пользователя...'}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className={styles.searchInput}
                />
                {searchType === 'groups' && (
                  <div className={styles.filterWrapper} ref={filterRef}>
                    <button
                      className={styles.filterButton}
                      onClick={() => setIsFilterOpen(!isFilterOpen)}
                    >
                      <Image 
                        src="/sorting.png" 
                        alt="Filter" 
                        width={20} 
                        height={20} 
                      />
                    </button>
                    {isFilterOpen && (
                      <div className={styles.filterMenu}>
                        <div className={styles.filterSection}>
                          <h4 className={styles.filterTitle}>Сортировка по категориям</h4>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="categories"
                              checked={selectedCategory === 'Все'}
                              onChange={() => setSelectedCategory('Все')}
                            />
                            <span>Все</span>
                          </label>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="categories"
                              checked={selectedCategory === 'Фильмы'}
                              onChange={() => setSelectedCategory('Фильмы')}
                            />
                            <span>Кино</span>
                          </label>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="categories"
                              checked={selectedCategory === 'Игры'}
                              onChange={() => setSelectedCategory('Игры')}
                            />
                            <span>Игры</span>
                          </label>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="categories"
                              checked={selectedCategory === 'Настольные игры'}
                              onChange={() => setSelectedCategory('Настольные игры')}
                            />
                            <span>Настольные игры</span>
                          </label>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="categories"
                              checked={selectedCategory === 'Другое'}
                              onChange={() => setSelectedCategory('Другое')}
                            />
                            <span>Другое</span>
                          </label>
                        </div>
                        
                        <div className={styles.filterSection}>
                          <h4 className={styles.filterTitle}>Сортировка по участникам</h4>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="participants"
                              checked={selectedParticipantSort === 'По возрастанию'}
                              onChange={() => setSelectedParticipantSort('По возрастанию')}
                            />
                            <span>По возрастанию</span>
                          </label>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="participants"
                              checked={selectedParticipantSort === 'По убыванию'}
                              onChange={() => setSelectedParticipantSort('По убыванию')}
                            />
                            <span>По убыванию</span>
                          </label>
                        </div>
                        
                        <div className={styles.filterSection}>
                          <h4 className={styles.filterTitle}>Сортировка по регистрации</h4>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="registration"
                              checked={selectedRegistrationSort === 'По возрастанию'}
                              onChange={() => setSelectedRegistrationSort('По возрастанию')}
                            />
                            <span>По возрастанию</span>
                          </label>
                          <label className={styles.filterOption}>
                            <input
                              type="radio"
                              name="registration"
                              checked={selectedRegistrationSort === 'По убыванию'}
                              onChange={() => setSelectedRegistrationSort('По убыванию')}
                            />
                            <span>По убыванию</span>
                          </label>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
              
              <button
                className={styles.toggleButton}
                onClick={handleToggleSearchType}
              >
                <div className={`${styles.toggleSide} ${searchType === 'groups' ? styles.active : styles.inactive}`}>
                  <Image 
                    src='/icons/groups.png'
                    alt='Groups' 
                    width={20} 
                    height={20} 
                    title="Группы"
                  />
                </div>
                <div className={styles.toggleDivider} />
                <div className={`${styles.toggleSide} ${searchType === 'users' ? styles.active : styles.inactive}`}>
                  <Image 
                    src='/icons/persons.png'
                    alt='Users' 
                    width={30} 
                    height={30} 
                    title="Люди"
                  />
                </div>
              </button>
            </div>

            {isInitialLoading ? (
              <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
                <LoadingIndicator text={searchType === 'groups' ? 'Загрузка групп...' : 'Загрузка пользователей...'} />
              </div>
            ) : (
              <>
                <div className={styles.resultsContainer} style={{ opacity: isLoading ? 0.5 : 1, transition: 'opacity 0.2s' }}>
                  {searchType === 'groups' ? (
                    <div className={styles.groupsList}>
                      {groups.length > 0 ? (
                        groups.map((group) => (
                          <div key={`group-${group.id}`} className={styles.groupItem}>
                            <div className={styles.groupContent}>
                              <div 
                                className={styles.groupImageWrapper}
                                onClick={() => handleGroupClick(group.id)}
                                style={{ cursor: 'pointer' }}
                              >
                                <Image
                                  src={group.image || '/default/group.jpg'}
                                  alt={group.name}
                                  width={60}
                                  height={60}
                                  className={styles.groupImage}
                                />
                              </div>
                              
                              <div className={styles.groupInfo}>
                                <div className={styles.groupHeader}>
                                  <h3 
                                    className={styles.groupName}
                                    onClick={() => handleGroupClick(group.id)}
                                    style={{ cursor: 'pointer' }}
                                  >
                                    {group.name}
                                    {group.category && group.category.length > 0 && (
                                      <span className={styles.groupIcons}>
                                        {group.category.map((cat, index) => {
                                          const engCategories = convertCategRuToEng([cat]);
                                          const engCat = engCategories[0] || 'other';
                                          return (
                                            <Image
                                              key={index}
                                              src={getCategoryIcon(engCat)}
                                              alt={cat}
                                              width={16}
                                              height={16}
                                              className={styles.categoryIcon}
                                            />
                                          );
                                        })}
                                      </span>
                                    )}
                                  </h3>
                                </div>
                                
                                <p className={styles.groupDescription}>{group.description}</p>
                              </div>
                            </div>
                            
                            <div className={styles.groupActions}>
                              <button
                                className={`${styles.joinButton} ${
                                  joinedGroups.has(group.id) ? styles.joinedButton :
                                  requestedGroups.has(group.id) ? styles.requestedButton : ''
                                }`}
                                onClick={(e) => handleJoinGroup(group.id, group.isPrivate || false, e)}
                                disabled={
                                  joinedGroups.has(group.id) || 
                                  requestedGroups.has(group.id) || 
                                  loadingGroups.has(group.id)
                                }
                              >
                                {loadingGroups.has(group.id) ? 'Загрузка...' :
                                 joinedGroups.has(group.id) ? 'Присоединен' :
                                 requestedGroups.has(group.id) ? 'Заявка отправлена' : 
                                 group.isPrivate ? 'Подать заявку' : 'Присоединиться'}
                              </button>
                              <span className={styles.memberCount}>{group.count} участников</span>
                            </div>
                          </div>
                        ))
                      ) : (
                        <div style={{ 
                          textAlign: 'center', 
                          padding: '60px 20px', 
                          color: 'var(--color-text-primary)',
                          fontSize: '18px'
                        }}>
                          Групп не найдено
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className={styles.usersList}>
                      {users.length > 0 ? (
                        users.map((user) => (
                          <div key={user.id} className={styles.userItem}>
                            <div className={styles.userContent}>
                              <div 
                                className={styles.userImageWrapper}
                                onClick={() => handleUserClick(user.us)}
                                style={{ cursor: 'pointer' }}
                              >
                                <Image
                                  src={user.image || '/default-avatar.png'}
                                  alt={user.name}
                                  width={60}
                                  height={60}
                                  className={styles.userImage}
                                />
                              </div>
                              
                              <div className={styles.userInfo}>
                                <h3 
                                  className={styles.userName}
                                  onClick={() => handleUserClick(user.us)}
                                  style={{ cursor: 'pointer' }}
                                >
                                  {user.name}
                                </h3>
                                <p className={styles.userStatus}>{user.status}</p>
                                <p className={styles.userDescription}>{user.us}</p>
                              </div>
                            </div>
                            
                            <button
                              className={styles.addButton}
                              onClick={(e) => handleAddUser(user.id, e)}
                            >
                              Добавить
                            </button>
                          </div>
                        ))
                      ) : (
                        <div style={{ 
                          textAlign: 'center', 
                          padding: '60px 20px', 
                          color: 'var(--color-text-primary)',
                          fontSize: '18px'
                        }}>
                          Пользователей не найдено
                        </div>
                      )}
                    </div>
                  )}
                </div>

                {!isLoading && renderPagination()}
              </>
            )}
          </div>
        </div>
      )}

      <AddUserModal 
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        userId={selectedUserId}
      />
    </div>
  );
}