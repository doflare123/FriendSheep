'use client';

import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import styles from '../../../styles/Groups/profile/GroupProfile.module.css';
import CategorySection from '../../Events/CategorySection';
import SocialIcon from '@/components/SocialIcon';
import VerifiedBadge from '@/components/VerifiedBadge';
import { SectionData, EventCardProps } from '../../../types/Events';
import { GroupProfileProps, GroupData, Contact, SessionWithMetadata } from '../../../types/Groups';
import {getCategoryIcon, getAccesToken, formatDate, convertCategRuToEng, getUserData} from '../../../Constants'
import {joinGroup} from '@/api/groups/joinGroup';
import {leaveGroup} from '@/api/groups/leaveGroup';
import {getOwnGroups} from '@/api/get_owngroups';
import {getOtherUserInfo} from '@/api/profile/getProfile';
import {getUserInfo} from '@/api/profile/getOwnProfile';
import LoadingIndicator from '@/components/LoadingIndicator';
import { showNotification } from '@/utils';
import { useRouter } from 'next/navigation';

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

const getEventType = (categorySession?: string): 'games' | 'movies' | 'board' | 'other' => {
  if (!categorySession) return 'other';
  const types = convertCategRuToEng([categorySession]);
  return types.length > 0 ? types[0] : 'other';
};

interface CreatorInfo {
  name: string;
  us: string;
  image: string;
  enterprise: boolean;
}

const GroupProfile: React.FC<GroupProfileProps> = ({ groupData }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [isSubscribed, setIsSubscribed] = useState(groupData.subscription);
  const [isOwner, setIsOwner] = useState(false);
  const [creatorInfo, setCreatorInfo] = useState<CreatorInfo | null>(null);
  const router = useRouter();

  useEffect(() => {
    const checkOwnership = async () => {
      try {
        const accessToken = await getAccesToken(router);
        
        if (!accessToken) {
          setIsOwner(false);
          return;
        }

        const ownGroups = await getOwnGroups(accessToken);
        const isAdmin = ownGroups.some((group: any) => group.id === groupData.id);
        setIsOwner(isAdmin);
      } catch (error) {
        console.error('Ошибка при проверке владения группой:', error);
        setIsOwner(false);
      }
    };
    
    checkOwnership();
  }, [groupData.id, router]);

  useEffect(() => {
    const loadCreatorInfo = async () => {
      if (!groupData.creater) return;

      try {
        const accessToken = await getAccesToken(router);
        if (!accessToken) return;

        const currentUser = getUserData();
        let creatorData;

        if (currentUser && currentUser.id === groupData.creater) {
          creatorData = await getUserInfo(accessToken);
        } else {
          creatorData = await getOtherUserInfo(accessToken, groupData.creater);
        }

        setCreatorInfo({
          name: creatorData.name,
          us: creatorData.us,
          image: creatorData.image,
          enterprise: creatorData.enterprise || false
        });
      } catch (error) {
        console.error('Ошибка при загрузке информации о создателе:', error);
      }
    };

    loadCreatorInfo();
  }, [groupData.creater, router]);

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

  const handleCreatorClick = () => {
    if (creatorInfo) {
      router.push(`/profile/${creatorInfo.us}`);
    }
  };

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
        <div className={styles.leftSection}>
          <div className={styles.groupAvatar}>
            <Image
              src={groupData.image || "/default/group.jpg"}
              alt={groupData.name}
              width={200}
              height={200}
              className={styles.groupAvatarImage}
            />
          </div>

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

          {creatorInfo && (
            <div className={styles.creatorSection} onClick={handleCreatorClick}>
              <div className={styles.creatorAvatar}>
                <Image
                  src={creatorInfo.image}
                  alt={creatorInfo.name}
                  width={40}
                  height={40}
                  className={styles.creatorAvatarImage}
                />
              </div>
              <div className={styles.creatorInfo}>
                <span className={styles.creatorLabel}>Создатель:</span>
                <div className={styles.creatorNameWrapper}>
                  <span className={styles.creatorName} title={creatorInfo.name}>
                    {creatorInfo.name}
                  </span>
                  <VerifiedBadge isVerified={creatorInfo.enterprise} size={16} />
                </div>
              </div>
            </div>
          )}

          <div className={styles.contactsSection}>
            <h3>Наши контакты:</h3>
            <div className={styles.contactsList}>
              {groupData.contacts && groupData.contacts.length > 0 ? (
                groupData.contacts.map((contact, index) => (
                  <div 
                    key={index} 
                    className={styles.contactItem}
                  >
                    <div className={styles.contactIconWrapper}>
                      <SocialIcon
                        href={contact.link}
                        alt={contact.name}
                        size={52}
                      />
                    </div>
                    <span className={styles.contactName}>{contact.name}</span>
                  </div>
                ))
              ) : (
                <p className={styles.noContactsText}>Контакты не указаны</p>
              )}
            </div>
          </div>
        </div>

        <div className={styles.rightSection}>
          <div className={styles.topSection}>
            <div className={styles.headerSection}>
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
              
              {groupData.city && (
                <p className={styles.groupLocation}>{groupData.city}</p>
              )}
              
              {groupData.description && (
                <div className={styles.descriptionSection}>
                  <p className={styles.descriptionLabel}>Описание группы:</p>
                  <p className={styles.descriptionText}>{groupData.description}</p>
                </div>
              )}
            </div>

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