'use client';

import React, { useState, useEffect, useMemo, useRef } from 'react';
import Image from 'next/image';
import { notFound, useRouter } from 'next/navigation';
import styles from '@/styles/profile/ProfilePage.module.css';
import section1Styles from '@/styles/profile/ProfileSection1.module.css';
import section2Styles from '@/styles/profile/ProfileSection2.module.css';
import section3Styles from '@/styles/profile/ProfileSection3.module.css';
import section4Styles from '@/styles/profile/ProfileSection4.module.css';
import StatisticsTile from '@/components/profile/StatisticsTile';
import SubscriptionItem from '@/components/profile/SubscriptionItem';
import CategorySection from '@/components/Events/CategorySection';
import { getSocialIcon, getAccesToken } from '@/Constants';
import GenrePieChart from '@/components/profile/GenrePieChart';
import AddUserModal from '@/components/search/AddUserModal';
import { UserDataResponse } from '@/types/UserData';
import {SmallGroup} from '@/types/Groups';
import {Counters, UpdateProfileRequest} from '@/types/apiTypes';
import {editProfile} from '@/api/profile/editProfile';
import {editTiles} from '@/api/profile/editTiles';
import {getImage} from '@/api/getImage';

import {useAuth} from '@/contexts/AuthContext'

import {showNotification} from '@/utils';

interface ProfilePageProps {
  params: {
    id: string;
    data?: UserDataResponse;
    isOwn?: boolean;
    subs?: SmallGroup[];
  };
}

export default function ProfilePage({ params }: ProfilePageProps) {
  // Проверка наличия данных
  if (!params.data) {
    notFound();
  }

  const { forceRefreshToken } = useAuth();
  const profileData = params.data;
  const subsData = params.subs;
  const userId = parseInt(params.id);

  const [isEditMode, setIsEditMode] = useState(false);
  const [editedData, setEditedData] = useState({
    name: profileData.name,
    us: profileData.us,
    status: profileData.status,
    image: profileData.image,
    tiles: [...profileData.tiles]
  });
  const [selectedTileIndex, setSelectedTileIndex] = useState<number | null>(null);
  const [animatedChart, setAnimatedChart] = useState(false);
  const [avatarFile, setAvatarFile] = useState<File | null>(null);

  const [errors, setErrors] = useState({
    name: '',
    us: '',
    status: ''
  });

  const scrollRef = useRef<HTMLDivElement>(null);
  const avatarInputRef = useRef<HTMLInputElement>(null);
  const [showUpArrow, setShowUpArrow] = useState(false);
  const [showDownArrow, setShowDownArrow] = useState(false);
  const [activeTab, setActiveTab] = useState<'recent' | 'upcoming'>('recent');
  const [isOwnProfile, setOwnProfile] = useState(params.isOwn || false);
  const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);

  const [isSaving, setIsSaving] = useState(false);
  const router = useRouter();

  // Функция валидации
  const validateFields = () => {
    const newErrors = {
      name: '',
      us: '',
      status: ''
    };

    // Валидация имени (мин. 5, макс. 40)
    if (!editedData.name.trim()) {
      newErrors.name = 'Имя обязательно для заполнения';
    } else if (editedData.name.trim().length < 5) {
      newErrors.name = 'Имя должно содержать минимум 5 символов';
    } else if (editedData.name.trim().length > 40) {
      newErrors.name = 'Имя должно содержать максимум 40 символов';
    }

    // Валидация username (мин. 5, макс. 40)
    if (!editedData.us.trim()) {
      newErrors.us = 'Username обязателен для заполнения';
    } else if (editedData.us.trim().length < 5) {
      newErrors.us = 'Username должен содержать минимум 5 символов';
    } else if (editedData.us.trim().length > 40) {
      newErrors.us = 'Username должен содержать максимум 40 символов';
    }

    // Валидация статуса (мин. 1 если есть, макс. 50)
    if (editedData.status.trim() && editedData.status.trim().length < 1) {
      newErrors.status = 'Статус должен содержать минимум 1 символ';
    } else if (editedData.status.length > 50) {
      newErrors.status = 'Статус должен содержать максимум 50 символов';
    }

    setErrors(newErrors);
    return !newErrors.name && !newErrors.us && !newErrors.status;
  };

  // Функция для расчета размера шрифта в зависимости от длины текста
  const calculateFontSize = (text: string, maxSize: number, minSize: number, maxLength: number) => {
    const length = text.length;
    if (length === 0) return maxSize;
    if (length <= maxLength / 2) return maxSize;
    
    const ratio = Math.min(length / maxLength, 1);
    const fontSize = maxSize - (maxSize - minSize) * ratio;
    return Math.max(fontSize, minSize);
  };

  const nameFontSize = calculateFontSize(editedData.name, 22, 14, 40);
  const statusFontSize = calculateFontSize(editedData.status, 16, 12, 50);
  const usFontSize = calculateFontSize(editedData.us, 16, 12, 40);

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

  const handleSave = async () => {
    // Валидация перед сохранением
    if (!validateFields()) {
      showNotification(400, "Проверьте правильность заполнения полей");
      return;
    }

    const accessToken = await getAccesToken(router);
    if (!accessToken) {
      console.error("❌ Нет accessToken");
      showNotification(401, "Нет токена доступа");
      return;
    }

    try {
      setIsSaving(true);

      let imageUrl = editedData.image;

      if (avatarFile) {
        imageUrl = await getImage(accessToken, avatarFile);
      }

      const profileData: UpdateProfileRequest = {
        name: editedData.name,
        us: editedData.us,
        status: editedData.status,
        image: imageUrl,
      };

      await editProfile(accessToken, profileData);

      const counters: Counters = {
        count_all: editedData.tiles.includes("count_all"),
        count_films: editedData.tiles.includes("count_films"),
        count_games: editedData.tiles.includes("count_games"),
        count_other: editedData.tiles.includes("count_other"),
        count_table: editedData.tiles.includes("count_table"),
        spent_time: editedData.tiles.includes("spent_time"),
      };

      await editTiles(accessToken, counters);
      await forceRefreshToken();
      router.replace('/profile/' + editedData.us);

      // Выходим из режима редактирования перед навигацией
      setIsEditMode(false);
      setIsSaving(false);

      showNotification(200, "Изменения успешно сохранены");
      
      // Используем replace для избежания записи в историю
      

    } catch (error: any) {
      console.error("❌ Ошибка при сохранении профиля:", error);
      showNotification(
        error.response?.status || 500,
        "Не удалось сохранить изменения",
      );
      setIsSaving(false);
    }
  };


  const handleCancel = () => {
    setEditedData({
      name: profileData.name,
      us: profileData.us,
      status: profileData.status,
      image: profileData.image,
      tiles: [...profileData.tiles]
    });
    setAvatarFile(null);
    setErrors({ name: '', us: '', status: '' });
    setIsEditMode(false);
  };

  const handleTileClick = (index: number) => {
    if (isEditMode) {
      setSelectedTileIndex(index);
    }
  };

  const handleTileChange = (newTileType: string) => {
    if (selectedTileIndex !== null) {
      // Проверяем, используется ли уже эта плитка
      const isAlreadyUsed = editedData.tiles.some(
        (tile, index) => tile === newTileType && index !== selectedTileIndex
      );
      
      if (isAlreadyUsed) {
        showNotification(400, "Эта плитка уже используется");
        return;
      }
      
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

  const handleAvatarClick = () => {
    if (isEditMode && avatarInputRef.current) {
      avatarInputRef.current.click();
    }
  };

  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setAvatarFile(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setEditedData({...editedData, image: reader.result as string});
      };
      reader.readAsDataURL(file);
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
                <button 
                  onClick={handleSave} 
                  className={section1Styles.editBtn} 
                  disabled={isSaving}
                >
                  {isSaving ? (
                    <Image src="/loader.gif" alt="loading" width={20} height={20} />
                  ) : (
                    <Image src="/profile/edit_save.png" alt="save" width={20} height={20} />
                  )}
                </button>
                <button 
                  onClick={handleCancel} 
                  className={section1Styles.editBtn}
                  disabled={isSaving}
                >
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
                style={{ cursor: isEditMode ? 'pointer' : 'default' }}
                onClick={handleAvatarClick}
              />
              {isEditMode && (
                <>
                  <input
                    ref={avatarInputRef}
                    type="file"
                    accept="image/*"
                    onChange={handleAvatarChange}
                    className={section1Styles.avatarInput}
                  />
                  <div className={section1Styles.avatarEditIcon} onClick={handleAvatarClick}>
                    <Image src="/profile/edit.png" alt="edit" width={16} height={16} />
                  </div>
                </>
              )}
            </div>

            <div className={section1Styles.userInfo}>
              <div className={section1Styles.userNameContainer}>
                <div className={section1Styles.nameWrapper}>
                  {isEditMode ? (
                    <div style={{ position: 'relative' }}>
                      <input 
                        value={editedData.name}
                        onChange={(e) => {
                          const value = e.target.value.replace(/[^а-яА-ЯёЁa-zA-Z0-9\s]/g, '');
                          if (value.length <= 40) {
                            setEditedData({...editedData, name: value});
                            if (errors.name) {
                              setErrors({...errors, name: ''});
                            }
                          }
                        }}
                        className={`${section1Styles.nameInput} ${errors.name ? section1Styles.inputError : ''}`}
                        style={{ fontSize: `${nameFontSize}px` }}
                        title={editedData.name}
                      />
                      <div className={section1Styles.nameEditIcon}>
                        <Image src="/profile/edit.png" alt="edit" width={14} height={14} />
                      </div>
                      {errors.name && (
                        <div className={section1Styles.errorMessage}>{errors.name}</div>
                      )}
                    </div>
                  ) : (
                    <h2 
                      className={section1Styles.userName}
                      style={{ fontSize: `${nameFontSize}px` }}
                      title={editedData.name}
                    >
                      {editedData.name}
                    </h2>
                  )}
                  {profileData.enterprise && !isEditMode && (
                    <Image src="/profile/mark.png" alt="verified" width={20} height={20} className={section1Styles.verifiedMark} />
                  )}
                </div>
                
                {isOwnProfile && !isEditMode && (
                  <div className={section1Styles.telegramContainer}>
                    <div 
                      className={section1Styles.telegramIcon}
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
                      onChange={(e) => {
                        const value = e.target.value.replace(/[^а-яА-ЯёЁa-zA-Z0-9\s]/g, '');
                        if (value.length <= 40) {
                          setEditedData({...editedData, us: value});
                          if (errors.us) {
                            setErrors({...errors, us: ''});
                          }
                        }
                      }}
                      className={`${section1Styles.usInput} ${errors.us ? section1Styles.inputError : ''}`}
                      style={{ fontSize: `${usFontSize}px` }}
                      placeholder="username"
                      title={editedData.us}
                    />
                    <div className={section1Styles.editIconUs}>
                      <Image 
                        src="/profile/edit.png" 
                        alt="edit" 
                        width={14} 
                        height={14}
                      />
                    </div>
                    {errors.us && (
                      <div className={section1Styles.errorMessage}>{errors.us}</div>
                    )}
                  </div>
                ) : (
                  <p 
                    className={section1Styles.userUs}
                    style={{ fontSize: `${usFontSize}px` }}
                    title={`@${editedData.us}`}
                  >
                    @{editedData.us}
                  </p>
                )}
              </div>

              <div className={section1Styles.userStatusContainer}>
                {isEditMode ? (
                  <div className={section1Styles.statusEditWrapper}>
                    <textarea 
                      value={editedData.status}
                      onChange={(e) => {
                        const value = e.target.value.replace(/[^а-яА-ЯёЁa-zA-Z0-9\s]/g, '');
                        if (value.length <= 50) {
                          setEditedData({...editedData, status: value});
                          if (errors.status) {
                            setErrors({...errors, status: ''});
                          }
                        }
                      }}
                      className={`${section1Styles.statusInput} ${errors.status ? section1Styles.inputError : ''}`}
                      style={{ fontSize: `${statusFontSize}px` }}
                      rows={2}
                      placeholder="Введите статус"
                      title={editedData.status}
                    />
                    <div className={section1Styles.statusEditIcon}>
                      <Image 
                        src="/profile/edit.png" 
                        alt="edit" 
                        width={14} 
                        height={14}
                      />
                    </div>
                    {errors.status && (
                      <div className={section1Styles.errorMessage}>{errors.status}</div>
                    )}
                  </div>
                ) : (
                  <p 
                    className={section1Styles.userStatus}
                    style={{ fontSize: `${statusFontSize}px` }}
                    title={editedData.status}
                  >
                    {editedData.status}
                  </p>
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
                    width={42} 
                    height={42}
                  />
                </div>
                <div className={section1Styles.tileValue}>{getTileValue(tileType)}</div>
                <div className={section1Styles.tileLabel}>{getTileLabel(tileType)}</div>
                {isEditMode && (
                  <div className={section1Styles.tileEditIcon}>
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
          {profileData.popular_genres.length > 0 && (
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
          )}
        </div>

        {/* Секция 2: Статистика */}
        <div className={section2Styles.statisticsSection}>
          <h3>{isOwnProfile ? 'Ваша статистика' : 'Статистика'}</h3>
          <div className={section2Styles.statisticsContent}>
            {/* Диаграмма жанров */}
            <div className={section2Styles.chartContainer}>
              <div className={section2Styles.chartWrapper}>
                {profileData.popular_genres.length > 0 ? (
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
                ) : (
                  <div className={styles.emptyPlaceholder}>
                    Здесь пока пусто
                  </div>
                )}
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
              {subsData && subsData.length > 0 ? (
                subsData.map((group) => (
                  <SubscriptionItem
                    key={group.id}
                    id={group.id}
                    name={group.name}
                    small_description={group.small_description}
                    image={group.image || "/default/group.jpg"}
                  />
                ))
              ) : (
                <div className={styles.emptyPlaceholder}>
                  Здесь пока пусто
                </div>
              )}
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
                  title = "Завершенные события"
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
                  title = "Наступающие события"
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
            {(() => {
              const sessions = isOwnProfile 
                ? (activeTab === 'recent' ? profileData.recent_sessions : profileData.upcoming_sessions)
                : profileData.recent_sessions;
              
              const hasEvents = sessions && sessions.length > 0;
              
              return hasEvents ? (
                <CategorySection
                  section={{
                    categories: sessions,
                    pattern: "" // или любой паттерн если нужен
                  }}
                  title=""
                  showCategoryLabel={false}
                />
              ) : (
                <div className={styles.emptyPlaceholderEvents}>
                  Здесь пока пусто
                </div>
              );
            })()}
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