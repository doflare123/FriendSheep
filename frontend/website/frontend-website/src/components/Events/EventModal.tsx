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
import { editEvent } from '@/api/events/editEvent';
import { delEvent } from '@/api/events/delEvent';
import { getImage } from '@/api/getImage';
import { getGenres } from '@/api/events/getGenres';
import { showNotification } from '@/utils';
import LoadingIndicator from '@/components/LoadingIndicator';
import { useRouter } from 'next/navigation';

interface EventModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: () => void;
  onDelete?: () => void;
  eventData?: Partial<EventCardProps>;
  mode?: 'create' | 'edit';
  groupId?: number;
}

const CATEGORIES = [
  { id: 'movies', icon: getCategoryIcon("movies") },
  { id: 'games', icon: getCategoryIcon("games") },
  { id: 'board', icon: getCategoryIcon("board") },
  { id: 'other', icon: getCategoryIcon("other") }
];

// Константы для валидации
const VALIDATION_RULES = {
  title: { min: 5, max: 40 },
  description: { min: 5, max: 300 }
};

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
  description?: string;
  genres?: string;
  maxParticipants?: string;
  selectedGroup?: string;
  image?: string;
  date?: string;
  location?: string;
}

export default function EventModal({ 
  isOpen, 
  onClose, 
  onSave, 
  onDelete, 
  eventData, 
  mode = 'create',
  groupId 
}: EventModalProps) {
  const [formData, setFormData] = useState<Partial<EventCardProps>>({
    title: '',
    type: 'movies',
    date: '',
    start_time: '',
    duration: '',
    location: 'offline',
    adress: '',
    participants: 0,
    maxParticipants: 0,
    genres: [],
    image: ''
  });

  const [originalData, setOriginalData] = useState<Partial<EventCardProps>>({});
  const router = useRouter();

  const [isOnline, setIsOnline] = useState(false);
  const [showGenreSelector, setShowGenreSelector] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState<number | null>(null);
  const [publisher, setPublisher] = useState('');
  const [year, setYear] = useState('');
  const [ageLimit, setAgeLimit] = useState('');
  const [description, setDescription] = useState('');
  const [originalDescription, setOriginalDescription] = useState('');
  const [city, setCity] = useState('');
  const [originalCity, setOriginalCity] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showMapModal, setShowMapModal] = useState(false);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [kinopoiskImageUrl, setKinopoiskImageUrl] = useState<string | null>(null);
  
  const [groups, setGroups] = useState<ApiGroup[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});
  const [isSaving, setIsSaving] = useState(false);
  
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
      const accessToken = getAccesToken(router) || '';
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
      const accessToken = getAccesToken(router) || '';
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

  const convertToDatetimeLocal = (date?: string, startTime?: string): string => {
    if (!date) return '';
    
    try {
      const [day, month, year] = date.split('.');
      
      if (!day || !month || !year) {
        console.error('Неверный формат даты:', date);
        return '';
      }
      
      if (startTime && startTime.includes(':')) {
        const [hours, minutes] = startTime.split(':');
        return `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}T${hours.padStart(2, '0')}:${minutes.padStart(2, '0')}`;
      }
      
      return `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}T00:00`;
      
    } catch (error) {
      console.error('Ошибка конвертации даты/времени:', error);
      return '';
    }
  };

  useEffect(() => {
    if (eventData) {
      console.log('EventData получен:', eventData);
      
      const datetimeLocal = convertToDatetimeLocal(eventData.date, eventData.start_time);
      
      const initialFormData = {
        ...eventData,
        date: datetimeLocal
      };
      
      setFormData(initialFormData);
      setOriginalData(initialFormData);
      
      const eventDescription = eventData.description || '';
      setDescription(eventDescription);
      setOriginalDescription(eventDescription);
      
      setPublisher(eventData.publisher || '');
      setYear(eventData.year?.toString() || '');
      setAgeLimit(eventData.ageLimit || '');
      
      setSelectedGroup(groupId || null);
      console.log('selectedGroup установлен из props groupId:', groupId);
      
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
      
      if (eventData.fields) {
        const cityMatch = eventData.fields.match(/city:([^,]+)/);
        if (cityMatch) {
          setCity(cityMatch[1]);
          setOriginalCity(cityMatch[1]);
        }
      }
    } else {
      const defaultData = {
        title: '',
        type: 'movies' as const,
        date: '',
        start_time: '',
        duration: '',
        location: 'offline' as const,
        adress: '',
        participants: 0,
        maxParticipants: 0,
        genres: [],
        image: ''
      };
      setFormData(defaultData);
      setOriginalData(defaultData);
      setDescription('');
      setOriginalDescription('');
      setPublisher('');
      setYear('');
      setAgeLimit('');
      setCity('');
      setOriginalCity('');
      setIsOnline(false);
      
      setSelectedGroup(groupId || null);
      console.log('selectedGroup установлен из props groupId при создании:', groupId);
      
      setImageFile(null);
      setKinopoiskImageUrl(null);
    }
    setShowGenreSelector(false);
    setValidationErrors({});
  }, [eventData, isOpen, groupId]);

  if (!isOpen) return null;

  const validateForm = (): boolean => {
    const errors: ValidationErrors = {};

    // Валидация названия
    if (!formData.title || formData.title.trim() === '') {
      errors.title = 'Название обязательно для заполнения';
    } else if (formData.title.trim().length < VALIDATION_RULES.title.min) {
      errors.title = `Название должно содержать минимум ${VALIDATION_RULES.title.min} символов`;
    } else if (formData.title.trim().length > VALIDATION_RULES.title.max) {
      errors.title = `Название не должно превышать ${VALIDATION_RULES.title.max} символов`;
    }

    // Валидация описания (если заполнено)
    if (description && description.trim() !== '') {
      if (description.trim().length < VALIDATION_RULES.description.min) {
        errors.description = `Описание должно содержать минимум ${VALIDATION_RULES.description.min} символов`;
      } else if (description.trim().length > VALIDATION_RULES.description.max) {
        errors.description = `Описание не должно превышать ${VALIDATION_RULES.description.max} символов`;
      }
    }

    if (mode === 'create') {
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

      if (isOnline) {
        if (!formData.adress || formData.adress.trim() === '') {
          errors.location = 'Необходимо указать ссылку на онлайн-событие';
        }
      } else {
        if (!formData.adress || formData.adress.trim() === '' || !city) {
          errors.location = 'Необходимо выбрать место на карте';
        }
      }
    } else if (mode === 'edit') {
      const hasChanges = 
        formData.title !== originalData.title ||
        formData.maxParticipants !== originalData.maxParticipants ||
        formData.duration !== originalData.duration ||
        formData.date !== originalData.date ||
        formData.type !== originalData.type ||
        JSON.stringify(formData.genres) !== JSON.stringify(originalData.genres) ||
        year !== (originalData.year?.toString() || '') ||
        ageLimit !== (originalData.ageLimit || '') ||
        formData.location !== originalData.location ||
        formData.adress !== originalData.adress ||
        city !== originalCity ||
        description !== originalDescription ||
        imageFile !== null ||
        kinopoiskImageUrl !== null;

      if (!hasChanges && formData.image === originalData.image) {
        errors.title = 'Необходимо изменить хотя бы одно поле';
      }

      if (formData.maxParticipants && formData.maxParticipants <= 0) {
        errors.maxParticipants = 'Количество участников должно быть больше 0';
      }

      if (formData.date && formData.date === '') {
        errors.date = 'Необходимо указать дату и время события';
      }
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
      await handleEditEvent();
    }
  };

  const handleDelete = async () => {
    if (!eventData?.id) {
      showNotification(400, 'ID события не найден');
      return;
    }

    const confirmed = window.confirm('Вы уверены, что хотите удалить это событие?');
    if (!confirmed) return;

    setIsSaving(true);
    
    try {
      const accessToken = getAccesToken(router);
      if (!accessToken) {
        showNotification(401, 'Необходима авторизация');
        return;
      }

      await delEvent(accessToken, eventData.id);

      showNotification(200, 'Событие успешно удалено!');
      
      onClose();
      window.location.reload();
    } catch (error: any) {
      console.error('Ошибка удаления события:', error);
      
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Не удалось удалить событие';
      
      showNotification(statusCode, errorMessage);
    } finally {
      setIsSaving(false);
    }
  };

  const buildFieldsString = (): string | undefined => {
    const fields: string[] = [];
    
    if (publisher) {
      fields.push(`publisher:${publisher}`);
    }
    
    if (!isOnline && city) {
      fields.push(`city:${city}`);
    }
    
    return fields.length > 0 ? fields.join(',') : undefined;
  };

  const handleCreateEvent = async () => {
    setIsSaving(true);
    
    try {
      const accessToken = getAccesToken(router);
      if (!accessToken) {
        showNotification(401, 'Необходима авторизация');
        return;
      }

      const startTimeRFC3339 = convertToRFC3339(formData.date!);
      const sessionPlace = isOnline ? '1' : '2';
      const genresString = formData.genres!.join(',');
      const durationMinutes = parseDuration(formData.duration);
      
      const sessionTypeRu = convertCategEngToRu([formData.type!])[0] || 'Другое';
      
      const fieldsString = buildFieldsString();

      // ИЗМЕНЕНИЕ: Определяем URL изображения для передачи
      let imageToSend: string | undefined;
      
      if (imageFile) {
        // Если загружен локальный файл - сначала загружаем на сервер и получаем URL
        imageToSend = await getImage(accessToken, imageFile);
      } else if (kinopoiskImageUrl) {
        // Если получена ссылка с Кинопоиска - передаем URL напрямую
        imageToSend = kinopoiskImageUrl;
      } else if (formData.image) {
        // Если есть существующее изображение
        imageToSend = formData.image;
      }

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
        fieldsString,
        formData.adress || undefined,
        year ? parseInt(year) : undefined,
        undefined,
        ageLimit || undefined,
        description || undefined,
        imageToSend // Передаем URL строку
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
      setIsSaving(false);
    }
  };

  const handleEditEvent = async () => {
    setIsSaving(true);
    
    try {
      const accessToken = getAccesToken(router);
      if (!accessToken) {
        showNotification(401, 'Необходима авторизация');
        return;
      }

      if (!eventData?.id) {
        showNotification(400, 'ID события не найден');
        return;
      }

      // ИЗМЕНЕНИЕ: Определяем URL изображения
      let imageUrl: string | undefined = undefined;
      
      if (imageFile) {
        // Загружаем локальный файл и получаем URL
        imageUrl = await getImage(accessToken, imageFile);
      } else if (kinopoiskImageUrl && kinopoiskImageUrl !== originalData.image) {
        // Используем URL с Кинопоиска напрямую
        imageUrl = kinopoiskImageUrl;
      }
      // Если ничего не изменилось - imageUrl останется undefined

      const updatedTitle = formData.title !== originalData.title ? formData.title : undefined;
      const updatedMaxParticipants = formData.maxParticipants !== originalData.maxParticipants 
        ? formData.maxParticipants 
        : undefined;
      const updatedDuration = formData.duration !== originalData.duration 
        ? (formData.duration ? parseInt(formData.duration) : undefined)
        : undefined;
      const updatedStartTime = formData.date !== originalData.date 
        ? convertToRFC3339(formData.date!)
        : undefined;
      
      const updatedYear = year !== (originalData.year?.toString() || '') 
        ? (year ? parseInt(year) : undefined)
        : undefined;
      const updatedAgeLimit = ageLimit !== (originalData.ageLimit || '') 
        ? (ageLimit || undefined)
        : undefined;
      const updatedGenres = JSON.stringify(formData.genres) !== JSON.stringify(originalData.genres)
        ? formData.genres
        : undefined;
      const updatedLocation = formData.adress !== originalData.adress 
        ? formData.adress 
        : undefined;
      
      let updatedSessionPlaceId: string | undefined = undefined;
      if (formData.location !== originalData.location) {
        updatedSessionPlaceId = formData.location === 'online' ? '1' : '2';
      }
      
      let updatedSessionTypeId: string | undefined = undefined;
      if (formData.type !== originalData.type) {
        const sessionTypeRu = convertCategEngToRu([formData.type!])[0] || 'Другое';
        updatedSessionTypeId = sessionTypeRu;
      }

      const finalYear = updatedYear !== undefined ? updatedYear : (eventData.year || 2025);

      await editEvent(
        accessToken,
        eventData.id,
        finalYear,
        updatedMaxParticipants,
        updatedDuration,
        imageUrl, // Передаем URL строку (или undefined)
        updatedStartTime,
        updatedTitle,
        updatedAgeLimit,
        updatedGenres,
        updatedLocation,
        updatedSessionPlaceId,
        updatedSessionTypeId
      );

      showNotification(200, 'Событие успешно отредактировано!');
      onSave();
      onClose();
    } catch (error: any) {
      console.error('Ошибка редактирования события:', error);
      
      const statusCode = error.response?.status || 500;
      const errorMessage = error.response?.data?.message || 'Не удалось отредактировать событие';
      
      showNotification(statusCode, errorMessage);
    } finally {
      setIsSaving(false);
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
        
        const movieDescription = movie.description || movie.shortDescription || '';
        const truncatedDescription = movieDescription.length > VALIDATION_RULES.description.max 
          ? movieDescription.substring(0, VALIDATION_RULES.description.max - 3).trim() + '...'
          : movieDescription;
        
        setDescription(truncatedDescription);
        setPublisher(movie.countries?.[0]?.country || '');
        setYear(movie.year?.toString() || '');
        setAgeLimit(movie.ratingAgeLimits || '');
        
        if (movie.filmLength) {
          const timeParts = movie.filmLength.split(':');
          let totalMinutes = 0;
          
          if (timeParts.length === 2) {
            const hours = parseInt(timeParts[0]) || 0;
            const minutes = parseInt(timeParts[1]) || 0;
            totalMinutes = hours * 60 + minutes;
          } else {
            totalMinutes = parseInt(movie.filmLength) || 0;
          }
          
          console.log('Длительность в минутах:', totalMinutes);
          setFormData(prev => ({ 
            ...prev, 
            duration: `${totalMinutes}`
          }));
        }
        
        if (movie.genres) {
          const movieGenres = movie.genres.map(g => g.genre);
          
          const matchedGenres: string[] = [];
          const unmatchedGenres: string[] = [];
          
          movieGenres.forEach(movieGenre => {
            const exactMatch = genres.find(apiGenre => 
              apiGenre.toLowerCase() === movieGenre.toLowerCase()
            );
            
            if (exactMatch) {
              matchedGenres.push(exactMatch);
            } else {
              const partialMatch = genres.find(apiGenre => 
                apiGenre.toLowerCase().includes(movieGenre.toLowerCase()) ||
                movieGenre.toLowerCase().includes(apiGenre.toLowerCase())
              );
              
              if (partialMatch) {
                matchedGenres.push(partialMatch);
              } else {
                unmatchedGenres.push(movieGenre);
              }
            }
          });
          
          const finalGenres = [...matchedGenres, ...unmatchedGenres];
          
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
        
        if (movie.posterUrl || movie.posterUrlPreview) {
          const imageUrl = movie.posterUrl || movie.posterUrlPreview;
          
          console.log('URL постера:', imageUrl);
          
          setFormData(prev => ({ 
            ...prev, 
            image: imageUrl! 
          }));
          
          setKinopoiskImageUrl(imageUrl!);
          
          setValidationErrors(prev => ({ 
            ...prev, 
            image: undefined 
          }));
          
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
    console.log('Выбрано место:', place);
    
    setFormData(prev => ({ 
      ...prev, 
      adress: place.address
    }));
    
    if (place.city) {
      setCity(place.city);
    } else {
      const addressParts = place.address.split(',');
      if (addressParts.length > 0) {
        const cityPart = addressParts[0].trim();
        setCity(cityPart);
      }
    }
    
    if (validationErrors.location) {
      setValidationErrors(prev => ({ ...prev, location: undefined }));
    }
    
    setShowMapModal(false);
  };

  const handleImageUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setImageFile(file);
      setKinopoiskImageUrl(null);
      
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
    
    if (newIsOnline) {
      setCity('');
    }
    
    if (validationErrors.location) {
      setValidationErrors(prev => ({ ...prev, location: undefined }));
    }
  };

  const handleOnlineLinkChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, adress: e.target.value }));
    if (validationErrors.location) {
      setValidationErrors(prev => ({ ...prev, location: undefined }));
    }
  };

  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    const limitedValue = value.slice(0, VALIDATION_RULES.title.max);
    setFormData(prev => ({ ...prev, title: limitedValue }));
    if (validationErrors.title) {
      setValidationErrors(prev => ({ ...prev, title: undefined }));
    }
  };

  const handleDescriptionChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    const limitedValue = value.slice(0, VALIDATION_RULES.description.max);
    setDescription(limitedValue);
    if (validationErrors.description) {
      setValidationErrors(prev => ({ ...prev, description: undefined }));
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

  const isFieldDisabled = false;

  if (isSaving) {
    return (
      <div className={styles.modalOverlay}>
        <div className={styles.modal}>
          <LoadingIndicator text={mode === 'create' ? "Создаем событие..." : "Сохраняем изменения..."} />
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
                    maxLength={VALIDATION_RULES.title.max}
                  />
                  {mode === 'create' && (
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
                  )}
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  {validationErrors.title && (
                    <span className="errorMessage">{validationErrors.title}</span>
                  )}
                  <span style={{ 
                    fontSize: '12px', 
                    color: '#999',
                    marginLeft: 'auto',
                    marginTop: '4px'
                  }}>
                    {(formData.title || '').length}/{VALIDATION_RULES.title.max}
                  </span>
                </div>
              </div>

              <div className={styles.fieldGroup}>
                <label className={styles.label}>Описание</label>
                <textarea
                  className={`${styles.textarea} ${validationErrors.description ? 'error' : ''}`}
                  value={description}
                  onChange={handleDescriptionChange}
                  placeholder="Описание события..."
                  maxLength={VALIDATION_RULES.description.max}
                  disabled={isFieldDisabled}
                />
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '4px' }}>
                  {validationErrors.description && (
                    <span className="errorMessage">{validationErrors.description}</span>
                  )}
                  <span style={{ 
                    fontSize: '12px', 
                    color: '#999',
                    marginLeft: 'auto'
                  }}>
                    {description.length}/{VALIDATION_RULES.description.max}
                  </span>
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
                      disabled={isFieldDisabled}
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
                  disabled={genresLoading || isFieldDisabled}
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
                          disabled={isFieldDisabled}
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
                  disabled={isFieldDisabled}
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
                      disabled={isFieldDisabled}
                    />
                  </div>
                  <div className={styles.fieldGroup}>
                    <input
                      type="text"
                      className={styles.input}
                      value={ageLimit}
                      onChange={(e) => setAgeLimit(e.target.value)}
                      placeholder="Ограничение"
                      disabled={isFieldDisabled}
                    />
                  </div>
                </div>

                <div className={styles.fieldColumnWithButton}>
                  <button
                    type="button"
                    className={`${styles.onlineButton} ${isOnline ? styles.online : styles.offline}`}
                    onClick={handleOnlineToggle}
                    title={isOnline ? "Онлайн событие" : "Офлайн событие"}
                    disabled={isFieldDisabled}
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
                  disabled={groupsLoading || isFieldDisabled || mode === 'edit'}
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
                {isOnline ? (
                  <>
                    <input
                      type="text"
                      className={`${styles.input} ${validationErrors.location ? 'error' : ''}`}
                      value={formData.adress || ''}
                      onChange={handleOnlineLinkChange}
                      placeholder="Ссылка на онлайн-событие *"
                      disabled={isFieldDisabled}
                    />
                    {validationErrors.location && (
                      <span className="errorMessage">{validationErrors.location}</span>
                    )}
                  </>
                ) : (
                  <>
                    <div className={styles.locationRow}>
                      <input
                        type="text"
                        className={`${styles.input} ${validationErrors.location ? 'error' : ''}`}
                        value={formData.adress || ''}
                        placeholder="Место проведения *"
                        disabled={true}
                        readOnly
                      />
                      <button 
                        type="button"
                        className={styles.mapButton} 
                        onClick={handleMapSelect}
                        title="Выбрать на карте"
                        disabled={isFieldDisabled}
                      >
                        <Image src="/events/offline.png" alt="Карта" width={16} height={16} />
                      </button>
                    </div>
                    {validationErrors.location && (
                      <span className="errorMessage">{validationErrors.location}</span>
                    )}
                  </>
                )}
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
              <button className={styles.deleteButton} onClick={handleDelete}>
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