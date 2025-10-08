import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import styles from '../../styles/Events/EventModal.module.css';
import MapModal from './MapModal';
import { EventCardProps } from '../../types/Events';
import { GroupData } from '../../types/Groups';
import { kinopoiskAPI, PlaceInfo } from '../../lib/api';
import { getCategoryIcon, getAccesToken, convertToRFC3339, parseDuration, convertCategEngToRu } from '../../Constants';
import { getOwnGroups } from '../../api/get_owngroups';
import { addEvent } from '@/api/events/addEvent';
import { getImage } from '@/api/getImage';
import { getGenres } from '@/api/events/getGenres';
import { showNotification } from '@/utils';
import LoadingIndicator from '@/components/LoadingIndicator';

interface EventModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: () => void;
  onDelete?: () => void;
  eventData?: Partial<EventCardProps>;
  mode?: 'create' | 'edit';
}

const CATEGORIES = [
  { id: 'movies', icon: getCategoryIcon("movies") },
  { id: 'games', icon: getCategoryIcon("games") },
  { id: 'board', icon: getCategoryIcon("board") },
  { id: 'other', icon: getCategoryIcon("other") }
];

interface ApiGroup {
  category: string[];
  id: number;
  image: string;
  member_count: number;
  name: string;
  small_description: string;
  type: string;
}

interface ValidationErrors {
  title?: string;
  genres?: string;
  maxParticipants?: string;
  selectedGroup?: string;
  image?: string;
  date?: string;
}

export default function EventModal({ 
  isOpen, 
  onClose, 
  onSave, 
  onDelete, 
  eventData, 
  mode = 'create' 
}: EventModalProps) {
  const [formData, setFormData] = useState<Partial<EventCardProps>>({
    title: '',
    type: 'movies',
    date: '',
    duration: '',
    location: 'offline',
    adress: '',
    participants: 0,
    maxParticipants: 0,
    genres: [],
    image: ''
  });

  const [isOnline, setIsOnline] = useState(false);
  const [showGenreSelector, setShowGenreSelector] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState<number | null>(null);
  const [publisher, setPublisher] = useState('');
  const [year, setYear] = useState('');
  const [ageLimit, setAgeLimit] = useState('');
  const [description, setDescription] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showMapModal, setShowMapModal] = useState(false);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [kinopoiskImageUrl, setKinopoiskImageUrl] = useState<string | null>(null);
  
  const [groups, setGroups] = useState<ApiGroup[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});
  const [isCreating, setIsCreating] = useState(false);
  
  const [genres, setGenres] = useState<string[]>([]);
  const [genresLoading, setGenresLoading] = useState(false);
  const [genresError, setGenresError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      loadGroups();
      loadGenres();
    }
  }, [isOpen]);

  const loadGroups = async () => {
    setGroupsLoading(true);
    setGroupsError(null);
    
    try {
      const accessToken = getAccesToken() || '';
      const groupsData = await getOwnGroups(accessToken);
      setGroups(groupsData);
    } catch (error) {
      console.error('Ошибка загрузки групп:', error);
      setGroupsError('Не удалось загрузить группы');
      setGroups([]);
    } finally {
      setGroupsLoading(false);
    }
  };

  const loadGenres = async () => {
    setGenresLoading(true);
    setGenresError(null);
    
    try {
      const accessToken = getAccesToken() || '';
      const genresData = await getGenres(accessToken);
      setGenres(genresData);
    } catch (error) {
      console.error('Ошибка загрузки жанров:', error);
      setGenresError('Не удалось загрузить жанры');
      setGenres([]);
    } finally {
      setGenresLoading(false);
    }
  };

  useEffect(() => {
    if (eventData) {
      setFormData(eventData);
      setDescription(eventData.description || '');
      setPublisher(eventData.publisher || '');
      setYear(eventData.year?.toString() || '');
      setAgeLimit(eventData.ageLimit || '');
      setSelectedGroup(eventData.groupId || null);
      
      const isEventOnline = eventData.location === 'online';
      setIsOnline(isEventOnline);
      
      if (eventData.location !== 'online' && eventData.location !== 'offline') {
        if (eventData.location?.includes('онлайн') || eventData.location?.includes('online')) {
          setIsOnline(true);
          setFormData(prev => ({ ...prev, location: 'online', adress: eventData.location || '' }));
        } else {
          setIsOnline(false);
          setFormData(prev => ({ ...prev, location: 'offline', adress: eventData.location || '' }));
        }
      }
    } else {
      setFormData({
        title: '',
        type: 'movies',
        date: '',
        duration: '',
        location: 'offline',
        adress: '',
        participants: 0,
        maxParticipants: 0,
        genres: [],
        image: ''
      });
      setDescription('');
      setPublisher('');
      setYear('');
      setAgeLimit('');
      setIsOnline(false);
      setSelectedGroup(null);
      setImageFile(null);
      setKinopoiskImageUrl(null);
    }
    setShowGenreSelector(false);
    setValidationErrors({});
  }, [eventData, isOpen]);

  if (!isOpen) return null;

  const validateForm = (): boolean => {
    const errors: ValidationErrors = {};

    if (!formData.title || formData.title.trim() === '') {
      errors.title = 'Название обязательно для заполнения';
    }

    if (!formData.genres || formData.genres.length === 0) {
      errors.genres = 'Необходимо выбрать хотя бы один жанр';
    }

    if (!formData.maxParticipants || formData.maxParticipants <= 0) {
      errors.maxParticipants = 'Укажите количество участников (больше 0)';
    }

    if (!selectedGroup) {
      errors.selectedGroup = 'Необходимо выбрать группу';
    }

    if (!imageFile && !kinopoiskImageUrl && !formData.image) {
      errors.image = 'Необходимо загрузить изображение';
    }

    if (!formData.date || formData.date === '') {
      errors.date = 'Необходимо указать дату и время события';
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleCategorySelect = (categoryId: string) => {
    setFormData(prev => ({ ...prev, type: categoryId as EventCardProps['type'] }));
  };

  const handleGenreToggle = (genre: string) => {
    setFormData(prev => ({
      ...prev,
      genres: prev.genres?.includes(genre) 
        ? prev.genres.filter(g => g !== genre)
        : [...(prev.genres || []), genre]
    }));
    
    if (validationErrors.genres) {
      setValidationErrors(prev => ({ ...prev, genres: undefined }));
    }
  };

  const handleSave = async () => {
    if (!validateForm()) {
      return;
    }

    if (mode === 'create') {
      await handleCreateEvent();
    } else {
      const eventToSave = {
        ...formData,
        description,
        publisher,
        year: year ? parseInt(year) : undefined,
        ageLimit,
        groupId: selectedGroup
      };
      onSave();
    }
  };

  const handleCreateEvent = async () => {
    setIsCreating(true);
    
    try {
      const accessToken = getAccesToken();
      if (!accessToken) {
        showNotification(401, 'Необходима авторизация');
        return;
      }

      const startTimeRFC3339 = convertToRFC3339(formData.date!);
      const sessionPlace = isOnline ? 1 : 0;
      const genresString = formData.genres!.join(',');
      const durationMinutes = parseDuration(formData.duration);
      
      const sessionTypeRu = convertCategEngToRu([formData.type!])[0] || 'Другое';
      
      let fieldsString = '';
      if (publisher) {
        fieldsString = `publisher:${publisher}`;
      }

      // Передаем либо imageFile, либо kinopoiskImageUrl
      const imageToSend = imageFile || kinopoiskImageUrl || undefined;

      await addEvent(
        accessToken,
        formData.title!,
        sessionTypeRu,
        sessionPlace,
        selectedGroup!,
        startTimeRFC3339,
        formData.maxParticipants!,
        durationMinutes,
        genresString,
        fieldsString || undefined,
        formData.adress || undefined,
        year ? parseInt(year) : undefined,
        undefined,
        ageLimit || undefined,
        description || undefined,
        imageToSend
      );

      showNotification(200, 'Событие успешно создано!');
      onSave();
      onClose();
    } catch (error: any) {
      console.error('Ошибка создания события:', error);
      
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Не удалось создать событие';
      
      showNotification(statusCode, errorMessage);
    } finally {
      setIsCreating(false);
    }
  };

  const handleKinopoiskSearch = async () => {
    if (!formData.title) {
      showNotification(400, 'Введите название фильма для поиска');
      return;
    }
    
    setIsLoading(true);
    try {
      const movies = await kinopoiskAPI.searchMovies(formData.title);
      
      if (movies.length > 0) {
        const movie = movies[0];
        
        console.log('Данные фильма из Кинопоиска:', movie);
        console.log('Длительность фильма:', movie.filmLength);
        
        // Обрезаем описание до 300 символов
        const movieDescription = movie.description || movie.shortDescription || '';
        const truncatedDescription = movieDescription.length > 300 
          ? movieDescription.substring(0, 300).trim() + '...'
          : movieDescription;
        
        setDescription(truncatedDescription);
        setPublisher(movie.countries?.[0]?.country || '');
        setYear(movie.year?.toString() || '');
        setAgeLimit(movie.ratingAgeLimits || '');
        
        if (movie.filmLength) {
          // filmLength приходит в формате "HH:MM", конвертируем в минуты
          const timeParts = movie.filmLength.split(':');
          let totalMinutes = 0;
          
          if (timeParts.length === 2) {
            const hours = parseInt(timeParts[0]) || 0;
            const minutes = parseInt(timeParts[1]) || 0;
            totalMinutes = hours * 60 + minutes;
          } else {
            // Если формат не "HH:MM", пробуем парсить как число
            totalMinutes = parseInt(movie.filmLength) || 0;
          }
          
          console.log('Длительность в минутах:', totalMinutes);
          setFormData(prev => ({ 
            ...prev, 
            duration: `${totalMinutes} мин` 
          }));
        }
        
        if (movie.genres) {
          const movieGenres = movie.genres.map(g => g.genre);
          
          // Сопоставляем жанры из Кинопоиска с жанрами из API
          const matchedGenres: string[] = [];
          const unmatchedGenres: string[] = [];
          
          movieGenres.forEach(movieGenre => {
            // Ищем точное совпадение
            const exactMatch = genres.find(apiGenre => 
              apiGenre.toLowerCase() === movieGenre.toLowerCase()
            );
            
            if (exactMatch) {
              matchedGenres.push(exactMatch);
            } else {
              // Ищем частичное совпадение
              const partialMatch = genres.find(apiGenre => 
                apiGenre.toLowerCase().includes(movieGenre.toLowerCase()) ||
                movieGenre.toLowerCase().includes(apiGenre.toLowerCase())
              );
              
              if (partialMatch) {
                matchedGenres.push(partialMatch);
              } else {
                // Если не найдено совпадений, добавляем как новый жанр
                unmatchedGenres.push(movieGenre);
              }
            }
          });
          
          // Объединяем найденные жанры и добавляем несовпавшие
          const finalGenres = [...matchedGenres, ...unmatchedGenres];
          
          // Также добавляем несовпавшие жанры в общий список для отображения
          if (unmatchedGenres.length > 0) {
            setGenres(prev => [...prev, ...unmatchedGenres]);
          }
          
          setFormData(prev => ({ 
            ...prev, 
            genres: finalGenres
          }));
          
          setValidationErrors(prev => ({ 
            ...prev, 
            genres: undefined
          }));
        }
        
        // Устанавливаем URL изображения для превью и отправки
        if (movie.posterUrl || movie.posterUrlPreview) {
          const imageUrl = movie.posterUrl || movie.posterUrlPreview;
          
          console.log('URL постера:', imageUrl);
          
          // Сохраняем URL для отображения превью
          setFormData(prev => ({ 
            ...prev, 
            image: imageUrl! 
          }));
          
          // Сохраняем URL для отправки на сервер
          setKinopoiskImageUrl(imageUrl!);
          
          // Убираем ошибку валидации изображения
          setValidationErrors(prev => ({ 
            ...prev, 
            image: undefined 
          }));
          
          // Очищаем imageFile на случай если было загружено ранее
          setImageFile(null);
        }
        
        const movieName = movie.nameRu || movie.nameEn || movie.nameOriginal;
        showNotification(200, `Данные фильма "${movieName}" успешно загружены`);
      } else {
        showNotification(404, 'Фильм не найден в Кинопоиске');
      }
    } catch (error) {
      console.error('Ошибка поиска:', error);
      showNotification(500, 'Ошибка при поиске фильма. Проверьте настройки API');
    } finally {
      setIsLoading(false);
    }
  };

  const handleMapSelect = () => {
    setShowMapModal(true);
  };

  const handlePlaceSelected = (place: PlaceInfo) => {
    setFormData(prev => ({ 
      ...prev, 
      adress: place.address
    }));
    setShowMapModal(false);
  };

  const handleImageUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setImageFile(file);
      setKinopoiskImageUrl(null); // Очищаем URL из Кинопоиска при загрузке нового файла
      
      const reader = new FileReader();
      reader.onload = (e) => {
        setFormData(prev => ({ ...prev, image: e.target?.result as string }));
        if (validationErrors.image) {
          setValidationErrors(prev => ({ ...prev, image: undefined }));
        }
      };
      reader.readAsDataURL(file);
    }
  };

  const handleOnlineToggle = () => {
    const newIsOnline = !isOnline;
    setIsOnline(newIsOnline);
    
    setFormData(prev => ({ 
      ...prev, 
      location: newIsOnline ? 'online' : 'offline',
      adress: ''
    }));
  };

  const handleAdressChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, adress: e.target.value }));
  };

  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, title: e.target.value }));
    if (validationErrors.title) {
      setValidationErrors(prev => ({ ...prev, title: undefined }));
    }
  };

  const handleMaxParticipantsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value) || 0;
    setFormData(prev => ({ ...prev, maxParticipants: value }));
    if (validationErrors.maxParticipants && value > 0) {
      setValidationErrors(prev => ({ ...prev, maxParticipants: undefined }));
    }
  };

  const handleGroupChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = Number(e.target.value) || null;
    setSelectedGroup(value);
    if (validationErrors.selectedGroup && value) {
      setValidationErrors(prev => ({ ...prev, selectedGroup: undefined }));
    }
  };

  const handleDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, date: e.target.value }));
    if (validationErrors.date) {
      setValidationErrors(prev => ({ ...prev, date: undefined }));
    }
  };

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      setShowGenreSelector(false);
    }
  };

  if (isCreating) {
    return (
      <div className={styles.modalOverlay}>
        <div className={styles.modal}>
          <LoadingIndicator text="Создаем событие..." />
        </div>
      </div>
    );
  }

  return (
    <div className={styles.modalOverlay} onClick={handleOverlayClick}>
        <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
          <button className={styles.closeButton} onClick={onClose}>
            <Image src="/icons/close.png" alt="Закрыть" width={24} height={24} />
          </button>
          
          <div className={styles.modalContent}>
            <div className={styles.leftColumn}>
              <div className={styles.fieldGroup}>
                <div className={styles.titleRow}>
                  <input
                    type="text"
                    className={`${styles.input} ${validationErrors.title ? 'error' : ''}`}
                    value={formData.title || ''}
                    onChange={handleTitleChange}
                    placeholder="Название *"
                  />
                  <button 
                    className={styles.kinopoiskButton} 
                    onClick={handleKinopoiskSearch}
                    disabled={isLoading}
                    title="Поиск в Кинопоиске"
                  >
                    {isLoading ? (
                      <span>...</span>
                    ) : (
                      <Image src="/social/kp.png" alt="Кинопоиск" width={24} height={24} />
                    )}
                  </button>
                </div>
                {validationErrors.title && (
                  <span className="errorMessage">{validationErrors.title}</span>
                )}
              </div>

              <div className={styles.fieldGroup}>
                <label className={styles.label}>Описание</label>
                <textarea
                  className={styles.textarea}
                  value={description}
                  onChange={(e) => {
                    const value = e.target.value;
                    if (value.length <= 300) {
                      setDescription(value);
                    } else {
                      setDescription(value.substring(0, 300));
                    }
                  }}
                  placeholder="Описание события..."
                  maxLength={300}
                />
                <div className={styles.charCounter}>
                  {description.length}/300
                </div>
              </div>

              <div className={styles.fieldGroup}>
                <label className={styles.label}>Выберите категорию:</label>
                <div className={styles.categories}>
                  {CATEGORIES.map((category) => (
                    <button
                      key={category.id}
                      type="button"
                      className={`${styles.categoryItem} ${
                        formData.type === category.id ? styles.selected : ''
                      }`}
                      onClick={() => handleCategorySelect(category.id)}
                    >
                      <Image src={category.icon} alt={category.id} width={24} height={24} />
                    </button>
                  ))}
                </div>
              </div>

              <div className={styles.fieldGroup}>
                <button
                  type="button"
                  className={`${styles.genreButton} ${validationErrors.genres ? 'error' : ''}`}
                  onClick={() => setShowGenreSelector(!showGenreSelector)}
                  disabled={genresLoading}
                >
                  {genresLoading 
                    ? 'Загрузка жанров...'
                    : formData.genres && formData.genres.length > 0 
                      ? `Выбрано жанров: ${formData.genres.length}` 
                      : 'Выберите жанры... *'
                  }
                </button>
                {validationErrors.genres && (
                  <span className="errorMessage">{validationErrors.genres}</span>
                )}
                {genresError && (
                  <div className={styles.errorMessage}>
                    {genresError}
                    <button 
                      onClick={loadGenres}
                      className={styles.retryButton}
                    >
                      Повторить
                    </button>
                  </div>
                )}
                {showGenreSelector && !genresLoading && genres.length > 0 && (
                  <div className={styles.genreSelector}>
                    {genres.map((genre) => (
                      <label key={genre} className={styles.genreOption}>
                        <input
                          type="checkbox"
                          checked={formData.genres?.includes(genre) || false}
                          onChange={() => handleGenreToggle(genre)}
                        />
                        {genre}
                      </label>
                    ))}
                  </div>
                )}
              </div>

              <div className={styles.fieldGroup}>
                <input
                  type="text"
                  className={styles.input}
                  value={publisher}
                  onChange={(e) => setPublisher(e.target.value)}
                  placeholder="Издатель"
                />
              </div>

              <div className={styles.fieldRow}>
                <div className={styles.fieldColumn}>
                  <div className={styles.fieldGroup}>
                    <input
                      type="number"
                      className={styles.input}
                      value={year}
                      onChange={(e) => setYear(e.target.value)}
                      placeholder="Год"
                      min="1900"
                      max="2030"
                    />
                  </div>
                  <div className={styles.fieldGroup}>
                    <input
                      type="text"
                      className={styles.input}
                      value={ageLimit}
                      onChange={(e) => setAgeLimit(e.target.value)}
                      placeholder="Ограничение"
                    />
                  </div>
                </div>

                <div className={styles.fieldColumnWithButton}>
                  <button
                    type="button"
                    className={`${styles.onlineButton} ${isOnline ? styles.online : styles.offline}`}
                    onClick={handleOnlineToggle}
                    title={isOnline ? "Онлайн событие" : "Офлайн событие"}
                  >
                    <Image 
                      src={isOnline ? "/events/online.png" : "/events/offline.png"} 
                      alt={isOnline ? "Онлайн" : "Офлайн"} 
                      width={30} 
                      height={30} 
                    />
                  </button>
                </div>
              </div>
            </div>

            <div className={styles.rightColumn}>
              <div className={styles.fieldGroup}>
                <select
                  className={`${styles.select} ${validationErrors.selectedGroup ? 'error' : ''}`}
                  value={selectedGroup || ''}
                  onChange={handleGroupChange}
                  disabled={groupsLoading}
                >
                  <option value="">
                    {groupsLoading ? 'Загрузка групп...' : 'Выберите группу... *'}
                  </option>
                  {groups.map((group) => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
                {validationErrors.selectedGroup && (
                  <span className="errorMessage">{validationErrors.selectedGroup}</span>
                )}
                {groupsError && (
                  <div className={styles.errorMessage}>
                    {groupsError}
                    <button 
                      onClick={loadGroups}
                      className={styles.retryButton}
                    >
                      Повторить
                    </button>
                  </div>
                )}
              </div>

              <div className={styles.fieldRow}>
                <div className={styles.fieldGroup}>
                  <input
                    type="datetime-local"
                    className={`${styles.input} ${validationErrors.date ? 'error' : ''}`}
                    value={formData.date || ''}
                    onChange={handleDateChange}
                    placeholder="Дата и время *"
                  />
                  {validationErrors.date && (
                    <span className="errorMessage">{validationErrors.date}</span>
                  )}
                </div>
                <div className={styles.fieldGroup}>
                  <input
                    type="text"
                    className={styles.input}
                    value={formData.duration || ''}
                    onChange={(e) => setFormData(prev => ({ ...prev, duration: e.target.value }))}
                    placeholder="Длительность (мин)"
                  />
                </div>
              </div>

              <div className={styles.fieldGroup}>
                <div className={styles.locationRow}>
                  <input
                    type="text"
                    className={styles.input}
                    value={formData.adress || ''}
                    onChange={handleAdressChange}
                    placeholder={isOnline ? "Ссылка на онлайн-событие" : "Место проведения"}
                  />
                  {!isOnline && (
                    <button 
                      type="button"
                      className={styles.mapButton} 
                      onClick={handleMapSelect}
                      title="Выбрать на карте"
                    >
                      <Image src="/events/offline.png" alt="Карта" width={16} height={16} />
                    </button>
                  )}
                </div>
              </div>

              <div className={styles.fieldGroup}>
                <input
                  type="number"
                  className={`${styles.input} ${validationErrors.maxParticipants ? 'error' : ''}`}
                  value={formData.maxParticipants || ''}
                  onChange={handleMaxParticipantsChange}
                  placeholder="Кол-во участников *"
                  min="1"
                />
                {validationErrors.maxParticipants && (
                  <span className="errorMessage">{validationErrors.maxParticipants}</span>
                )}
              </div>

              <div className={styles.imageUpload}>
                <input
                  type="file"
                  id="imageUpload"
                  accept="image/*"
                  onChange={handleImageUpload}
                  style={{ display: 'none' }}
                />
                <label 
                  htmlFor="imageUpload" 
                  className={`${styles.uploadLabel} ${validationErrors.image ? 'error' : ''}`}
                >
                  {formData.image ? (
                    <div className={styles.imagePreview}>
                      <Image 
                        src={formData.image} 
                        alt="Preview" 
                        width={360} 
                        height={360}
                        style={{ objectFit: 'cover' }}
                      />
                    </div>
                  ) : (
                    <>
                      <Image src="/default/load_img.png" alt="Загрузить" width={120} height={120} />
                      <span>Загрузить... *</span>
                    </>
                  )}
                </label>
                {validationErrors.image && (
                  <span className="errorMessage">{validationErrors.image}</span>
                )}
              </div>
            </div>
          </div>

          <div className={styles.modalFooter}>
            {onDelete && mode === 'edit' && (
              <button className={styles.deleteButton} onClick={onDelete}>
                Удалить событие
              </button>
            )}
            <button className={styles.saveButton} onClick={handleSave}>
              {mode === 'edit' ? 'Сохранить изменения' : 'Создать'}
            </button>
          </div>
        </div>

        <MapModal
            isOpen={showMapModal}
            onClose={() => setShowMapModal(false)}
            onPlaceSelected={handlePlaceSelected}
        />
      </div>
  );
}