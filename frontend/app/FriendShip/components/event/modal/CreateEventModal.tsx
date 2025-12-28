import kinopoiskService from '@/api/services/kinopoisk/kinopoiskService';
import permissionsService from '@/api/services/permissionsService';
import rawgService from '@/api/services/rawg/rawgService';
import ConfirmationModal from '@/components/ConfirmationModal';
import { useTheme } from '@/components/ThemeContext';
import { Event } from '@/components/event/EventCard';
import KinopoiskButton from '@/components/event/KinopoiskButton';
import RawgButton from '@/components/event/RawgButton';
import CategorySelector from '@/components/event/modal/CategorySelector';
import EventTypeSelector from '@/components/event/modal/EventTypeSelector';
import GenreSelector from '@/components/event/modal/GenreSelector';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import profanityFilter from '@/utils/profanityFilter';
import { validateFullDescription } from '@/utils/validators';
import * as ImagePicker from 'expo-image-picker';
import React, { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  BackHandler,
  Dimensions,
  Image,
  ImageBackground,
  Linking,
  Modal,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View
} from 'react-native';
import DatePicker from 'react-native-date-picker';
import DescriptionModal from './DescriptionModal';
import YandexMapModal from './YandexMapModal';

const screenHeight = Dimensions.get("window").height;

interface CreateEditEventModalProps {
  visible: boolean;
  onClose: () => void;
  onCreate?: (eventData: any) => void;
  onDelete?: (eventId: string) => void;
  onUpdate?: (eventId: string, eventData: any) => void;
  groupName?: string;
  editMode?: boolean;
  initialData?: Event;
  availableGenres?: string[];
  isLoading?: boolean;
}

const CreateEditEventModal: React.FC<CreateEditEventModalProps> = ({ 
  visible, 
  onClose, 
  onCreate,
  onDelete,
  onUpdate,
  groupName = 'Мега крутая группа',
  editMode = false,
  initialData,
  availableGenres = [],
  isLoading = false,
}) => {
  const colors = useThemedColors();
  const { isDark } = useTheme();
  const [eventName, setEventName] = useState('');
  const [description, setDescription] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [eventType, setEventType] = useState<'online' | 'offline'>('online');
  const [selectedGenres, setSelectedGenres] = useState<string[]>([]);
  const [publisher, setPublisher] = useState('');
  const [publishYear, setPublishYear] = useState('');
  const [country, setCountry] = useState('');
  const [ageRating, setAgeRating] = useState('');
  const [eventDate, setEventDate] = useState<Date>(new Date());
  const [duration, setDuration] = useState('');
  const [eventPlace, setEventPlace] = useState('');
  const [maxParticipants, setMaxParticipants] = useState('');
  const [eventImage, setEventImage] = useState<string>('');
  const [imageFile, setImageFile] = useState<any>(null);

  const [datePickerVisible, setDatePickerVisible] = useState(false);

  const [kinopoiskModalVisible, setKinopoiskModalVisible] = useState(false);
  const [isLoadingKinopoisk, setIsLoadingKinopoisk] = useState(false);

  const [rawgModalVisible, setRawgModalVisible] = useState(false);
  const [isLoadingRawg, setIsLoadingRawg] = useState(false);

  const [mapModalVisible, setMapModalVisible] = useState(false);

  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const [descriptionModalVisible, setDescriptionModalVisible] = useState(false);

   useEffect(() => {
    if (!visible) return;

    const backHandler = BackHandler.addEventListener('hardwareBackPress', () => {
      handleClose();
      return true;
    });

    return () => backHandler.remove();
  }, [visible]);

  useEffect(() => {
    if (editMode && initialData && visible) {
      console.log('[CreateEditEventModal] 🔧 Загрузка данных для редактирования');
      console.log('[CreateEditEventModal] 📦 initialData:', JSON.stringify(initialData, null, 2));

      setEventName(initialData.title);
      console.log('[CreateEditEventModal] ✅ Название:', initialData.title);

      setDescription(initialData.description || '');
      console.log('[CreateEditEventModal] ✅ Описание:', initialData.description || '(пусто)');

      setSelectedCategory(initialData.category);
      console.log('[CreateEditEventModal] ✅ Категория:', initialData.category);

      setEventType(initialData.typePlace);
      console.log('[CreateEditEventModal] ✅ Тип:', initialData.typePlace);

      setSelectedGenres(initialData.genres || []);
      console.log('[CreateEditEventModal] ✅ Жанры:', initialData.genres);

      setPublisher(initialData.publisher || '');
      console.log('[CreateEditEventModal] ✅ Издатель:', initialData.publisher || '(пусто)');
 
      if (initialData.publicationDate) {
        const yearStr = initialData.publicationDate.toString();
        if (yearStr.length === 4 && !isNaN(parseInt(yearStr))) {
          setPublishYear(yearStr);
          console.log('[CreateEditEventModal] ✅ Год (строка):', yearStr);
        } else {
          try {
            const year = new Date(initialData.publicationDate).getFullYear();
            if (!isNaN(year) && year > 1900) {
              setPublishYear(year.toString());
              console.log('[CreateEditEventModal] ✅ Год (дата):', year);
            }
          } catch (e) {
            console.log('[CreateEditEventModal] ⚠️ Не удалось распарсить год:', initialData.publicationDate);
          }
        }
      }

      setAgeRating(initialData.ageRating || '');
      console.log('[CreateEditEventModal] ✅ Ограничение:', initialData.ageRating || '(пусто)');

      const durationMatch = initialData.duration.match(/\d+/);
      if (durationMatch) {
        setDuration(durationMatch[0]);
        console.log('[CreateEditEventModal] ✅ Длительность:', durationMatch[0]);
      }

      setEventPlace(initialData.eventPlace || '');
      console.log('[CreateEditEventModal] ✅ Место:', initialData.eventPlace || '(пусто)');

      setMaxParticipants(initialData.maxParticipants.toString());
      console.log('[CreateEditEventModal] ✅ Участников:', initialData.maxParticipants);

      setEventImage(initialData.imageUri);
      console.log('[CreateEditEventModal] ✅ Изображение установлено');

      try {
        const dateParts = initialData.date.split(' ');
        if (dateParts.length >= 1) {
          const [day, month, year] = dateParts[0].split('.');
          const hours = dateParts.length > 1 ? dateParts[1].split(':')[0] : '0';
          const minutes = dateParts.length > 1 ? dateParts[1].split(':')[1] : '0';
          
          const parsedDate = new Date(
            parseInt(year),
            parseInt(month) - 1,
            parseInt(day),
            parseInt(hours || '0'),
            parseInt(minutes || '0')
          );
          
          setEventDate(parsedDate);
          console.log('[CreateEditEventModal] ✅ Дата:', parsedDate);
        }
      } catch (e) {
        console.error('[CreateEditEventModal] ❌ Ошибка парсинга даты:', e);
      }
      
      console.log('[CreateEditEventModal] ✅ ВСЕ ДАННЫЕ ЗАГРУЖЕНЫ');
    } else if (!editMode && visible) {
      resetForm();
    }
  }, [editMode, initialData, visible]);

  const categories = [
    { id: 'movie', label: 'Медиа', icon: require('@/assets/images/event_card/movie.png') },
    { id: 'game', label: 'Видеоигры', icon: require('@/assets/images/event_card/game.png') },
    { id: 'table_game', label: 'Настолки', icon: require('@/assets/images/event_card/table_game.png') },
    { id: 'other', label: 'Другое', icon: require('@/assets/images/event_card/other.png') },
  ];

  const showPermissionAlert = (onRetry: () => void) => {
    Alert.alert(
      'Разрешение на доступ к фото',
      'Для загрузки изображений необходимо разрешить доступ к галерее. Хотите предоставить разрешение?',
      [
        {
          text: 'Отмена',
          style: 'cancel',
        },
        {
          text: 'Настройки',
          onPress: () => {
            Linking.openSettings();
          },
        },
        {
          text: 'Разрешить',
          onPress: onRetry,
        },
      ]
    );
  };

  const handleImagePicker = async () => {
    try {
      console.log('[CreateEditEventModal] 🖼️ Запрос на выбор изображения');

      const hasPermission = await permissionsService.checkMediaPermission();
      
      if (!hasPermission) {
        console.log('[CreateEditEventModal] ⚠️ Нет разрешения на медиатеку');

        const granted = await permissionsService.requestMediaPermission();
        
        if (!granted) {
          console.log('[CreateEditEventModal] ❌ Пользователь отклонил разрешение');
          showPermissionAlert(handleImagePicker);
          return;
        }
      }

      console.log('[CreateEditEventModal] ✅ Разрешение получено, открываем галерею');

      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: ['images'],
        allowsEditing: true,
        aspect: [16, 9],
        quality: 0.8,
      });

      console.log('[CreateEditEventModal] 📦 Результат выбора:', result);

      if (!result.canceled && result.assets && result.assets[0]) {
        const asset = result.assets[0];
        console.log('[CreateEditEventModal] ✅ Изображение выбрано:', asset.uri);
        
        setEventImage(asset.uri);
        
        const filename = asset.uri.split('/').pop() || `event_${Date.now()}.jpg`;
        const fileType = filename.split('.').pop()?.toLowerCase() || 'jpg';
        
        setImageFile({
          uri: asset.uri,
          name: filename,
          type: `image/${fileType === 'jpg' ? 'jpeg' : fileType}`,
        });
        
        console.log('[CreateEditEventModal] ✅ Изображение сохранено в состоянии');
      } else {
        console.log('[CreateEditEventModal] ℹ️ Пользователь отменил выбор изображения');
      }
    } catch (error) {
      console.error('[CreateEditEventModal] ❌ Ошибка выбора изображения:', error);
      Alert.alert('Ошибка', 'Не удалось загрузить изображение');
    }
  };

  const handleKinopoiskButtonPress = () => {
    if (!eventName.trim()) {
      Alert.alert('Ошибка', 'Сначала введите название фильма');
      return;
    }
    setKinopoiskModalVisible(true);
  };

  const handleKinopoiskConfirm = async () => {
    try {
      setIsLoadingKinopoisk(true);
      console.log('[CreateEditEventModal] 🎬 Загрузка данных с Кинопоиска для:', eventName);

      const autoFillData = await kinopoiskService.getAutoFillData(eventName);

      if (!autoFillData) {
        Alert.alert(
          'Не найдено',
          'Фильм с таким названием не найден на Кинопоиске. Попробуйте изменить название.'
        );
        setKinopoiskModalVisible(false);
        setIsLoadingKinopoisk(false);
        return;
      }

      console.log('[CreateEditEventModal] ✅ Автозаполнение данными:', autoFillData);
      
      setDescription(autoFillData.description);
      setSelectedGenres(autoFillData.genres);
      setPublisher(autoFillData.publisher);
      setPublishYear(autoFillData.year.toString());
      setCountry(autoFillData.country);
      setAgeRating(autoFillData.ageRating);
      setDuration(autoFillData.duration);

      if (autoFillData.imageUrl) {
        setEventImage(autoFillData.imageUrl);

        const filename = `kinopoisk_${Date.now()}.jpg`;
        setImageFile({
          uri: autoFillData.imageUrl,
          name: filename,
          type: 'image/jpeg',
        });
      }

      setKinopoiskModalVisible(false);
      Alert.alert('Успешно', 'Данные загружены с Кинопоиска!');
    } catch (error: any) {
      console.error('[CreateEditEventModal] ❌ Ошибка загрузки с Кинопоиска:', error);
      Alert.alert(
        'Ошибка',
        error.message || 'Не удалось загрузить данные с Кинопоиска'
      );
    } finally {
      setIsLoadingKinopoisk(false);
      setKinopoiskModalVisible(false);
    }
  };

  const handleRawgButtonPress = () => {
    if (!eventName.trim()) {
      Alert.alert('Ошибка', 'Сначала введите название игры');
      return;
    }
    setRawgModalVisible(true);
  };

  const handleRawgConfirm = async () => {
    try {
      setIsLoadingRawg(true);
      console.log('[CreateEditEventModal] 🎮 Загрузка данных с RAWG для:', eventName);

      const autoFillData = await rawgService.getAutoFillData(eventName);

      if (!autoFillData) {
        Alert.alert(
          'Не найдено',
          'Игра с таким названием не найдена на RAWG. Попробуйте изменить название.'
        );
        setRawgModalVisible(false);
        setIsLoadingRawg(false);
        return;
      }

      console.log('[CreateEditEventModal] ✅ Автозаполнение данными:', autoFillData);
      
      setDescription(autoFillData.description);
      setSelectedGenres(autoFillData.genres);
      setPublisher(autoFillData.publisher);
      setPublishYear(autoFillData.year.toString());
      setAgeRating(autoFillData.ageRating);
      setDuration(autoFillData.duration.toString());

      if (autoFillData.imageUrl) {
        setEventImage(autoFillData.imageUrl);

        const filename = `rawg_${Date.now()}.jpg`;
        setImageFile({
          uri: autoFillData.imageUrl,
          name: filename,
          type: 'image/jpeg',
        });
      }

      setRawgModalVisible(false);
      Alert.alert('Успешно', 'Данные загружены с RAWG!');
    } catch (error: any) {
      console.error('[CreateEditEventModal] ❌ Ошибка загрузки с RAWG:', error);
      Alert.alert(
        'Ошибка',
        error.message || 'Не удалось загрузить данные с RAWG'
      );
    } finally {
      setIsLoadingRawg(false);
      setRawgModalVisible(false);
    }
  };

  const formatDateToRFC3339 = (date: Date): string => {
    const timezoneOffset = -date.getTimezoneOffset();
    const sign = timezoneOffset >= 0 ? '+' : '-';
    const hours = String(Math.floor(Math.abs(timezoneOffset) / 60)).padStart(2, '0');
    const minutes = String(Math.abs(timezoneOffset) % 60).padStart(2, '0');
    
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hour = String(date.getHours()).padStart(2, '0');
    const minute = String(date.getMinutes()).padStart(2, '0');
    const second = String(date.getSeconds()).padStart(2, '0');
    
    return `${year}-${month}-${day}T${hour}:${minute}:${second}${sign}${hours}:${minutes}`;
  };

  const handleGenreToggle = (genre: string) => {
    setSelectedGenres(prev => 
      prev.includes(genre) 
        ? prev.filter(g => g !== genre)
        : [...prev, genre]
    );
  };

  const handleAgeRatingChange = (text: string) => {
    if (text === '' || text === '+') {
      setAgeRating('');
      return;
    }

    const digitsOnly = text.replace(/\D/g, '');

    if (!digitsOnly) {
      setAgeRating('');
      return;
    }

    const limited = digitsOnly.slice(0, 2);
    setAgeRating(limited + '+');
  };

  const handleMapPress = () => {
    if (eventType === 'offline') {
      setMapModalVisible(true);
    }
  };

  const handleSelectAddress = (address: string) => {
    setEventPlace(address);
    setMapModalVisible(false);
  };

  const formatDate = (date: Date) => {
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    
    return `${day}.${month}.${year} ${hours}:${minutes}`;
  };

  const validateForm = (): boolean => {
    if (!eventName.trim()) {
      Alert.alert('Ошибка', 'Введите название события');
      return false;
    }

    if (eventName.trim().length < 5) {
      Alert.alert('Ошибка', 'Название должно содержать минимум 5 символов');
      return false;
    }

    if (description.trim()) {
      const descValidation = validateFullDescription(description);
      if (!descValidation.isValid) {
        Alert.alert(
          'Ошибка', 
          `Описание: ${descValidation.missingRequirements.join(', ')}`
        );
        return false;
      }
    }

    if (!selectedCategory) {
      Alert.alert('Ошибка', 'Выберите категорию события');
      return false;
    }

    if (selectedGenres.length === 0) {
      Alert.alert('Ошибка', 'Выберите хотя бы один жанр');
      return false;
    }

    if (!duration || isNaN(parseInt(duration))) {
      Alert.alert('Ошибка', 'Введите корректную длительность в минутах');
      return false;
    }

    if (!maxParticipants || isNaN(parseInt(maxParticipants)) || parseInt(maxParticipants) < 1) {
      Alert.alert('Ошибка', 'Введите корректное количество участников (минимум 1)');
      return false;
    }

    if (eventType === 'offline' && !eventPlace.trim()) {
      Alert.alert('Ошибка', 'Укажите место проведения для оффлайн события');
      return false;
    }

    if (!editMode && !imageFile) {
      Alert.alert('Ошибка', 'Загрузите изображение события');
      return false;
    }

    return true;
  };

  const handleDeletePress = () => {
    setShowDeleteModal(true);
  };

  const handleConfirmDelete = () => {
    if (initialData && onDelete) {
      onDelete(initialData.id);
      setShowDeleteModal(false);
    }
  };

  const handleCancelDelete = () => {
    setShowDeleteModal(false);
  };

  const handleSubmit = () => {
  console.log('[CreateEditEventModal] 🔍 selectedCategory перед отправкой:', selectedCategory);
    if (!validateForm()) {
      return;
    }

    const eventData = {
      title: profanityFilter.clean(eventName),
      description: profanityFilter.clean(description.trim()),
      category: selectedCategory,
      typeEvent: profanityFilter.clean(publisher.trim()) || 'Не указано',
      typePlace: eventType,
      genres: selectedGenres,
      publisher: profanityFilter.clean(publisher.trim()),
      year: publishYear && publishYear.trim() ? parseInt(publishYear) : undefined,
      country: country.trim(),
      ageRating: ageRating.trim(),
      date: formatDate(eventDate),
      start_time: formatDateToRFC3339(eventDate),
      duration: parseInt(duration),
      eventPlace: eventPlace.trim(),
      maxParticipants: parseInt(maxParticipants),
      image: imageFile,
      imageUri: eventImage,
      currentParticipants: initialData?.currentParticipants || 0,
    };
    console.log('[CreateEditEventModal] 📦 Финальные данные события:', eventData);
    
    if (editMode && initialData) {
      onUpdate?.(initialData.id, eventData);
    } else {
      onCreate?.(eventData);
    }
  };

  const resetForm = () => {
    setEventName('');
    setDescription('');
    setSelectedCategory('');
    setEventType('online');
    setSelectedGenres([]);
    setPublisher('');
    setPublishYear('');
    setCountry('');
    setAgeRating('');
    setEventDate(new Date());
    setDuration('');
    setEventPlace('');
    setMaxParticipants('');
    setEventImage('');
    setImageFile(null);
  };

  const handleClose = () => {
    if (!isLoading) {
      resetForm();
      onClose();
    }
  };

  useEffect(() => {
    console.log('[CreateEditEventModal] ⚠️ selectedCategory изменился на:', selectedCategory);
  }, [selectedCategory]);

  const showKinopoiskButton = selectedCategory === 'movie' && !editMode;
  const showRawgButton = selectedCategory === 'game' && !editMode;

  const isFormValid = (): boolean => {
    const isDescriptionValid = description.trim() === '' || 
    (description.trim().length >= 5 && description.trim().length <= 300);
    return !!(
      eventName.trim() &&
      eventName.trim().length >= 5 &&
      isDescriptionValid &&
      selectedCategory &&
      selectedGenres.length > 0 &&
      duration &&
      !isNaN(parseInt(duration)) &&
      maxParticipants &&
      !isNaN(parseInt(maxParticipants)) &&
      parseInt(maxParticipants) >= 1 &&
      (eventType === 'online' || eventPlace.trim()) &&
      (editMode || imageFile)
    );
  };
  return (
    <Modal visible={visible} animationType="fade" transparent>
      <View style={styles.overlay}>
        <View style={[styles.modal, { backgroundColor: colors.white }]}>
          <ScrollView
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
            nestedScrollEnabled
            bounces={false}
            scrollEnabled={!isLoading}
          >
            <View style={[styles.header, { backgroundColor: colors.white }]}>
              <Text style={[styles.title, { color: colors.black }]}>
                {editMode ? 'Редактирование события' : 'Создание события'}
              </Text>
              
              <TouchableOpacity 
                style={styles.closeButton} 
                onPress={handleClose}
                disabled={isLoading}
              >
                <Image
                  tintColor={colors.black}
                  style={{ width: 35, height: 35, resizeMode: 'cover' }}
                  source={require('@/assets/images/event_card/back.png')}
                />
              </TouchableOpacity>
            </View>

            {(isLoading || isLoadingKinopoisk || isLoadingRawg) && (
              <View style={[styles.loadingOverlay, { backgroundColor: colors.white + 'E6' }]}>
                <ActivityIndicator size="large" color={colors.lightBlue} />
                <Text style={[styles.loadingText, { color: colors.black }]}>
                  {isLoadingKinopoisk 
                    ? 'Загрузка данных с Кинопоиска...' 
                    : isLoadingRawg
                    ? 'Загрузка данных с RAWG...'
                    : editMode ? 'Сохранение изменений...' : 'Создание события...'
                  }
                </Text>
              </View>
            )}

            <View style={[styles.content, { backgroundColor: colors.white }, (isLoading || isLoadingKinopoisk || isLoadingRawg) && styles.contentDisabled]}>
              <View style={styles.nameInputContainer}>
                {showKinopoiskButton && (
                  <KinopoiskButton
                    onPress={handleKinopoiskButtonPress}
                    disabled={isLoading || !eventName.trim()}
                    loading={isLoadingKinopoisk}
                  />
                )}
                {showRawgButton && (
                  <RawgButton
                    onPress={handleRawgButtonPress}
                    disabled={isLoading || !eventName.trim()}
                    loading={isLoadingRawg}
                  />
                )}
                <TextInput
                  style={[
                    styles.input,
                    {
                      borderBottomColor: colors.grey,
                      color: colors.black
                    },
                    (showKinopoiskButton || showRawgButton) && styles.inputWithButton
                  ]}
                  placeholder="Название *"
                  placeholderTextColor={colors.grey}
                  value={eventName}
                  onChangeText={setEventName}
                  maxLength={40}
                  editable={!isLoading}
                />
              </View>
              {eventName.length > 0 && (
                <Text style={[styles.charCounter, { color: colors.grey }]}>
                  {eventName.length} / 40 {eventName.length < 5 && '(мин. 5)'}
                </Text>
              )}

              <View style={styles.descriptionContainer}>
                <TextInput
                  style={[
                    styles.input,
                    styles.textArea,
                    {
                      borderColor: colors.grey,
                      color: colors.black
                    }
                  ]}
                  placeholder="Описание"
                  placeholderTextColor={colors.grey}
                  value={description}
                  onChangeText={setDescription}
                  multiline
                  numberOfLines={4}
                  textAlignVertical="top"
                  maxLength={300}
                  editable={!isLoading}
                />
                <TouchableOpacity
                  style={styles.expandButton}
                  onPress={() => setDescriptionModalVisible(true)}
                  disabled={isLoading}
                >
                  <Image
                    source={require('@/assets/images/event_card/back.png')}
                    style={[styles.expandIcon, { tintColor: colors.grey }]}
                  />
                </TouchableOpacity>
              </View>
              {description.length > 0 && (
                <Text style={[styles.charCounter, { color: colors.grey }]}>
                  {description.length} / 300 {description.length < 5 && '(мин. 5)'}
                </Text>
              )}

              {!editMode ? (
                <CategorySelector
                  categories={categories}
                  selected={selectedCategory}
                  onSelect={setSelectedCategory}
                />
              ) : (
                <View style={styles.disabledFieldContainer}>
                  <Text style={[styles.disabledFieldLabel, { color: colors.black }]}>
                    Категория
                  </Text>
                  <View style={[
                    styles.input,
                    styles.disabledInput,
                    {
                      backgroundColor: colors.veryLightGrey,
                      borderBottomColor: colors.lightGrey
                    }
                  ]}>
                    <Text style={[styles.disabledText, { color: colors.grey }]}>
                      {categories.find(c => c.id === selectedCategory)?.label || 'Не выбрана'}
                    </Text>
                  </View>
                </View>
              )}

              {!editMode ? (
                <EventTypeSelector
                  selected={eventType}
                  onSelect={setEventType}
                />
              ) : (
                <View style={styles.disabledFieldContainer}>
                  <Text style={[styles.disabledFieldLabel, { color: colors.black }]}>
                    Тип события
                  </Text>
                  <View style={[
                    styles.input,
                    styles.disabledInput,
                    {
                      backgroundColor: colors.veryLightGrey,
                      borderBottomColor: colors.lightGrey
                    }
                  ]}>
                    <Text style={[styles.disabledText, { color: colors.grey }]}>
                      {eventType === 'online' ? 'Онлайн' : 'Оффлайн'}
                    </Text>
                  </View>
                </View>
              )}

              <GenreSelector
                selected={selectedGenres}
                onToggle={handleGenreToggle}
              />

              <TextInput
                style={[
                  styles.input,
                  {
                    borderBottomColor: colors.grey,
                    color: colors.black
                  }
                ]}
                placeholder="Издатель"
                placeholderTextColor={colors.grey}
                value={publisher}
                onChangeText={setPublisher}
                maxLength={50}
                editable={!isLoading}
              />

              <View style={styles.rowContainer}>
                <TextInput
                  style={[
                    styles.input,
                    {
                      flex: 1,
                      marginRight: 8,
                      borderBottomColor: colors.grey,
                      color: colors.black
                    }
                  ]}
                  placeholder="Год издания"
                  placeholderTextColor={colors.grey}
                  value={publishYear}
                  onChangeText={setPublishYear}
                  keyboardType="numeric"
                  maxLength={4}
                  editable={!isLoading}
                />
                <TextInput
                  style={[
                    styles.input,
                    {
                      flex: 1,
                      borderBottomColor: colors.grey,
                      color: colors.black
                    }
                  ]}
                  placeholder="Ограничение (12+)"
                  placeholderTextColor={colors.grey}
                  value={ageRating}
                  onChangeText={handleAgeRatingChange}
                  keyboardType="numeric"
                  maxLength={3}
                  editable={!isLoading}
                />
              </View>

              <View style={[
                styles.input,
                styles.disabledInput,
                {
                  backgroundColor: colors.veryLightGrey,
                  borderBottomColor: colors.lightGrey
                }
              ]}>
                <Text style={[styles.disabledText, { color: colors.grey }]}>
                  {groupName}
                </Text>
              </View>

              <View style={styles.rowContainer}>
                <TouchableOpacity 
                  style={[
                    styles.dateButton,
                    {
                      flex: 1,
                      marginRight: 6,
                      borderBottomColor: colors.grey
                    }
                  ]}
                  onPress={() => !isLoading && setDatePickerVisible(true)}
                  disabled={isLoading}
                >
                  <Text style={[styles.dateText, { color: colors.black }]}>
                    {formatDate(eventDate)}
                  </Text>
                  <Image 
                    source={require('@/assets/images/top_bar/search_bar/event-bar.png')}
                    style={[styles.calendarIcon, {tintColor: colors.grey}]}
                  />
                </TouchableOpacity>

                <TextInput
                  style={[
                    styles.input,
                    {
                      flex: 1,
                      fontSize: 14,
                      borderBottomColor: colors.grey,
                      color: colors.black
                    }
                  ]}
                  placeholder="Длительность (мин) *"
                  placeholderTextColor={colors.grey}
                  value={duration}
                  onChangeText={setDuration}
                  keyboardType="numeric"
                  maxLength={4}
                  editable={!isLoading}
                />
              </View>
              
              <View style={styles.placeContainer}>
                <TextInput
                  style={[
                    styles.input,
                    styles.placeInput,
                    {
                      borderBottomColor: colors.grey,
                      color: colors.black
                    }
                  ]}
                  placeholder={eventType === 'offline' ? 'Место проведения (адрес)' : 'Место проведения (ссылка)'}
                  placeholderTextColor={colors.grey}
                  value={eventPlace}
                  onChangeText={setEventPlace}
                  maxLength={100}
                  editable={!isLoading}
                />
                {eventType === 'offline' && (
                  <TouchableOpacity 
                    style={[styles.mapButton, { borderColor: colors.lightBlue }]} 
                    onPress={handleMapPress}
                    disabled={isLoading}
                  >
                    <Image 
                      source={require('@/assets/images/event_card/offline.png')} 
                      style={[styles.mapIcon, {tintColor: colors.black}]}
                    />
                  </TouchableOpacity>
                )}
              </View>

              <TextInput
                style={[
                  styles.input,
                  {
                    borderBottomColor: colors.grey,
                    color: colors.black
                  }
                ]}
                placeholder="Кол-во участников *"
                placeholderTextColor={colors.grey}
                value={maxParticipants}
                onChangeText={setMaxParticipants}
                keyboardType="numeric"
                maxLength={5}
                editable={!isLoading}
              />

              <TouchableOpacity 
                style={styles.imageUpload}
                onPress={handleImagePicker}
                disabled={isLoading}
                activeOpacity={0.7}
              >
                <View style={[
                  styles.uploadPlaceholder,
                  {
                    backgroundColor: colors.veryLightGrey,
                    borderColor: colors.lightGrey
                  }
                ]}>
                  {eventImage ? (
                    <Image source={{ uri: eventImage }} style={styles.eventImage} />
                  ) : (
                    <>
                      <Image 
                        source={require('@/assets/images/groups/upload_image.png')} 
                        style={[styles.uploadIcon, { tintColor: colors.grey }]}
                      />
                      <Text style={[styles.uploadText, { color: colors.grey }]}>
                        Загрузите изображение события *
                      </Text>
                    </>
                  )}
                </View>
              </TouchableOpacity>
            </View>

            <ImageBackground
              source={require('@/assets/images/event_card/bottom_rectangle.png')}
              style={styles.bottomBackground}
              resizeMode="stretch"
              tintColor={Colors.lightBlue}
            >
              <View style={styles.bottomContent}>
                <View style={styles.buttonsRow}>
                  {editMode && (
                    <TouchableOpacity 
                      style={[styles.deleteButtonBottom, isLoading && styles.buttonDisabled]} 
                      onPress={handleDeletePress}
                      disabled={isLoading}
                    >
                      <Text style={styles.deleteButtonText}>Удалить</Text>
                    </TouchableOpacity>
                  )}
                  
                <TouchableOpacity 
                  style={[
                    styles.createButton, 
                    (!isFormValid() || isLoading) && styles.buttonDisabled,
                    editMode && styles.createButtonWithDelete
                  ]} 
                  onPress={handleSubmit}
                  disabled={!isFormValid() || isLoading}
                >
                  {isLoading ? (
                    <ActivityIndicator size="small" color={Colors.blue3} />
                  ) : (
                    <Text style={styles.createButtonText}>
                      {editMode ? 'Сохранить' : 'Создать событие'}
                    </Text>
                  )}
                </TouchableOpacity>
                </View>
              </View>
            </ImageBackground>
          </ScrollView>
        </View>
      </View>

      <ConfirmationModal
        visible={kinopoiskModalVisible}
        title="Загрузить данные с Кинопоиска?"
        message={eventName 
          ? `Будет выполнен поиск фильма "${eventName}" и автоматическое заполнение полей формы.`
          : 'Будет выполнен поиск по названию фильма и автоматическое заполнение полей формы'
        }
        onConfirm={handleKinopoiskConfirm}
        onCancel={() => setKinopoiskModalVisible(false)}
      />

      <ConfirmationModal
        visible={rawgModalVisible}
        title="Загрузить данные с RAWG?"
        message={eventName 
          ? `Будет выполнен поиск игры "${eventName}" и автоматическое заполнение полей формы.`
          : 'Будет выполнен поиск по названию игры и автоматическое заполнение полей формы'
        }
        onConfirm={handleRawgConfirm}
        onCancel={() => setRawgModalVisible(false)}
      />

      <DescriptionModal
        visible={descriptionModalVisible}
        onClose={() => setDescriptionModalVisible(false)}
        description={description}
        onChangeDescription={setDescription}
      />

      <YandexMapModal
        visible={mapModalVisible}
        onClose={() => setMapModalVisible(false)}
        onSelectAddress={handleSelectAddress}
        initialAddress={eventPlace}
      />

      <ConfirmationModal
        visible={showDeleteModal}
        title="Удалить событие?"
        message="Это действие нельзя отменить. Событие будет удалено безвозвратно"
        onConfirm={handleConfirmDelete}
        onCancel={handleCancelDelete}
      />

      <Modal
        visible={datePickerVisible}
        transparent
        animationType="slide"
        onRequestClose={() => setDatePickerVisible(false)}
      >
        <View style={styles.datePickerModalOverlay}>
          <View style={[styles.datePickerModalContent, { backgroundColor: colors.white }]}>
            <View style={styles.datePickerHeader}>
              <TouchableOpacity onPress={() => setDatePickerVisible(false)}>
                <Text style={[styles.datePickerButton, { color: colors.grey }]}>
                  Отмена
                </Text>
              </TouchableOpacity>
              <Text style={[styles.datePickerTitle, { color: colors.black }]}>
                Выберите дату и время
              </Text>
              <TouchableOpacity 
                onPress={() => {
                  setDatePickerVisible(false);
                }}
              >
                <Text style={[styles.datePickerButton, { color: Colors.lightBlue }]}>
                  Готово
                </Text>
              </TouchableOpacity>
            </View>
            
            <View style={{alignItems: 'center'}}>
              <DatePicker
                date={eventDate}
                onDateChange={setEventDate}
                mode="datetime"
                minimumDate={new Date()}
                locale="ru"
                theme={isDark ? 'dark' : 'light'}
                is24hourSource="locale"
              />
            </View>
          </View>
        </View>
      </Modal>
    </Modal>
  );
};
const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
  },
  modal: {
    marginHorizontal: 12,
    borderRadius: 25,
    overflow: "hidden",
    maxHeight: screenHeight * 0.9,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 15,
    paddingHorizontal: 16,
    position: 'relative',
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    textAlign: 'left',
  },
  closeButton: {
    position: 'absolute',
    right: 16,
    zIndex: 1,
  },
  content: {
    paddingHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 16,
  },
  contentDisabled: {
    opacity: 0.6,
  },
  loadingOverlay: {
    position: 'absolute',
    top: 80,
    left: 0,
    right: 0,
    alignItems: 'center',
    zIndex: 10,
    padding: 20,
    borderRadius: 12,
    marginHorizontal: 40,
  },
  loadingText: {
    marginTop: 10,
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  nameInputContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  input: {
    flex: 1,
    borderBottomWidth: 1,
    paddingVertical: 8,
    paddingHorizontal: 0,
    marginBottom: 12,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  inputWithButton: {
    flex: 1,
    marginBottom: 0,
  },
  textArea: {
    borderWidth: 1,
    borderRadius: 8,
    padding: 16,
    minHeight: 100,
  },
  sectionLabel: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    marginBottom: 10,
  },
  rowContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center'
  },
  disabledInput: {
    borderBottomWidth: 1,
  },
  disabledText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  dateButton: {
    flexDirection: 'row',
    alignItems: 'center',
    borderBottomWidth: 1,
    paddingVertical: 9,
    marginBottom: 16,
  },
  dateText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    flex: 1,
  },
  calendarIcon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
  },
  placeContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  placeInput: {
    flex: 1,
    marginRight: 12,
  },
  mapButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    borderWidth: 2,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  mapIcon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
  },
  imageUpload: {
    alignItems: 'center',
    marginBottom: 20,
  },
  uploadPlaceholder: {
    width: '100%',
    height: 200,
    borderRadius: 12,
    borderWidth: 2,
    borderStyle: 'dashed',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: -16,
  },
  eventImage: {
    width: '100%',
    height: '100%',
    borderRadius: 10,
  },
  uploadIcon: {
    width: 60,
    height: 60,
    resizeMode: 'contain',
    marginBottom: 8,
  },
  uploadText: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    textAlign: 'center',
  },
  bottomBackground: {
    width: "100%",
    marginTop: -2
  },
  bottomContent: {
    padding: 16,
  },
  buttonsRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
  },
  deleteButtonBottom: {
    flex: 1,
    backgroundColor: Colors.white,
    paddingVertical: 8,
    paddingHorizontal: 24,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 44,
  },
  deleteButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.red,
  },
  createButton: {
    flex: 1,
    backgroundColor: Colors.white,
    paddingVertical: 8,
    paddingHorizontal: 40,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 44,
  },
  createButtonWithDelete: {
    paddingHorizontal: 24,
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  createButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.blue3,
  },
  imageUploadContainer: {
    marginBottom: 20,
  },
  disabledFieldContainer: {
    marginBottom: 16,
  },
  disabledFieldLabel: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    marginBottom: 8,
  },
  charCounter: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    marginTop: -12,
    marginBottom: 12,
  },
  descriptionContainer: {
    position: 'relative',
    marginBottom: 12,
  },
  expandButton: {
    position: 'absolute',
    right: 0,
    bottom: 12,
    width: 40,
    height: 40,
    borderRadius: 16,
    justifyContent: 'center',
    alignItems: 'center',
  },
  expandIcon: {
    width: 20,
    height: 20,
    transform: [{ rotate: '90deg' }],
  },
  datePickerModalOverlay: {
    flex: 1,
    justifyContent: 'flex-end',
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
  },
  datePickerModalContent: {
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    paddingBottom: 20,
  },
  datePickerHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: Colors.veryLightGrey,
  },
  datePickerTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
  },
  datePickerButton: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
  },
});

export default CreateEditEventModal;