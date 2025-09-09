'use client';

import React, { useState } from 'react';
import Image from 'next/image';
import styles from '../../../styles/profile/ProfilePage.module.css';
import {getSocialIcon} from '../../../Constants';

// Тестовые данные на основе API
const testProfileData = {
  data_register: "21.01.2024",
  enterprise: true,
  image: "/default-avatar.png",
  name: "Чебаксар",
  status: "I'm feel so sigma!",
  popular_genres: [
    { count: 30, name: "Боевик" },
    { count: 27, name: "Приколы" },
    { count: 25, name: "Веселье" },
    { count: 23, name: "Русские" },
    { count: 21, name: "Анекдоты" }
  ],
  tiles: ["count_films", "spent_time", "count_games", "count_all"],
  telegram_link: false,
  user_stats: {
    count_all: 20,
    count_another: 240,
    count_create_session: 322,
    count_films: 20,
    count_games: 20,
    count_table_games: 5,
    most_big_session: 8,
    series_session_count: 6,
    spent_time: 20
  },
  recent_sessions: [],
  upcoming_sessions: []
};

const testSubscriptions = [
  {
    id: 1,
    name: "Группа крутых пацанят",
    small_description: "47 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["games"]
  },
  {
    id: 2,
    name: "Группа крутых пацанят",
    small_description: "47 участников", 
    image: "/default/group.jpg",
    type: "group",
    category: ["movies"]
  },
  {
    id: 3,
    name: "Группа крутых пацанят",
    small_description: "47 участников",
    image: "/default/group.jpg", 
    type: "group",
    category: ["board"]
  }
];

interface ProfilePageProps {
  params: {
    id: string;
  };
}

export default function ProfilePage({ params }: ProfilePageProps) {
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedData, setEditedData] = useState({
    name: testProfileData.name,
    status: testProfileData.status,
    image: testProfileData.image,
    tiles: [...testProfileData.tiles]
  });
  const [selectedTileIndex, setSelectedTileIndex] = useState<number | null>(null);

  // Все доступные типы плиток
  const availableTiles = ['count_all', 'count_films', 'count_games', 'count_other', 'count_table', 'spent_time'];

  const getTileIcon = (tileType: string) => {
    return `/profile/${tileType}.png`;
  };

  const getTileLabel = (tileType: string) => {
    const labels: { [key: string]: string } = {
      count_all: "Всего",
      count_films: "Кино",
      count_games: "Игры", 
      count_other: "Другое",
      count_table: "Настолки",
      spent_time: "Часов"
    };
    return labels[tileType] || tileType;
  };

  const getTileValue = (tileType: string) => {
    const stats = testProfileData.user_stats;
    const values: { [key: string]: number } = {
      count_all: stats.count_all,
      count_films: stats.count_films,
      count_games: stats.count_games,
      count_other: stats.count_another,
      count_table: stats.count_table_games,
      spent_time: stats.spent_time
    };
    return values[tileType] || 0;
  };

  const handleSettingsClick = () => {
    if (isEditMode) {
      // Показать кнопки сохранения/отмены
    } else {
      setIsEditMode(true);
    }
  };

  const handleSave = () => {
    // Логика сохранения
    setIsEditMode(false);
    console.log('Сохранение данных:', editedData);
  };

  const handleCancel = () => {
    // Возврат к исходным данным
    setEditedData({
      name: testProfileData.name,
      status: testProfileData.status,
      image: testProfileData.image,
      tiles: [...testProfileData.tiles]
    });
    setIsEditMode(false);
  };

  const handleTileClick = (index: number) => {
    if (isEditMode) {
      setSelectedTileIndex(index);
    }
  };

  const handleTileChange = (newTileType: string) => {
    if (selectedTileIndex !== null) {
      const newTiles = [...editedData.tiles];
      newTiles[selectedTileIndex] = newTileType;
      setEditedData({...editedData, tiles: newTiles});
      setSelectedTileIndex(null);
    }
  };

  const handleTelegramClick = () => {
    if (!testProfileData.telegram_link) {
      window.open('https://t.me/FriendShipNotify_bot', '_blank');
    }
  };

  return (
    <div className="bgPage">
      <div className={styles.profileContainer}>
        {/* Секция 1: Краткая информация */}
        <div className={styles.profileInfo}>
          <div className={styles.settingsButton}>
            {!isEditMode ? (
              <button onClick={handleSettingsClick} className={styles.settingsBtn}>
                <Image src="/setting.png" alt="settings" width={24} height={24} />
              </button>
            ) : (
              <div className={styles.editButtons}>
                <button onClick={handleSave} className={styles.editBtn}>
                  <Image src="/profile/edit_save.png" alt="save" width={20} height={20} />
                </button>
                <button onClick={handleCancel} className={styles.editBtn}>
                  <Image src="/profile/edit_cancel.png" alt="cancel" width={20} height={20} />
                </button>
              </div>
            )}
          </div>

          <div className={styles.userHeader}>
            <div className={styles.avatarContainer}>
              <Image 
                src={editedData.image} 
                alt="avatar" 
                width={120} 
                height={120}
                className={styles.avatar}
              />
              {isEditMode && (
                <div className={styles.editIcon}>
                  <Image src="/profile/edit.png" alt="edit" width={16} height={16} />
                </div>
              )}
            </div>

            <div className={styles.userInfo}>
              <div className={styles.userNameContainer}>
                <div className={styles.nameWrapper}>
                  {isEditMode ? (
                    <input 
                      value={editedData.name}
                      onChange={(e) => setEditedData({...editedData, name: e.target.value})}
                      className={styles.nameInput}
                    />
                  ) : (
                    <h2 className={styles.userName}>{editedData.name}</h2>
                  )}
                  {testProfileData.enterprise && !isEditMode && (
                    <Image src="/profile/mark.png" alt="verified" width={20} height={20} className={styles.verifiedMark} />
                  )}
                  {isEditMode && (
                    <Image src="/profile/edit.png" alt="edit" width={16} height={16} className={styles.editIcon} />
                  )}
                </div>
                
                {!isEditMode ? (
                  <div className={styles.telegramContainer}>
                    <div 
                      className={styles.telegramIcon}
                      onClick={handleTelegramClick}
                      title={!testProfileData.telegram_link ? "У вас не привязан телеграм, чтобы получать напоминания о событиях, на которые вы записались. Но это можно исправить зайдя в этого бота в телеграме: @FriendShipNotify_bot" : ""}
                    >
                      <Image 
                        src={getSocialIcon("tg")} 
                        alt="telegram" 
                        width={24} 
                        height={24}
                        className={!testProfileData.telegram_link ? styles.telegramDisabled : ''}
                      />
                    </div>
                  </div>
                ) : null}
              </div>
              
              <div className={styles.userContainer}>
                <div className={styles.userText}>@user</div>
              </div>

              <div className={styles.userStatusContainer}>
                {isEditMode ? (
                  <textarea 
                    value={editedData.status}
                    onChange={(e) => setEditedData({...editedData, status: e.target.value})}
                    className={styles.statusInput}
                    rows={2}
                    placeholder="Введите статус"
                  />
                ) : (
                  <p className={styles.userStatus}>{editedData.status}</p>
                )}
                {isEditMode && (
                  <Image src="/profile/edit.png" alt="edit" width={16} height={16} className={styles.editIcon} />
                )}
                <div className={styles.registerDate}>
                  Участник с {testProfileData.data_register}
                </div>
              </div>
              
            </div>
          </div>

          {/* Статистические плитки */}
          <div className={styles.statsTiles}>
            {editedData.tiles.map((tileType, index) => (
              <div key={index} className={styles.statTile} onClick={() => handleTileClick(index)}>
                <div className={styles.tileIcon}>
                  <Image 
                    src={getTileIcon(tileType)} 
                    alt={tileType} 
                    width={24} 
                    height={24}
                  />
                </div>
                <div className={styles.tileValue}>{getTileValue(tileType)}</div>
                <div className={styles.tileLabel}>{getTileLabel(tileType)}</div>
                {isEditMode && (
                  <div className={styles.editIcon}>
                    <Image src="/profile/edit.png" alt="edit" width={12} height={12} />
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Модальное окно выбора плитки */}
          {selectedTileIndex !== null && (
            <div className={styles.tileModal}>
              <div className={styles.tileModalContent}>
                <h4>Выберите тип плитки:</h4>
                <div className={styles.tileOptions}>
                  {availableTiles.map((tileType) => (
                    <div 
                      key={tileType} 
                      className={styles.tileOption}
                      onClick={() => handleTileChange(tileType)}
                    >
                      <Image src={getTileIcon(tileType)} alt={tileType} width={20} height={20} />
                      <span>{getTileLabel(tileType)}</span>
                    </div>
                  ))}
                </div>
                <button onClick={() => setSelectedTileIndex(null)} className={styles.closeModal}>
                  Отмена
                </button>
              </div>
            </div>
          )}

          {/* Топ жанров */}
          <div className={styles.topGenres}>
            <h4>Топ любимых жанров:</h4>
            <div className={styles.genresList}>
              {testProfileData.popular_genres.map((genre, index) => (
                <div key={index} className={styles.genreItem}>
                  <span className={styles.genreNumber}>{index + 1}.</span>
                  <span className={styles.genreName}>{genre.name}</span>
                  <span className={styles.genreCount}>{genre.count}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Секция 2: Статистика (заглушка) */}
        <div className={styles.statisticsSection}>
          <h3>Ваша статистика</h3>
          <div className={styles.placeholder}>
            Статистическая диаграмма и дополнительные плитки будут здесь
          </div>
        </div>

        {/* Секция 3: Подписки (заглушка) */}
        <div className={styles.subscriptionsSection}>
          <h3>Ваши подписки</h3>
          <div className={styles.placeholder}>
            Скролл подписок будет здесь
          </div>
        </div>

        {/* Секция 4: События (заглушка) */}
        <div className={styles.eventsSection}>
          <h3>События</h3>
          <div className={styles.placeholder}>
            Скролл событий будет здесь
          </div>
        </div>
      </div>
    </div>
  );
}