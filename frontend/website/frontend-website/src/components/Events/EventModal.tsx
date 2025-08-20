import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import styles from '../../styles/Events/EventModal.module.css';
import MapModal from './MapModal';
import { EventCardProps } from '../../types/Events';
import { GroupData } from '../../types/Groups';
import { kinopoiskAPI, PlaceInfo } from '../../lib/api';
import {getCategoryIcon, getAccesToken} from '../../Constants'
import {getOwnGroups} from '../../api/get_owngroups'

interface EventModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (eventData: Partial<EventCardProps>) => void;
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

const GENRES = [
  'Экшн', 'Комедия', 'Драма', 'Ужасы', 'Фантастика', 'Триллер',
  'Романтика', 'Приключения', 'Криминал', 'Детектив', 'Военный',
  'Биография', 'История', 'Мюзикл', 'Семейный', 'Анимация'
];

// Интерфейс для групп из API
interface ApiGroup {
  category: string[];
  id: number;
  image: string;
  member_count: number;
  name: string;
  small_description: string;
  type: string;
}

// Интерфейс для ошибок валидации
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
  
  // Состояния для групп
  const [groups, setGroups] = useState<ApiGroup[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsError, setGroupsError] = useState<string | null>(null);

  // Состояние для ошибок валидации
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});

  // Загрузка групп при открытии модала
  useEffect(() => {
    if (isOpen) {
      loadGroups();
    }
  }, [isOpen]);

  const loadGroups = async () => {
    setGroupsLoading(true);
    setGroupsError(null);
    
    try {
      const accessToken = getAccesToken() || ''; // Временное решение
      const groupsData = await getOwnGroups(accessToken);
      setGroups(groupsData);
    } catch (error) {
      console.error('Ошибка загрузки групп:', error);
      setGroupsError('Не удалось загрузить группы');
      setGroups([]); // Устанавливаем пустой массив при ошибке
    } finally {
      setGroupsLoading(false);
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
      
      // Определяем режим онлайн/офлайн по location
      const isEventOnline = eventData.location === 'online';
      setIsOnline(isEventOnline);
      
      // Если данные не содержат location в новом формате, определяем по старому
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
      // Сброс всех полей при создании нового события
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
    }
    setShowGenreSelector(false);
    setValidationErrors({}); // Сбрасываем ошибки валидации
  }, [eventData, isOpen]);

  if (!isOpen) return null;

  // Функция валидации обязательных полей
  const validateForm = (): boolean => {
    const errors: ValidationErrors = {};

    // Проверка названия
    if (!formData.title || formData.title.trim() === '') {
      errors.title = 'Название обязательно для заполнения';
    }

    // Проверка жанров
    if (!formData.genres || formData.genres.length === 0) {
      errors.genres = 'Необходимо выбрать хотя бы один жанр';
    }

    // Проверка количества участников
    if (!formData.maxParticipants || formData.maxParticipants <= 0) {
      errors.maxParticipants = 'Укажите количество участников (больше 0)';
    }

    // Проверка группы
    if (!selectedGroup) {
      errors.selectedGroup = 'Необходимо выбрать группу';
    }

    // Проверка картинки
    if (!formData.image || formData.image === '') {
      errors.image = 'Необходимо загрузить изображение';
    }

    // Проверка даты и времени
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
    
    // Очищаем ошибку жанров при выборе
    if (validationErrors.genres) {
      setValidationErrors(prev => ({ ...prev, genres: undefined }));
    }
  };

  const handleSave = () => {
    // Проверяем валидность формы
    if (!validateForm()) {
      return;
    }

    const eventToSave = {
      ...formData,
      description,
      publisher,
      year: year ? parseInt(year) : undefined,
      ageLimit,
      groupId: selectedGroup
    };
    onSave(eventToSave);
  };

  const handleKinopoiskSearch = async () => {
    if (!formData.title) {
      alert('Введите название фильма для поиска');
      return;
    }
    
    setIsLoading(true);
    try {
      const movies = await kinopoiskAPI.searchMovies(formData.title);
      
      if (movies.length > 0) {
        const movie = movies[0]; // Берем первый результат
        
        // Заполняем поля данными из Кинопоиска
        setDescription(movie.description || movie.shortDescription || '');
        setPublisher(movie.countries?.[0]?.country || '');
        setYear(movie.year?.toString() || '');
        setAgeLimit(movie.ratingAgeLimits || '');
        
        if (movie.filmLength) {
          setFormData(prev => ({ 
            ...prev, 
            duration: `${movie.filmLength} мин` 
          }));
        }
        
        if (movie.genres) {
          const movieGenres = movie.genres.map(g => g.genre);
          setFormData(prev => ({ 
            ...prev, 
            genres: movieGenres,
            image: movie.posterUrl || movie.posterUrlPreview || prev.image
          }));
          
          // Очищаем ошибки жанров и изображения если данные загрузились
          setValidationErrors(prev => ({ 
            ...prev, 
            genres: undefined,
            image: movie.posterUrl || movie.posterUrlPreview ? undefined : prev.image
          }));
        }
        
        alert(`Данные фильма "${movie.nameRu || movie.nameEn || movie.nameOriginal}" загружены!`);
      } else {
        alert('Фильм не найден');
      }
    } catch (error) {
      console.error('Ошибка поиска:', error);
      alert('Ошибка при поиске фильма. Проверьте настройки API.');
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
      const reader = new FileReader();
      reader.onload = (e) => {
        setFormData(prev => ({ ...prev, image: e.target?.result as string }));
        // Очищаем ошибку изображения
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
    // Очищаем ошибку названия при вводе
    if (validationErrors.title) {
      setValidationErrors(prev => ({ ...prev, title: undefined }));
    }
  };

  const handleMaxParticipantsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value) || 0;
    setFormData(prev => ({ ...prev, maxParticipants: value }));
    // Очищаем ошибку количества участников при вводе
    if (validationErrors.maxParticipants && value > 0) {
      setValidationErrors(prev => ({ ...prev, maxParticipants: undefined }));
    }
  };

  const handleGroupChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = Number(e.target.value) || null;
    setSelectedGroup(value);
    // Очищаем ошибку группы при выборе
    if (validationErrors.selectedGroup && value) {
      setValidationErrors(prev => ({ ...prev, selectedGroup: undefined }));
    }
  };

  const handleDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, date: e.target.value }));
    // Очищаем ошибку даты при вводе
    if (validationErrors.date) {
      setValidationErrors(prev => ({ ...prev, date: undefined }));
    }
  };

  // Закрытие селектора жанров при клике вне его
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      setShowGenreSelector(false);
    }
  };

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
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Описание события..."
                />
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
                >
                  {formData.genres && formData.genres.length > 0 
                    ? `Выбрано жанров: ${formData.genres.length}` 
                    : 'Выберите жанры... *'
                  }
                </button>
                {validationErrors.genres && (
                  <span className="errorMessage">{validationErrors.genres}</span>
                )}
                {showGenreSelector && (
                  <div className={styles.genreSelector}>
                    {GENRES.map((genre) => (
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
                    placeholder="Длительность"
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

        {/* Модальное окно карты как отдельный компонент */}
        <MapModal
            isOpen={showMapModal}
            onClose={() => setShowMapModal(false)}
            onPlaceSelected={handlePlaceSelected}
        />
      </div>
  );
}