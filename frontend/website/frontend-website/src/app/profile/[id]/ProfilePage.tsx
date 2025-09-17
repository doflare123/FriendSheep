'use client';

import React, { useState, useEffect, useMemo  } from 'react';
import Image from 'next/image';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import styles from '../../../styles/profile/ProfilePage.module.css';
import section1Styles from '../../../styles/profile/ProfileSection1.module.css';
import section2Styles from '../../../styles/profile/ProfileSection2.module.css';
import section3Styles from '../../../styles/profile/ProfileSection3.module.css';
import section4Styles from '../../../styles/profile/ProfileSection4.module.css';
import StatisticsTile from '../../../components/profile/StatisticsTile';
import SubscriptionItem from '../../../components/profile/SubscriptionItem';
import CategorySection from '../../../components/Events/CategorySection';
import {getSocialIcon} from '../../../Constants';
import GenrePieChart from '../../../components/profile/GenrePieChart';

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
    { count: 1, name: "Анекдоты" }
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
  recent_sessions: [
    {
      id: '1',
      type: 'games',
      image: '/event_card.jpg',
      date: '12 сен',
      title: 'Мафия в центре',
      genres: ['Настольные игры'],
      participants: 8,
      maxParticipants: 12,
      duration: '2ч',
      location: 'offline',
    },
    {
      id: '2',
      type: 'movies',
      image: '/event_card.jpg',
      date: '5 сен',
      title: 'Киновечер: Интерстеллар',
      genres: ['Фантастика'],
      participants: 15,
      maxParticipants: 20,
      duration: '3ч',
      location: 'offline',
    },
    {
      id: '3',
      type: 'board',
      image: '/event_card.jpg',
      date: '1 сен',
      title: 'Монополия батл',
      genres: ['Экономика'],
      participants: 6,
      maxParticipants: 6,
      duration: '1ч 30м',
      location: 'online',
    },
    {
      id: '4',
      type: 'games',
      image: '/event_card.jpg',
      date: '28 авг',
      title: 'Квиз вечер',
      genres: ['Викторина'],
      participants: 10,
      maxParticipants: 12,
      duration: '2ч',
      location: 'offline',
    },
    {
      id: '5',
      type: 'other',
      image: '/event_card.jpg',
      date: '20 авг',
      title: 'Караоке-вечеринка',
      genres: ['Музыка'],
      participants: 12,
      maxParticipants: 15,
      duration: '3ч',
      location: 'offline',
    },
    {
      id: '6',
      type: 'movies',
      image: '/event_card.jpg',
      date: '15 авг',
      title: 'Фильм: Матрица',
      genres: ['Фантастика'],
      participants: 10,
      maxParticipants: 15,
      duration: '2ч 30м',
      location: 'online',
    },
  ],

  upcoming_sessions: [
    {
      id: '7',
      type: 'games',
      image: '/event_card.jpg',
      date: '20 сен',
      title: 'Among Us Online',
      genres: ['Кооператив'],
      participants: 5,
      maxParticipants: 10,
      duration: '1ч',
      location: 'online',
    },
    {
      id: '8',
      type: 'movies',
      image: '/event_card.jpg',
      date: '25 сен',
      title: 'Показ фильма: Начало',
      genres: ['Триллер'],
      participants: 20,
      maxParticipants: 30,
      duration: '2ч 30м',
      location: 'offline',
    },
    {
      id: '9',
      type: 'board',
      image: '/event_card.jpg',
      date: '28 сен',
      title: 'Каркассон турнир',
      genres: ['Настолки'],
      participants: 8,
      maxParticipants: 12,
      duration: '2ч',
      location: 'offline',
    },
    {
      id: '10',
      type: 'games',
      image: '/event_card.jpg',
      date: '30 сен',
      title: 'CS:GO турнир',
      genres: ['Шутер'],
      participants: 10,
      maxParticipants: 10,
      duration: '3ч',
      location: 'online',
    },
    {
      id: '11',
      type: 'other',
      image: '/event_card.jpg',
      date: '2 окт',
      title: 'Кулинарный мастер-класс',
      genres: ['Еда'],
      participants: 6,
      maxParticipants: 10,
      duration: '2ч',
      location: 'offline',
    },
    {
      id: '12',
      type: 'movies',
      image: '/event_card.jpg',
      date: '5 окт',
      title: 'Вечер комедий',
      genres: ['Комедия'],
      participants: 12,
      maxParticipants: 20,
      duration: '2ч',
      location: 'offline',
    },
    {
      id: '13',
      type: 'games',
      image: '/event_card.jpg',
      date: '10 окт',
      title: 'PUBG Squad Night',
      genres: ['Battle Royale'],
      participants: 4,
      maxParticipants: 4,
      duration: '2ч',
      location: 'online',
    },
  ],
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
    name: "Киноманы",
    small_description: "128 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["movies"]
  },
  {
    id: 3,
    name: "Настольщики",
    small_description: "63 участника",
    image: "/default/group.jpg",
    type: "group",
    category: ["board"]
  },
  {
    id: 4,
    name: "Футбольный чат",
    small_description: "212 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["sport"]
  },
  {
    id: 5,
    name: "Любители музыки",
    small_description: "98 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["music"]
  },
  {
    id: 6,
    name: "Путешественники",
    small_description: "76 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["travel"]
  },
  {
    id: 7,
    name: "Игроманы",
    small_description: "301 участник",
    image: "/default/group.jpg",
    type: "group",
    category: ["games"]
  },
  {
    id: 8,
    name: "Аниме клуб",
    small_description: "145 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["anime"]
  },
  {
    id: 9,
    name: "IT тусовка",
    small_description: "512 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["tech"]
  },
  {
    id: 10,
    name: "Книжный клуб",
    small_description: "84 участника",
    image: "/default/group.jpg",
    type: "group",
    category: ["books"]
  },
  {
    id: 11,
    name: "Фанаты Marvel",
    small_description: "133 участника",
    image: "/default/group.jpg",
    type: "group",
    category: ["movies"]
  },
  {
    id: 12,
    name: "Сериальщики",
    small_description: "210 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["series"]
  },
  {
    id: 13,
    name: "Гурманы",
    small_description: "59 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["food"]
  },
  {
    id: 14,
    name: "Беговой клуб",
    small_description: "42 участника",
    image: "/default/group.jpg",
    type: "group",
    category: ["sport"]
  },
  {
    id: 15,
    name: "Фотографы",
    small_description: "77 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["photo"]
  },
  {
    id: 16,
    name: "Художники",
    small_description: "64 участника",
    image: "/default/group.jpg",
    type: "group",
    category: ["art"]
  },
  {
    id: 17,
    name: "Киберспорт",
    small_description: "189 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["games"]
  },
  {
    id: 18,
    name: "Фанаты DC",
    small_description: "101 участник",
    image: "/default/group.jpg",
    type: "group",
    category: ["movies"]
  },
  {
    id: 19,
    name: "Мастера настолок",
    small_description: "55 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["board"]
  },
  {
    id: 20,
    name: "Поклонники классики",
    small_description: "39 участников",
    image: "/default/group.jpg",
    type: "group",
    category: ["music"]
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

  const scrollRef = React.useRef<HTMLDivElement>(null);
  const [showUpArrow, setShowUpArrow] = useState(false);
  const [showDownArrow, setShowDownArrow] = useState(false);
  const [activeTab, setActiveTab] = useState<'recent' | 'upcoming'>('recent');

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
    const topGenres = testProfileData.popular_genres.slice(0, 5);
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
  }, [testProfileData.popular_genres]);

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
                  <GenrePieChart data={chartData} animated={animatedChart} />
                  {/* кастомная легенда остаётся как была */}
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

        {/* Секция 3: Подписки */}
        <div className={section3Styles.subscriptionsSection}>
          <h3 className={section3Styles.subscriptionsTitle}>Ваши подписки</h3>

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
            {/* Кнопки сразу справа от заголовка */}
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
          </div>

          <div className={section4Styles.content}>
            <CategorySection
              section={
                activeTab === 'recent'
                  ? {categories: testProfileData.recent_sessions}
                  : {categories: testProfileData.upcoming_sessions}
              }
              title=""
              showCategoryLabel={false}
            />
          </div>
        </div>
      </div>
    </div>
  );
}