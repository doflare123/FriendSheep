import kinopoiskService from '@/api/services/kinopoisk/kinopoiskService';
import ConfirmationModal from '@/components/ConfirmationModal';
import { Event } from '@/components/event/EventCard';
import KinopoiskButton from '@/components/event/KinopoiskButton';
import CategorySelector from '@/components/event/modal/CategorySelector';
import EventTypeSelector from '@/components/event/modal/EventTypeSelector';
import GenreSelector from '@/components/event/modal/GenreSelector';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import * as ImagePicker from 'expo-image-picker';
import React, { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Image,
  ImageBackground,
  Modal,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View
} from 'react-native';
import DateTimePickerModal from 'react-native-modal-datetime-picker';
import YandexMapModal from './YandexMapModal';

const screenHeight = Dimensions.get("window").height;

interface CreateEditEventModalProps {
  visible: boolean;
  onClose: () => void;
  onCreate?: (eventData: any) => void;
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
  onUpdate,
  groupName = 'Мега крутая группа',
  editMode = false,
  initialData,
  availableGenres = [],
  isLoading = false,
}) => {
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
  const [isDatePickerVisible, setDatePickerVisibility] = useState(false);

  const [kinopoiskModalVisible, setKinopoiskModalVisible] = useState(false);
  const [isLoadingKinopoisk, setIsLoadingKinopoisk] = useState(false);

  const [mapModalVisible, setMapModalVisible] = useState(false);

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
    { id: 'movie', label: 'Фильм', icon: require('@/assets/images/event_card/movie.png') },
    { id: 'game', label: 'Игра', icon: require('@/assets/images/event_card/game.png') },
    { id: 'table_game', label: 'Настолка', icon: require('@/assets/images/event_card/table_game.png') },
    { id: 'other', label: 'Другое', icon: require('@/assets/images/event_card/other.png') },
  ];

  const handleImagePicker = async () => {
    try {
      console.log('[CreateEditEventModal] 🖼️ handleImagePicker вызван');
      
      const permissionResult = await ImagePicker.requestMediaLibraryPermissionsAsync();
      
      if (!permissionResult.granted) {
        Alert.alert(
          'Требуется разрешение',
          'Для загрузки фото необходимо разрешение на доступ к галерее'
        );
        return;
      }

      console.log('[CreateEditEventModal] ✅ Разрешение получено');

      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: ['images'],
        allowsEditing: true,
        aspect: [16, 9],
        quality: 0.8,
      });

      console.log('[CreateEditEventModal] 📦 Result:', result);

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
        
        console.log('[CreateEditEventModal] ✅ Состояние обновлено');
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

  const handleMapPress = () => {
    if (eventType === 'offline') {
      setMapModalVisible(true);
    }
  };

  const handleSelectAddress = (address: string) => {
    setEventPlace(address);
    setMapModalVisible(false);
  };

  const handleDateConfirm = (date: Date) => {
    setEventDate(date);
    setDatePickerVisibility(false);
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

    if (!selectedCategory) {
      Alert.alert('Ошибка', 'Выберите категорию события');
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

  const handleSubmit = () => {
    console.log('[CreateEditEventModal] 🔍 selectedCategory перед отправкой:', selectedCategory);
    if (!validateForm()) {
      return;
    }

    const eventData = {
      title: eventName,
      description: description.trim(),
      category: selectedCategory,
      typeEvent: publisher.trim() || 'Не указано',
      typePlace: eventType,
      genres: selectedGenres,
      publisher: publisher.trim(),
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

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <View style={styles.overlay}>
        <View style={styles.modal}>
          <ScrollView
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
            nestedScrollEnabled
            bounces={false}
            scrollEnabled={!isLoading}
          >
            <View style={styles.header}>
              <Text style={styles.title}>
                {editMode ? 'Редактирование события' : 'Создание события'}
              </Text>
              <TouchableOpacity 
                style={styles.closeButton} 
                onPress={handleClose}
                disabled={isLoading}
              >
                <Image
                  tintColor={Colors.black}
                  style={{ width: 35, height: 35, resizeMode: 'cover' }}
                  source={require('@/assets/images/event_card/back.png')}
                />
              </TouchableOpacity>
            </View>

            {(isLoading || isLoadingKinopoisk) && (
              <View style={styles.loadingOverlay}>
                <ActivityIndicator size="large" color={Colors.lightBlue3} />
                <Text style={styles.loadingText}>
                  {isLoadingKinopoisk 
                    ? 'Загрузка данных с Кинопоиска...' 
                    : editMode ? 'Сохранение изменений...' : 'Создание события...'
                  }
                </Text>
              </View>
            )}

            <View style={[styles.content, (isLoading || isLoadingKinopoisk) && styles.contentDisabled]}>
              <View style={styles.nameInputContainer}>
                {showKinopoiskButton && (
                  <KinopoiskButton
                    onPress={handleKinopoiskButtonPress}
                    disabled={isLoading || !eventName.trim()}
                    loading={isLoadingKinopoisk}
                  />
                )}
                <TextInput
                  style={[styles.input, showKinopoiskButton && styles.inputWithButton]}
                  placeholder="Название *"
                  placeholderTextColor={Colors.grey}
                  value={eventName}
                  onChangeText={setEventName}
                  maxLength={100}
                  editable={!isLoading}
                />
              </View>

              <TextInput
                style={[styles.input, styles.textArea]}
                placeholder="Описание"
                placeholderTextColor={Colors.grey}
                value={description}
                onChangeText={setDescription}
                multiline
                numberOfLines={4}
                textAlignVertical="top"
                maxLength={500}
                editable={!isLoading}
              />

              {!editMode ? (
                <CategorySelector
                  categories={categories}
                  selected={selectedCategory}
                  onSelect={setSelectedCategory}
                />
              ) : (
                <View style={styles.disabledFieldContainer}>
                  <Text style={styles.disabledFieldLabel}>Категория</Text>
                  <View style={[styles.input, styles.disabledInput]}>
                    <Text style={styles.disabledText}>
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
                  <Text style={styles.disabledFieldLabel}>Тип события</Text>
                  <View style={[styles.input, styles.disabledInput]}>
                    <Text style={styles.disabledText}>
                      {eventType === 'online' ? 'Онлайн' : 'Оффлайн'}
                    </Text>
                  </View>
                </View>
              )}

              <GenreSelector
                selected={selectedGenres}
                onToggle={handleGenreToggle}
                genres={availableGenres.length > 0 ? availableGenres : undefined}
              />

              <TextInput
                style={styles.input}
                placeholder="Издатель"
                placeholderTextColor={Colors.grey}
                value={publisher}
                onChangeText={setPublisher}
                maxLength={50}
                editable={!isLoading}
              />

              <View style={styles.rowContainer}>
                <TextInput
                  style={[styles.input, { flex: 1, marginRight: 8 }]}
                  placeholder="Год издания"
                  placeholderTextColor={Colors.grey}
                  value={publishYear}
                  onChangeText={setPublishYear}
                  keyboardType="numeric"
                  maxLength={4}
                  editable={!isLoading}
                />
                <TextInput
                  style={[styles.input, { flex: 1 }]}
                  placeholder="Ограничение (12+)"
                  placeholderTextColor={Colors.grey}
                  value={ageRating}
                  onChangeText={setAgeRating}
                  maxLength={3}
                  editable={!isLoading}
                />
              </View>

              <View style={[styles.input, styles.disabledInput]}>
                <Text style={styles.disabledText}>{groupName}</Text>
              </View>

              <View style={styles.rowContainer}>
                <TouchableOpacity 
                  style={[styles.dateButton, { flex: 1, marginRight: 8 }]}
                  onPress={() => !isLoading && setDatePickerVisibility(true)}
                  disabled={isLoading}
                >
                  <Text style={styles.dateText}>{formatDate(eventDate)}</Text>
                  <Image 
                    source={require('@/assets/images/top_bar/search_bar/event-bar.png')}
                    style={styles.calendarIcon}
                  />
                </TouchableOpacity>

                <DateTimePickerModal
                  isVisible={isDatePickerVisible}
                  mode="datetime"
                  onConfirm={handleDateConfirm}
                  onCancel={() => setDatePickerVisibility(false)}
                  minimumDate={new Date()}
                />

                <TextInput
                  style={[styles.input, { flex: 1, fontSize: 14 }]}
                  placeholder="Длительность (мин) *"
                  placeholderTextColor={Colors.grey}
                  value={duration}
                  onChangeText={setDuration}
                  keyboardType="numeric"
                  maxLength={4}
                  editable={!isLoading}
                />
              </View>
              
              <View style={styles.placeContainer}>
                <TextInput
                  style={[styles.input, styles.placeInput]}
                  placeholder={eventType === 'offline' ? 'Место проведения (адрес)' : 'Место проведения (ссылка)'}
                  placeholderTextColor={Colors.grey}
                  value={eventPlace}
                  onChangeText={setEventPlace}
                  maxLength={100}
                  editable={!isLoading}
                />
                {eventType === 'offline' && (
                  <TouchableOpacity 
                    style={styles.mapButton} 
                    onPress={handleMapPress}
                    disabled={isLoading}
                  >
                    <Image 
                      source={require('@/assets/images/event_card/offline.png')} 
                      style={styles.mapIcon}
                    />
                  </TouchableOpacity>
                )}
              </View>

              <TextInput
                style={styles.input}
                placeholder="Кол-во участников *"
                placeholderTextColor={Colors.grey}
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
                <View style={styles.uploadPlaceholder}>
                  {eventImage ? (
                    <Image source={{ uri: eventImage }} style={styles.eventImage} />
                  ) : (
                    <>
                      <Image 
                        source={require('@/assets/images/groups/upload_image.png')} 
                        style={styles.uploadIcon}
                      />
                      <Text style={styles.uploadText}>Загрузите изображение события *</Text>
                    </>
                  )}
                </View>
              </TouchableOpacity>
            </View>

            <ImageBackground
              source={require('@/assets/images/event_card/bottom_rectangle.png')}
              style={styles.bottomBackground}
              resizeMode="stretch"
              tintColor={Colors.lightBlue3}
            >
              <View style={styles.bottomContent}>
                <TouchableOpacity 
                  style={[styles.createButton, isLoading && styles.createButtonDisabled]} 
                  onPress={handleSubmit}
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <ActivityIndicator size="small" color={Colors.blue3} />
                  ) : (
                    <Text style={styles.createButtonText}>
                      {editMode ? 'Сохранить изменения' : 'Создать событие'}
                    </Text>
                  )}
                </TouchableOpacity>
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
          : 'Будет выполнен поиск по названию фильма и автоматическое заполнение полей формы.'
        }
        onConfirm={handleKinopoiskConfirm}
        onCancel={() => setKinopoiskModalVisible(false)}
      />

      <YandexMapModal
        visible={mapModalVisible}
        onClose={() => setMapModalVisible(false)}
        onSelectAddress={handleSelectAddress}
        initialAddress={eventPlace}
      />
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
    backgroundColor: Colors.white,
    maxHeight: screenHeight * 0.9,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 15,
    paddingHorizontal: 16,
    backgroundColor: Colors.white,
    position: 'relative',
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
  },
  closeButton: {
    position: 'absolute',
    right: 16,
    zIndex: 1,
  },
  content: {
    backgroundColor: Colors.white,
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
    backgroundColor: 'rgba(255, 255, 255, 0.9)',
    padding: 20,
    borderRadius: 12,
    marginHorizontal: 40,
  },
  loadingText: {
    marginTop: 10,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
  },
  nameInputContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  input: {
    flex: 1,
    borderBottomWidth: 1,
    borderBottomColor: Colors.grey,
    paddingVertical: 8,
    paddingHorizontal: 0,
    marginBottom: 12,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
  },
  inputWithButton: {
    flex: 1,
    marginBottom: 0,
  },
  textArea: {
    borderWidth: 1,
    borderColor: Colors.grey,
    borderRadius: 8,
    padding: 16,
    minHeight: 100,
  },
  sectionLabel: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.black,
    marginBottom: 10,
  },
  rowContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center'
  },
  disabledInput: {
    backgroundColor: Colors.lightLightGrey,
    borderBottomColor: Colors.lightGrey,
  },
  disabledText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  dateButton: {
    flexDirection: 'row',
    alignItems: 'center',
    borderBottomWidth: 1,
    borderBottomColor: Colors.grey,
    paddingVertical: 8,
    marginBottom: 16,
  },
  dateText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    flex: 1,
  },
  calendarIcon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
    marginLeft: 10,
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
    borderColor: Colors.lightBlue3,
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
    backgroundColor: Colors.lightLightGrey,
    borderRadius: 12,
    borderWidth: 2,
    borderColor: Colors.lightGrey,
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
    tintColor: Colors.grey,
    marginBottom: 8,
  },
  uploadText: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.grey,
    textAlign: 'center',
  },
  bottomBackground: {
    width: "100%",
    marginTop: -2
  },
  bottomContent: {
    padding: 16,
  },
  createButton: {
    backgroundColor: Colors.white,
    marginHorizontal: 60,
    paddingVertical: 8,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 44,
  },
  createButtonDisabled: {
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
    color: Colors.black,
    marginBottom: 8,
  },
});

export default CreateEditEventModal;
