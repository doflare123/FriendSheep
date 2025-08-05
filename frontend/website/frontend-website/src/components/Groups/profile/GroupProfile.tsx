'use client';

import React from 'react';
import Image from 'next/image';
import styles from '../../../styles/Groups/profile/GroupProfile.module.css';

interface GroupProfileProps {
  groupId: string;
}

// Тестовые данные (в реальном приложении будут получаться по API на основе groupId)
const mockGroupData = {
  1: {
    id: 1,
    name: 'Мега крутая группа',
    location: 'Калининград, Россия',
    description: 'МАКСИМАЛЬНО КРУТАЯ ГРУППА, ПРОСТО ЖЕСТЬ! Здесь вы сможете поболтать и посмотреть фильмы с интеллегентами!!! 😎😁',
    avatar: '/group-avatar.jpg',
    membersCount: 10000,
    categories: ['events/board.png', 'events/games.png', 'events/movies.png'],
    contacts: [
      { name: 'Группка', link: 'https://discord.gg/example', icon: 'social/ds.png' },
      { name: 'Канальчик', link: 'https://t.me/example', icon: 'social/tg.png' },
      { name: 'Страница автора', link: 'https://vk.com/example', icon: 'social/vk.png' },
      { name: 'WhatsApp группа', link: 'https://wa.me/example', icon: 'social/wa.png' },
      { name: 'Snapchat', link: 'https://snapchat.com/example', icon: 'social/snap.png' },
      { name: 'Custom Social', link: 'https://custom.com/example', icon: 'default/soc_net.png' },
    ]
  }
};

const mockSubscribers = [
  { id: 1, name: 'Чел', avatar: '/user1.jpg' },
  { id: 2, name: 'Flex228666', avatar: '/user2.jpg' },
  { id: 3, name: 'Крутой', avatar: '/user3.jpg' },
  { id: 4, name: 'Чел', avatar: '/user4.jpg' },
  { id: 5, name: 'Flex228666', avatar: '/user5.jpg' },
  { id: 6, name: 'Тигр', avatar: '/user6.jpg' }
];

const GroupProfile: React.FC<GroupProfileProps> = ({ groupId }) => {
  // В реальном приложении здесь будет запрос к API
  const groupData = mockGroupData[groupId as keyof typeof mockGroupData] || mockGroupData[1];

  const handleJoinGroup = () => {
    console.log(`Присоединение к группе с ID: ${groupId}`);
    console.log('Данные группы:', groupData);
  };

  const handleContactClick = (contact: { name: string; link: string }) => {
    console.log(`Переход по ссылке: ${contact.link}`);
    window.open(contact.link, '_blank');
  };

  return (
    <div className='bgPage' style={{ display: 'flex' }}>
      <div className={styles.mainContainer}>
        {/* Левая часть */}
        <div className={styles.leftSection}>
          {/* Аватар */}
          <div className={styles.groupAvatar}>
            <Image
              src="/default/group.jpg"
              alt={groupData.name}
              width={300}
              height={300}
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
              {groupData.contacts.map((contact, index) => (
                <div 
                  key={index} 
                  className={styles.contactItem}
                  onClick={() => handleContactClick(contact)}
                >
                  <div className={styles.contactIconWrapper}>
                    <Image 
                      src={`/${contact.icon}`}
                      alt={contact.name}
                      width={32}
                      height={32}
                      className={styles.contactIcon}
                    />
                  </div>
                  <span className={styles.contactName}>{contact.name}</span>
                </div>
              ))}
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
                <div className={styles.categoryIcons}>
                  {groupData.categories.map((category, index) => (
                    <Image
                      key={index}
                      src={`/${category}`}
                      alt="Category"
                      width={24}
                      height={24}
                      className={styles.categoryIcon}
                    />
                  ))}
                </div>
              </div>
              
              {/* Город */}
              <p className={styles.groupLocation}>{groupData.location}</p>
              
              {/* Описание */}
              <div className={styles.descriptionSection}>
                <p className={styles.descriptionLabel}>Описание группы:</p>
                <p className={styles.descriptionText}>{groupData.description}</p>
              </div>
            </div>

            {/* Подписчики */}
            <div className={styles.subscribersSection}>
              <h3>Подписчики: {groupData.membersCount.toLocaleString()}</h3>
              <div className={styles.subscribersList}>
                {mockSubscribers.map(subscriber => (
                  <div key={subscriber.id} className={styles.subscriberItem}>
                    <div className={styles.subscriberAvatar}>
                      <Image
                        src="/default-avatar.png"
                        alt={subscriber.name}
                        width={70}
                        height={70}
                        className={styles.subscriberAvatarImage}
                      />
                    </div>
                    <span className={styles.subscriberName}>{subscriber.name}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Нижняя часть - события */}
          <div className={styles.eventsSection}>
            <h3>Запланированные мероприятия:</h3>
            <div className={styles.eventsContent}>
              <p className={styles.emptyMessage}>Здесь пока пусто</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default GroupProfile;