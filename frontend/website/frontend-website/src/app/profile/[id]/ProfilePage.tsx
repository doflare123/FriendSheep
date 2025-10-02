'use client';

import React, { useState, useEffect, useMemo } from 'react';
import Image from 'next/image';
import { notFound } from 'next/navigation';
import styles from '../../../styles/profile/ProfilePage.module.css';
import section1Styles from '../../../styles/profile/ProfileSection1.module.css';
import section2Styles from '../../../styles/profile/ProfileSection2.module.css';
import section3Styles from '../../../styles/profile/ProfileSection3.module.css';
import section4Styles from '../../../styles/profile/ProfileSection4.module.css';
import StatisticsTile from '../../../components/profile/StatisticsTile';
import SubscriptionItem from '../../../components/profile/SubscriptionItem';
import CategorySection from '../../../components/Events/CategorySection';
import { getSocialIcon } from '../../../Constants';
import GenrePieChart from '../../../components/profile/GenrePieChart';
import AddUserModal from '../../../components/search/AddUserModal';
import { UserDataResponse } from '../../../types/UserData';
import {GroupData} from '@/types/Groups';

interface ProfilePageProps {
  params: {
    id: string;
    data?: UserDataResponse;
    isOwn?: boolean;
    subs?: GroupData[];
  };
}

// Временные тестовые подписки (если их нет в API)
const testSubscriptions = [
  {
    id: 1,
    name: "Группа крутых пацанят",
    small_description: "47 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["games"]
  },
  // ... остальные подписки можно оставить или получать через API
];

export default function ProfilePage({ params }: ProfilePageProps) {
  // Проверка наличия данных
  if (!params.data) {
    notFound();
  }

  const profileData = params.data;
  console.log("profileData", profileData);
  const userId = parseInt(params.id);

  const [isEditMode, setIsEditMode] = useState(false);
  const [editedData, setEditedData] = useState({
    name: profileData.name,
    us: profileData.us, // Добавлено
    status: profileData.status,
    image: profileData.image,
    tiles: [...profileData.tiles]
  });
  const [selectedTileIndex, setSelectedTileIndex] = useState<number | null>(null);
  const [animatedChart, setAnimatedChart] = useState(false);

  const scrollRef = React.useRef<HTMLDivElement>(null);
  const [showUpArrow, setShowUpArrow] = useState(false);
  const [showDownArrow, setShowDownArrow] = useState(false);
  const [activeTab, setActiveTab] = useState<'recent' | 'upcoming'>('recent');
  const [isOwnProfile, setOwnProfile] = useState(params.isOwn || false);
  const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);

  // Анимация диаграммы при загрузке
  useEffect(() => {
    const timer = setTimeout(() => {
      setAnimatedChart(true);
    }, 500);
    return () => clearTimeout(timer);
  }, []);

  // Цвета для диаграммы
  const chartColors = ['#316BC2', '#1E5CB9', '#1851A6', '#134693', '#0D3E86', '#0A3677'];

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
    const stats = profileData.user_stats;
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
    // Логика сохранения (здесь нужно добавить API запрос)
    setIsEditMode(false);
    console.log('Сохранение данных:', editedData);
  };

  const handleCancel = () => {
    setEditedData({
      name: profileData.name,
      us: profileData.us, // Добавлено
      status: profileData.status,
      image: profileData.image,
      tiles: [...profileData.tiles]
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
    if (!profileData.telegram_link) {
      window.open('https://t.me/FriendShipNotify_bot', '_blank');
    }
  };

  const handleScroll = () => {
    if (scrollRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = scrollRef.current;
      setShowUpArrow(scrollTop > 0);
      setShowDownArrow(scrollTop + clientHeight < scrollHeight);
    }
  };

  const scrollUp = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollBy({ top: -100, behavior: 'smooth' });
    }
  };

  const scrollDown = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollBy({ top: 100, behavior: 'smooth' });
    }
  };

  useEffect(() => {
    handleScroll();
  }, []);

  const chartData = useMemo(() => {
    const topGenres = profileData.popular_genres.slice(0, 5);
    const totalTopCount = topGenres.reduce((sum, genre) => sum + genre.count, 0);
    const otherCount = Math.max(0, Math.floor(totalTopCount * 0.2));

    const data = topGenres.map((genre, index) => ({
      name: genre.name,
      value: genre.count,
      color: chartColors[index]
    }));

    if (otherCount > 0) {
      data.push({ name: 'Другое', value: otherCount, color: chartColors[5] });
    }

    return data;
  }, [profileData.popular_genres]);

  const handleAddUserClick = () => {
    setIsAddUserModalOpen(true);
  };

  return (
    <div className="bgPage">
      <div className={styles.profileContainer}>
        {/* Секция 1: Краткая информация */}
        <div className={section1Styles.profileInfo}>
          <div className={section1Styles.settingsButton}>
            {!isOwnProfile ? (
              <button 
                onClick={handleAddUserClick}
                className={section1Styles.addUserButton}
              >
                <Image src="/add_button.png" alt="add user" width={32} height={32} />
                <div className={section1Styles.addUserTooltip}>
                  Добавить в группу
                </div>
              </button>
            ) : !isEditMode ? (
              <button onClick={handleSettingsClick} className={section1Styles.settingsBtn}>
                <Image src="/setting.png" alt="settings" width={24} height={24} />
              </button>
            ) : (
              <div className={section1Styles.editButtons}>
                <button onClick={handleSave} className={section1Styles.editBtn}>
                  <Image src="/profile/edit_save.png" alt="save" width={20} height={20} />
                </button>
                <button onClick={handleCancel} className={section1Styles.editBtn}>
                  <Image src="/profile/edit_cancel.png" alt="cancel" width={20} height={20} />
                </button>
              </div>
            )}
          </div>

          <div className={section1Styles.userHeader}>
            <div className={section1Styles.avatarContainer}>
              <Image 
                src={editedData.image} 
                alt="avatar" 
                width={120} 
                height={120}
                className={section1Styles.avatar}
              />
              {isEditMode && (
                <div className={section1Styles.editIcon}>
                  <Image src="/profile/edit.png" alt="edit" width={16} height={16} />
                </div>
              )}
            </div>

            <div className={section1Styles.userInfo}>
              <div className={section1Styles.userNameContainer}>
                <div className={section1Styles.nameWrapper}>
                  {isEditMode ? (
                    <input 
                      value={editedData.name}
                      onChange={(e) => setEditedData({...editedData, name: e.target.value})}
                      className={section1Styles.nameInput}
                    />
                  ) : (
                    <h2 className={section1Styles.userName}>{editedData.name}</h2>
                  )}
                  {profileData.enterprise && !isEditMode && (
                    <Image src="/profile/mark.png" alt="verified" width={20} height={20} className={section1Styles.verifiedMark} />
                  )}
                  {isEditMode && (
                    <Image src="/profile/edit.png" alt="edit" width={16} height={16} className={section1Styles.editIcon} />
                  )}
                </div>
                
                {isOwnProfile && !isEditMode && (
                  <div className={section1Styles.telegramContainer}>
                    <div 
                      className={`${section1Styles.telegramIcon}`}
                      onClick={handleTelegramClick}
                    >
                      <Image 
                        src={!profileData.telegram_link ? "/social/tg_grey.png" : getSocialIcon("tg")} 
                        alt="telegram" 
                        width={24} 
                        height={24}
                      />
                      {!profileData.telegram_link && (
                        <div className={section1Styles.tooltip}>
                          У вас не привязан телеграм, чтобы получать напоминания о событиях, на которые вы записались. Но это можно исправить зайдя в этого бота в телеграме: <span className={section1Styles.botLink}>@FriendShipNotify_bot</span>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>

              <div className={section1Styles.userUsContainer}>
                {isEditMode ? (
                  <div className={section1Styles.usEditWrapper}>
                    <input 
                      value={editedData.us}
                      onChange={(e) => setEditedData({...editedData, us: e.target.value})}
                      className={section1Styles.usInput}
                      placeholder="username"
                    />
                    <Image 
                      src="/profile/edit.png" 
                      alt="edit" 
                      width={14} 
                      height={14} 
                      className={section1Styles.editIconUs} 
                    />
                  </div>
                ) : (
                  <p className={section1Styles.userUs}>@{editedData.us}</p>
                )}
              </div>

              <div className={section1Styles.userStatusContainer}>
                {isEditMode ? (
                  <div className={section1Styles.statusEditWrapper}>
                    <textarea 
                      value={editedData.status}
                      onChange={(e) => setEditedData({...editedData, status: e.target.value})}
                      className={section1Styles.statusInput}
                      rows={2}
                      placeholder="Введите статус"
                    />
                    <Image 
                      src="/profile/edit.png" 
                      alt="edit" 
                      width={14} 
                      height={14} 
                      className={section1Styles.editIconStatus} 
                    />
                  </div>
                ) : (
                  <p className={section1Styles.userStatus}>{editedData.status}</p>
                )}
                <div className={section1Styles.registerDate}>
                  Участник с {profileData.data_register}
                </div>
              </div>
            </div>
          </div>

          {/* Статистические плитки */}
          <div className={section1Styles.statsTiles}>
            {editedData.tiles.map((tileType, index) => (
              <div key={index} className={section1Styles.statTile} onClick={() => handleTileClick(index)}>
                <div className={section1Styles.tileIcon}>
                  <Image 
                    src={getTileIcon(tileType)} 
                    alt={tileType} 
                    width={24} 
                    height={24}
                  />
                </div>
                <div className={section1Styles.tileValue}>{getTileValue(tileType)}</div>
                <div className={section1Styles.tileLabel}>{getTileLabel(tileType)}</div>
                {isEditMode && (
                  <div className={section1Styles.editIcon}>
                    <Image src="/profile/edit.png" alt="edit" width={12} height={12} />
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Модальное окно выбора плитки */}
          {selectedTileIndex !== null && (
            <div className={section1Styles.tileModal}>
              <div className={section1Styles.tileModalContent}>
                <h4>Выберите тип плитки:</h4>
                <div className={section1Styles.tileOptions}>
                  {availableTiles.map((tileType) => (
                    <div 
                      key={tileType} 
                      className={section1Styles.tileOption}
                      onClick={() => handleTileChange(tileType)}
                    >
                      <Image src={getTileIcon(tileType)} alt={tileType} width={20} height={20} />
                      <span>{getTileLabel(tileType)}</span>
                    </div>
                  ))}
                </div>
                <button onClick={() => setSelectedTileIndex(null)} className={section1Styles.closeModal}>
                  Отмена
                </button>
              </div>
            </div>
          )}

          {/* Топ жанров */}
          <div className={section1Styles.topGenres}>
            <h4>Топ любимых жанров:</h4>
            <div className={section1Styles.genresList}>
              {profileData.popular_genres.map((genre, index) => (
                <div key={index} className={section1Styles.genreItem}>
                  <span className={section1Styles.genreNumber}>{index + 1}.</span>
                  <span className={section1Styles.genreName}>{genre.name}</span>
                  <span className={section1Styles.genreCount}>{genre.count}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Секция 2: Статистика */}
        <div className={section2Styles.statisticsSection}>
          <h3>{isOwnProfile ? 'Ваша статистика' : 'Статистика'}</h3>
          <div className={section2Styles.statisticsContent}>
            {/* Диаграмма жанров */}
            <div className={section2Styles.chartContainer}>
              <div className={section2Styles.chartWrapper}>
                <div className={section2Styles.chartWithLegend}>
                  <GenrePieChart data={chartData} animated={animatedChart} />
                  <div className={section2Styles.customLegend}>
                    <h4 className={section2Styles.genresTitle}>Жанры:</h4>
                    {chartData.map((entry, index) => (
                      <div key={index} className={section2Styles.legendItem}>
                        <span className={section2Styles.legendNumber}>{index + 1}.</span>
                        <span className={section2Styles.legendName}>{entry.name}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Плитки статистики */}
            <div className={section2Styles.statisticsTiles}>
              <StatisticsTile
                title="Самый популярный день"
                value={profileData.user_stats.most_pop_day}
                icon="/profile/calendar.png"
              />
              <StatisticsTile
                title="Создано сессий"
                value={profileData.user_stats.count_create_session}
                icon="/profile/create_session.png"
              />
              <StatisticsTile
                title="Серия сессий"
                value={`${profileData.user_stats.series_session_count} подряд`}
                icon="/profile/series.png"
              />
            </div>
          </div>
        </div>

        {/* Секция 3: Подписки */}
        <div className={section3Styles.subscriptionsSection}>
          <h3 className={section3Styles.subscriptionsTitle}>
            {isOwnProfile ? 'Ваши подписки' : 'Подписки'}
          </h3>

          <div className={section3Styles.scrollWrapper}>
            {showUpArrow && (
              <button
                className={`${section3Styles.scrollArrow} ${section3Styles.topArrow}`}
                onClick={scrollUp}
              >
                ↑
              </button>
            )}

            <div
              className={section3Styles.groupsList}
              ref={scrollRef}
              onScroll={handleScroll}
            >
              {testSubscriptions.map((group) => (
                <SubscriptionItem
                  key={group.id}
                  id={group.id}
                  name={group.name}
                  small_description={group.small_description}
                  image={group.image}
                />
              ))}
            </div>

            {showDownArrow && (
              <button
                className={`${section3Styles.scrollArrow} ${section3Styles.bottomArrow}`}
                onClick={scrollDown}
              >
                ↓
              </button>
            )}
          </div>
        </div>

        {/* Секция 4: События */}
        <div className={section4Styles.eventsSection}>
          <div className={section4Styles.header}>
            <h3>События</h3>
            
            {isOwnProfile && (
              <div className={section4Styles.buttonsGroup}>
                <button
                  type="button"
                  className={`${section4Styles.tabButton} ${activeTab === 'recent' ? section4Styles.active : ''}`}
                  onClick={() => setActiveTab('recent')}
                  aria-pressed={activeTab === 'recent'}
                >
                  <Image
                    src="/profile/recent_sessions.png"
                    alt="Завершённые"
                    width={32}
                    height={32}
                  />
                </button>

                <button
                  type="button"
                  className={`${section4Styles.tabButton} ${activeTab === 'upcoming' ? section4Styles.active : ''}`}
                  onClick={() => setActiveTab('upcoming')}
                  aria-pressed={activeTab === 'upcoming'}
                >
                  <Image
                    src="/profile/upcoming_sessions.png"
                    alt="Будущие"
                    width={32}
                    height={32}
                  />
                </button>
              </div>
            )}
          </div>

          <div className={section4Styles.content}>
            <CategorySection
              section={
                isOwnProfile 
                  ? (activeTab === 'recent'
                      ? {categories: profileData.recent_sessions}
                      : {categories: profileData.upcoming_sessions})
                  : {categories: profileData.recent_sessions}
              }
              title=""
              showCategoryLabel={false}
            />
          </div>
        </div>
        
        <AddUserModal
          isOpen={isAddUserModalOpen}
          onClose={() => setIsAddUserModalOpen(false)}
          userId={userId}
        />
      </div>
    </div>
  );
}