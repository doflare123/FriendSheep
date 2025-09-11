'use client';

import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import styles from '../../../styles/profile/ProfilePage.module.css';
import section1Styles from '../../../styles/profile/ProfileSection1.module.css';
import section2Styles from '../../../styles/profile/ProfileSection2.module.css';
import StatisticsTile from '../../../components/profile/StatisticsTile';
import {getSocialIcon} from '../../../Constants';

// Тестовые данные на основе API
const testProfileData = {
  data_register: "21.01.2024",
  enterprise: true,
  image: "/default-avatar.png",
  name: "Чебаксар",
  username: "user",
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
    spent_time: 20,
    most_popular_day: "Воскресенье" // добавлено для статистики
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
    username: testProfileData.username,
    status: testProfileData.status,
    image: testProfileData.image,
    tiles: [...testProfileData.tiles]
  });
  const [selectedTileIndex, setSelectedTileIndex] = useState<number | null>(null);
  const [animatedChart, setAnimatedChart] = useState(false);

  // Анимация диаграммы при загрузке
  useEffect(() => {
    const timer = setTimeout(() => {
      setAnimatedChart(true);
    }, 500);
    return () => clearTimeout(timer);
  }, []);

  // Цвета для диаграммы
  const chartColors = ['#316BC2', '#1E5CB9', '#1851A6', '#134693', '#0D3E86', '#0A3677'];

  // Подготовка данных для диаграммы
  const prepareChartData = () => {
    const topGenres = testProfileData.popular_genres.slice(0, 5);
    const totalTopCount = topGenres.reduce((sum, genre) => sum + genre.count, 0);
    
    // Предполагаем, что есть другие жанры
    const otherCount = Math.max(0, Math.floor(totalTopCount * 0.2)); // примерно 20% от топа
    
    const chartData = topGenres.map((genre, index) => ({
      name: genre.name,
      value: genre.count,
      color: chartColors[index]
    }));

    if (otherCount > 0) {
      chartData.push({
        name: 'Другое',
        value: otherCount,
        color: chartColors[5]
      });
    }

    return chartData;
  };

  const chartData = prepareChartData();

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
      username: testProfileData.username,
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

  const handleUsernameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    if (value.length <= 10) {
      setEditedData({...editedData, username: value});
    }
  };

  // Кастомный tooltip для диаграммы
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0];
      return (
        <div style={{
          background: 'white',
          border: '2px solid #37A2E6',
          borderRadius: '8px',
          padding: '8px 12px',
          fontSize: '14px',
          color: '#000',
          pointerEvents: 'none',
          zIndex: 1000
        }}>
          <p style={{ margin: 0, fontWeight: 'bold' }}>{data.name}</p>
          <p style={{ margin: 0, color: '#666' }}>Количество: {data.value}</p>
        </div>
      );
    }
    return null;
  };

  // Кастомная анимированная метка для номеров
  const AnimatedNumberLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, index }: any) => {
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.4; // Ближе к центру
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    return (
      <text 
        x={x} 
        y={y} 
        fill="white" 
        textAnchor="middle" 
        dominantBaseline="central"
        fontSize="18"
        fontWeight="bold"
        style={{
          opacity: animatedChart ? 1 : 0,
          transition: `opacity 0.8s ease-in-out ${0.2 * index}s`,
          pointerEvents: 'none'
        }}
      >
        {index + 1}
      </text>
    );
  };

  // Кастомная анимированная метка для процентов
  const AnimatedPercentLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, percent, index }: any) => {
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.8; // Дальше от центра
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    return (
      <text 
        x={x} 
        y={y} 
        fill="white" 
        textAnchor="middle" 
        dominantBaseline="central"
        fontSize="14"
        fontWeight="bold"
        style={{
          opacity: animatedChart ? 1 : 0,
          transition: `opacity 0.8s ease-in-out ${0.4 + 0.2 * index}s`,
          pointerEvents: 'none'
        }}
      >
        {`${(percent * 100).toFixed(0)}%`}
      </text>
    );
  };

  return (
    <div className="bgPage">
      <div className={styles.profileContainer}>
        {/* Секция 1: Краткая информация */}
        <div className={section1Styles.profileInfo}>
          <div className={section1Styles.settingsButton}>
            {!isEditMode ? (
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
                  {testProfileData.enterprise && !isEditMode && (
                    <Image src="/profile/mark.png" alt="verified" width={20} height={20} className={section1Styles.verifiedMark} />
                  )}
                  {isEditMode && (
                    <Image src="/profile/edit.png" alt="edit" width={16} height={16} className={section1Styles.editIcon} />
                  )}
                </div>
                
                {!isEditMode ? (
                  <div className={section1Styles.telegramContainer}>
                    <div 
                      className={`${section1Styles.telegramIcon}`}
                      onClick={handleTelegramClick}
                    >
                      <Image 
                        src={!testProfileData.telegram_link ? "/social/tg_grey.png" : getSocialIcon("tg")} 
                        alt="telegram" 
                        width={24} 
                        height={24}
                      />
                      {!testProfileData.telegram_link && (
                        <div className={section1Styles.tooltip}>
                          У вас не привязан телеграм, чтобы получать напоминания о событиях, на которые вы записались. Но это можно исправить зайдя в этого бота в телеграме: <span className={section1Styles.botLink}>@FriendShipNotify_bot</span>
                        </div>
                      )}
                    </div>
                  </div>
                ) : null}
              </div>
              
              <div className={section1Styles.userContainer}>
                <div className={section1Styles.usernameContainer}>
                  {isEditMode ? (
                    <div className={section1Styles.usernameEdit}>
                      <input 
                        value={editedData.username}
                        onChange={handleUsernameChange}
                        className={section1Styles.usernameInput}
                        placeholder="username"
                        maxLength={10}
                      />
                      <Image src="/profile/edit.png" alt="edit" width={16} height={16} className={section1Styles.editIcon} />
                    </div>
                  ) : (
                    <div className={section1Styles.userText}>@{editedData.username}</div>
                  )}
                </div>
              </div>

              <div className={section1Styles.userStatusContainer}>
                {isEditMode ? (
                  <textarea 
                    value={editedData.status}
                    onChange={(e) => setEditedData({...editedData, status: e.target.value})}
                    className={section1Styles.statusInput}
                    rows={2}
                    placeholder="Введите статус"
                  />
                ) : (
                  <p className={section1Styles.userStatus}>{editedData.status}</p>
                )}
                {isEditMode && (
                  <Image src="/profile/edit.png" alt="edit" width={16} height={16} className={section1Styles.editIcon} />
                )}
                <div className={section1Styles.registerDate}>
                  Участник с {testProfileData.data_register}
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
              {testProfileData.popular_genres.map((genre, index) => (
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
          <h3>Ваша статистика</h3>
          <div className={section2Styles.statisticsContent}>
            {/* Диаграмма жанров */}
            <div className={section2Styles.chartContainer}>
              <div className={section2Styles.chartWrapper}>
                <div className={section2Styles.chartWithLegend}>
                  <ResponsiveContainer width="60%" height={320}>
                    <PieChart>
                      <Pie
                        data={chartData}
                        cx="50%"
                        cy="50%"
                        outerRadius={130}
                        dataKey="value"
                        labelLine={false}
                        label={AnimatedNumberLabel}
                      >
                        {chartData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Pie
                        data={chartData}
                        cx="50%"
                        cy="50%"
                        outerRadius={130}
                        dataKey="value"
                        labelLine={false}
                        label={AnimatedPercentLabel}
                      >
                        {chartData.map((entry, index) => (
                          <Cell key={`cell-percent-${index}`} fill="transparent" />
                        ))}
                      </Pie>
                      <Tooltip content={<CustomTooltip />} />
                    </PieChart>
                  </ResponsiveContainer>
                  
                  {/* Кастомная легенда */}
                  <div className={section2Styles.customLegend}>
                    <h4 className={section2Styles.genresTitle}>Жанры</h4>
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
                value={testProfileData.user_stats.most_popular_day}
                icon="/profile/calendar.png"
              />
              <StatisticsTile
                title="Создано сессий"
                value={testProfileData.user_stats.count_create_session}
                icon="/profile/create_session.png"
              />
              <StatisticsTile
                title="Серия сессий"
                value={`${testProfileData.user_stats.series_session_count} подряд`}
                icon="/profile/series.png"
              />
            </div>
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