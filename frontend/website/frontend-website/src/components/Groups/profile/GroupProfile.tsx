'use client';

import React from 'react';
import Image from 'next/image';
import styles from '../../../styles/Groups/profile/GroupProfile.module.css';
import CategorySection from '../../Events/CategorySection';
import { SectionData, EventCardProps } from '../../../types/Events';
import { GroupProfileProps, GroupData, Contact, SessionWithMetadata } from '../../../types/Groups';

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
      duration: formatDuration(sessionItem.session.duration),
      location: sessionItem.session.session_place === 'online' ? 'online' : 'offline'
    }))
  };
};

// Функция для преобразования типа события
const getEventType = (sessionType: string): EventCardProps['type'] => {
  const lowerType = sessionType.toLowerCase();
  
  if (lowerType.includes('game') || lowerType.includes('игр')) {
    return 'games';
  } else if (lowerType.includes('movie') || lowerType.includes('film') || lowerType.includes('кино') || lowerType.includes('фильм')) {
    return 'movies';
  } else if (lowerType.includes('board') || lowerType.includes('настольн')) {
    return 'board';
  } else {
    return 'other';
  }
};

// Функция для форматирования даты
const formatDate = (dateString: string): string => {
  try {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', { 
      day: 'numeric', 
      month: 'short' 
    });
  } catch {
    return dateString;
  }
};

// Функция для форматирования длительности
const formatDuration = (duration: number): string => {
  if (duration < 60) {
    return `${duration} мин`;
  } else {
    const hours = Math.floor(duration / 60);
    const minutes = duration % 60;
    return minutes > 0 ? `${hours}ч ${minutes}м` : `${hours} часа`;
  }
};

// Функция для получения иконки социальной сети по ссылке
const getSocialIcon = (link: string, name: string): string => {
  const lowerLink = link.toLowerCase();
  const lowerName = name.toLowerCase();
  
  if (lowerLink.includes('discord') || lowerName.includes('discord')) {
    return 'social/ds.png';
  } else if (lowerLink.includes('t.me') || lowerLink.includes('telegram') || lowerName.includes('telegram')) {
    return 'social/tg.png';
  } else if (lowerLink.includes('vk.com') || lowerName.includes('вконтакте') || lowerName.includes('vk')) {
    return 'social/vk.png';
  } else if (lowerLink.includes('wa.me') || lowerLink.includes('whatsapp') || lowerName.includes('whatsapp')) {
    return 'social/wa.png';
  } else if (lowerLink.includes('snapchat') || lowerName.includes('snapchat')) {
    return 'social/snap.png';
  } else {
    return 'default/soc_net.png';
  }
};

// Функция для получения иконки категории
const getCategoryIcon = (category: string): string => {
  const lowerCategory = category.toLowerCase();
  
  if (lowerCategory.includes('game') || lowerCategory.includes('игр')) {
    return 'events/games.png';
  } else if (lowerCategory.includes('movie') || lowerCategory.includes('film') || lowerCategory.includes('кино') || lowerCategory.includes('фильм')) {
    return 'events/movies.png';
  } else if (lowerCategory.includes('board') || lowerCategory.includes('настольн')) {
    return 'events/board.png';
  } else {
    return 'events/games.png'; // default
  }
};

const GroupProfile: React.FC<GroupProfileProps> = ({ groupData }) => {
  const handleJoinGroup = () => {
    console.log(`Присоединение к группе с ID: ${groupData.id}`);
    console.log('Данные группы:', groupData);
  };

  const handleContactClick = (contact: Contact) => {
    console.log(`Переход по ссылке: ${contact.link}`);
    window.open(contact.link, '_blank');
  };

  // Преобразуем sessions в формат для CategorySection
  const eventsData = transformSessionsToEvents(groupData.sessions);

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

          {/* Кнопка присоединиться */}
          <button className={styles.joinButton} onClick={handleJoinGroup}>
            Присоединиться
          </button>

          {/* Контакты */}
          <div className={styles.contactsSection}>
            <h3>Наши контакты:</h3>
            <div className={styles.contactsList}>
              {groupData.contacts && groupData.contacts.length > 0 ? (
                groupData.contacts.map((contact, index) => (
                  <div 
                    key={index} 
                    className={styles.contactItem}
                    onClick={() => handleContactClick(contact)}
                  >
                    <div className={styles.contactIconWrapper}>
                      <Image 
                        src={`/${getSocialIcon(contact.link, contact.name)}`}
                        alt={contact.name}
                        width={52}
                        height={52}
                        className={styles.contactIcon}
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
                    {groupData.categories.map((category, index) => (
                      <Image
                        key={index}
                        src={`/${getCategoryIcon(category)}`}
                        alt={category}
                        width={32}
                        height={32}
                        className={styles.categoryIcon}
                      />
                    ))}
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
                    <div key={index} className={styles.subscriberItem}>
                      <div className={styles.subscriberAvatar}>
                        <Image
                          src={user.image || "/default-avatar.png"}
                          alt={user.name}
                          width={64}
                          height={64}
                          className={styles.subscriberAvatarImage}
                        />
                      </div>
                      <span className={styles.subscriberName}>{user.name}</span>
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
              <div className={styles.eventsContent}>
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