'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import styles from '../../styles/search/search.module.css';
import { getCategoryIcon } from '../../Constants';
import AddUserModal from '../../components/search/AddUserModal';

interface User {
  id: number;
  image: string;
  name: string;
  status: string;
  us: string;
}

interface Group {
  category: string[];
  count: number;
  description: string;
  id: number;
  image: string;
  name: string;
  isPrivate?: boolean;
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

const ITEMS_PER_PAGE_USERS = 5;
const ITEMS_PER_PAGE_GROUPS = 4;

// Тестовые данные
const mockUsers: User[] = Array.from({ length: 20 }, (_, i) => ({
  id: i + 1,
  image: '/default/avatar.png',
  name: `Ассемблер${i + 1}`,
  status: '@doflare',
  us: 'Ем жестко пельмени'
}));

const mockGroups: Group[] = Array.from({ length: 20 }, (_, i) => ({
  id: i + 1,
  name: `Группа крутых пацанят ${i + 1}`,
  description: 'Мы крутые пацанчре, въвъв 😎\nПрисоединяйтесь к нам!',
  image: '/default/group.png',
  count: 47,
  category: ['games', 'movies'][i % 2] ? ['games'] : ['movies'],
  isPrivate: i % 3 === 0
}));

export default function SearchPage() {
  const [searchType, setSearchType] = useState<'groups' | 'users'>('groups');
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [joinedGroups, setJoinedGroups] = useState<Set<number>>(new Set());
  const [requestedGroups, setRequestedGroups] = useState<Set<number>>(new Set());
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);

  const itemsPerPage = searchType === 'users' ? ITEMS_PER_PAGE_USERS : ITEMS_PER_PAGE_GROUPS;
  const mockData = searchType === 'users' ? mockUsers : mockGroups;
  
  // Фильтрация по поисковому запросу
  const filteredData = mockData.filter(item => 
    'name' in item ? item.name.toLowerCase().includes(searchQuery.toLowerCase()) :
    item.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const totalPages = Math.ceil(filteredData.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const currentData = filteredData.slice(startIndex, endIndex);

  // Сброс страницы при смене типа поиска
  useEffect(() => {
    setCurrentPage(1);
  }, [searchType]);

  const handleJoinGroup = (groupId: number, isPrivate: boolean) => {
    if (isPrivate) {
      setRequestedGroups(prev => new Set([...prev, groupId]));
    } else {
      setJoinedGroups(prev => new Set([...prev, groupId]));
    }
  };

  const handleAddUser = (userId: number) => {
    setSelectedUserId(userId);
    setIsModalOpen(true);
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
            </div>
            
            <button
              className={styles.toggleButton}
              onClick={() => setSearchType(searchType === 'groups' ? 'users' : 'groups')}
            >
              <Image 
                src={searchType === 'groups' ? '/events/clock.png' : '/events/person.png'} 
                alt={searchType === 'groups' ? 'Groups' : 'Users'} 
                width={20} 
                height={20} 
              />
            </button>
          </div>

        <div className={styles.resultsContainer}>
          {searchType === 'groups' ? (
            <div className={styles.groupsList}>
              {(currentData as Group[]).map((group) => (
                <div key={group.id} className={styles.groupItem}>
                  <div className={styles.groupContent}>
                    <div className={styles.groupImageWrapper}>
                      <Image
                        src={group.image}
                        alt={group.name}
                        width={60}
                        height={60}
                        className={styles.groupImage}
                      />
                    </div>
                    
                    <div className={styles.groupInfo}>
                      <div className={styles.groupHeader}>
                        <h3 className={styles.groupName}>{group.name}</h3>
                        <div className={styles.groupIcons}>
                          {group.category.map((cat, index) => (
                            <Image
                              key={index}
                              src={getCategoryIcon(cat)}
                              alt={cat}
                              width={16}
                              height={16}
                              className={styles.categoryIcon}
                            />
                          ))}
                          {group.isPrivate && (
                            <Image
                              src="/events/lock.png"
                              alt="Private"
                              width={16}
                              height={16}
                              className={styles.categoryIcon}
                            />
                          )}
                        </div>
                      </div>
                      
                      <p className={styles.groupDescription}>{group.description}</p>
                    </div>
                  </div>
                  
                  <div className={styles.groupActions}>
                    <span className={styles.memberCount}>{group.count} участников</span>
                    <button
                      className={`${styles.joinButton} ${
                        joinedGroups.has(group.id) ? styles.joinedButton :
                        requestedGroups.has(group.id) ? styles.requestedButton : ''
                      }`}
                      onClick={() => handleJoinGroup(group.id, group.isPrivate || false)}
                      disabled={joinedGroups.has(group.id) || requestedGroups.has(group.id)}
                    >
                      {joinedGroups.has(group.id) ? 'Присоединен' :
                       requestedGroups.has(group.id) ? 'Заявка отправлена' : 'Присоединиться'}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className={styles.usersList}>
              {(currentData as User[]).map((user) => (
                <div key={user.id} className={styles.userItem}>
                  <div className={styles.userContent}>
                    <div className={styles.userImageWrapper}>
                      <Image
                        src={user.image}
                        alt={user.name}
                        width={60}
                        height={60}
                        className={styles.userImage}
                      />
                    </div>
                    
                    <div className={styles.userInfo}>
                      <h3 className={styles.userName}>{user.name}</h3>
                      <p className={styles.userStatus}>{user.status}</p>
                      <p className={styles.userDescription}>{user.us}</p>
                    </div>
                  </div>
                  
                  <button
                    className={styles.addButton}
                    onClick={() => handleAddUser(user.id)}
                  >
                    Добавить
                  </button>
                </div>
              ))}
            </div>
          )}
                  </div>

          {renderPagination()}
        </div>
      </div>

      <AddUserModal 
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        userId={selectedUserId}
      />
    </div>
  );
}