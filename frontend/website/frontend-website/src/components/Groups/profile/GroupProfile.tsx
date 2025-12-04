'use client';

import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import styles from '../../../styles/Groups/profile/GroupProfile.module.css';
import CategorySection from '../../Events/CategorySection';
import { SectionData, EventCardProps } from '../../../types/Events';
import { GroupProfileProps, GroupData, Contact, SessionWithMetadata } from '../../../types/Groups';
import {getCategoryIcon, getSocialIcon, getAccesToken, formatDate, convertCategRuToEng, getUserData} from '../../../Constants'
import {joinGroup} from '@/api/groups/joinGroup';
import {leaveGroup} from '@/api/groups/leaveGroup';
import LoadingIndicator from '@/components/LoadingIndicator';
import { showNotification } from '@/utils';
import { useRouter } from 'next/navigation';

// Функция для преобразования sessions в формат для CategorySection
const transformSessionsToEvents = (sessions: SessionWithMetadata[]): SectionData => {
  return {
    title: '',
    pattern: '/patterns/games.png',
    categories: sessions.map(sessionItem => ({
      id: sessionItem.session.id,
      type: getEventType(sessionItem.session.session_type),
      image: sessionItem.session.image_url || '/event_card.jpg',
      date: formatDate(sessionItem.session.start_time),
      title: sessionItem.session.title,
      genres: sessionItem.metadata.genres || [],
      participants: sessionItem.session.current_users,
      maxParticipants: sessionItem.session.count_users_max,
      duration: sessionItem.session.duration,
      location: sessionItem.session.session_place === 'online' ? 'online' : 'offline'
    }))
  };
};

// Функция для преобразования типа события
const getEventType = (categorySession?: string): 'games' | 'movies' | 'board' | 'other' => {
  if (!categorySession) return 'other';
  const types = convertCategRuToEng([categorySession]);
  return types.length > 0 ? types[0] : 'other';
};

const GroupProfile: React.FC<GroupProfileProps> = ({ groupData }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [isSubscribed, setIsSubscribed] = useState(groupData.subscription);
  const [isOwner, setIsOwner] = useState(false);
  const router = useRouter();

  useEffect(() => {
    const checkOwnership = async () => {
      const userData = await getUserData();
      console.log(userData, groupData)
      if (userData && groupData.creater) {
        setIsOwner(userData.us === groupData.creater);
      }
    };
    
    checkOwnership();
  }, [groupData.creater]);

  const handleJoinGroup = async () => {
    const accessToken = await getAccesToken(router);
    
    if (!accessToken) {
      showNotification(401, 'Необходимо авторизоваться', 'error');
      return;
    }

    setIsLoading(true);

    try {
      await joinGroup(accessToken, groupData.id);
      showNotification(200, 'Вы успешно присоединились к группе');
      
      // Обновляем страницу после небольшой задержки для показа уведомления
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Ошибка при присоединении к группе';
      showNotification(error.response?.status || 500, errorMessage);
      setIsLoading(false);
    }
  };

  const handleLeaveGroup = async () => {
    const accessToken = await getAccesToken(router);
    
    if (!accessToken) {
      showNotification(401, 'Необходимо авторизоваться', 'error');
      return;
    }

    setIsLoading(true);

    try {
      await leaveGroup(accessToken, groupData.id);
      showNotification(200, 'Вы успешно отписались от группы');
      
      // Обновляем страницу после небольшой задержки для показа уведомления
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Ошибка при отписке от группы';
      showNotification(error.response?.status || 500, errorMessage);
      setIsLoading(false);
    }
  };

  const handleManageGroup = () => {
    router.push(`/groups/admin/${groupData.id}`);
  };

  const handleContactClick = (contact: Contact) => {
    console.log(`Переход по ссылке: ${contact.link}`);
    window.open(contact.link, '_blank');
  };

  // Преобразуем sessions в формат для CategorySection
  const eventsData = transformSessionsToEvents(groupData.sessions);

  if (isLoading) {
    return (
      <div className='bgPage' style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <LoadingIndicator text={isSubscribed ? "Отписка от группы..." : "Присоединение к группе..."} />
      </div>
    );
  }

  return (
    <div className='bgPage' style={{ display: 'flex' }}>
      <div className={styles.mainContainer}>
        {/* Левая часть */}
        <div className={styles.leftSection}>
          {/* Аватар */}
          <div className={styles.groupAvatar}>
            <Image
              src={groupData.image || "/default/group.jpg"}
              alt={groupData.name}
              width={200}
              height={200}
              className={styles.groupAvatarImage}
            />
          </div>

          {/* Кнопка управления/присоединиться/отписаться */}
          {isOwner ? (
            <button 
              className={styles.joinButton} 
              onClick={handleManageGroup}
              disabled={isLoading}
            >
              Управлять
            </button>
          ) : isSubscribed ? (
            <button 
              className={styles.leaveButton} 
              onClick={handleLeaveGroup}
              disabled={isLoading}
            >
              Отписаться
            </button>
          ) : (
            <button 
              className={styles.joinButton} 
              onClick={handleJoinGroup}
              disabled={isLoading}
            >
              Присоединиться
            </button>
          )}

          {/* Контакты */}
          <div className={styles.contactsSection}>
            <h3>Наши контакты:</h3>
            <div className={styles.contactsList}>
              {groupData.contacts && groupData.contacts.length > 0 ? (
                groupData.contacts.map((contact, index) => {
                  const iconPath = getSocialIcon(contact.link, contact.name);
                  // Убеждаемся, что путь начинается с одного слэша
                  const normalizedPath = iconPath.startsWith('/') ? iconPath : `/${iconPath}`;
                  
                  return (
                    <div 
                      key={index} 
                      className={styles.contactItem}
                      onClick={() => handleContactClick(contact)}
                    >
                      <div className={styles.contactIconWrapper}>
                        <Image 
                          src={normalizedPath}
                          alt={contact.name}
                          width={52}
                          height={52}
                          className={styles.contactIcon}
                        />
                      </div>
                      <span className={styles.contactName}>{contact.name}</span>
                    </div>
                  );
                })
              ) : (
                <p className={styles.noContactsText}>Контакты не указаны</p>
              )}
            </div>
          </div>
        </div>

        {/* Правая часть */}
        <div className={styles.rightSection}>
          {/* Верхняя часть - название, категории, описание */}
          <div className={styles.topSection}>
            <div className={styles.headerSection}>
              {/* Название и категории */}
              <div className={styles.titleAndCategories}>
                <h1 className={styles.groupName}>{groupData.name}</h1>
                {groupData.categories && groupData.categories.length > 0 && (
                  <div className={styles.categoryIcons}>
                    {groupData.categories.map((category, index) => {
                      const categoryIconPath = getCategoryIcon(category);
                      const normalizedCategoryPath = categoryIconPath.startsWith('/') ? categoryIconPath : `/${categoryIconPath}`;
                      
                      return (
                        <Image
                          key={index}
                          src={normalizedCategoryPath}
                          alt={category}
                          width={32}
                          height={32}
                          className={styles.categoryIcon}
                        />
                      );
                    })}
                  </div>
                )}
              </div>
              
              {/* Город */}
              {groupData.city && (
                <p className={styles.groupLocation}>{groupData.city}</p>
              )}
              
              {/* Описание */}
              {groupData.description && (
                <div className={styles.descriptionSection}>
                  <p className={styles.descriptionLabel}>Описание группы:</p>
                  <p className={styles.descriptionText}>{groupData.description}</p>
                </div>
              )}
            </div>

            {/* Подписчики */}
            <div className={styles.subscribersSection}>
              <h3>Подписчики: {groupData.count_members.toLocaleString()}</h3>
              {groupData.users && groupData.users.length > 0 && (
                <div className={styles.subscribersList}>
                  {groupData.users.slice(0, 6).map((user, index) => (
                    <div 
                      key={index} 
                      className={styles.subscriberItem}
                      onClick={() => router.push(`/profile/${user.us}`)}
                      style={{ cursor: 'pointer' }}
                    >
                      <div className={styles.subscriberAvatar}>
                        <Image
                          src={user.image || "/default-avatar.png"}
                          alt={user.name}
                          width={64}
                          height={64}
                          className={styles.subscriberAvatarImage}
                        />
                      </div>
                      <span className={styles.subscriberName} title={user.name}>
                        {user.name}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Нижняя часть - события */}
          {groupData.sessions && groupData.sessions.length > 0 && (
            <div className={styles.eventsSection}>
              <h3>Запланированные мероприятия:</h3>
              <div className={styles.eventsContent}>
                <CategorySection 
                  section={eventsData} 
                  title="" 
                  showCategoryLabel={false} 
                />
              </div>
            </div>
          )}
          
          {(!groupData.sessions || groupData.sessions.length === 0) && (
            <div className={styles.eventsSection}>
              <h3>Запланированные мероприятия:</h3>
              <div className={styles.eventsContentEmpty}>
                <p className={styles.noEventsText}>Пока нет запланированных мероприятий</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default GroupProfile;