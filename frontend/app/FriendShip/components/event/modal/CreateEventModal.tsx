import { Event } from '@/components/event/EventCard';
import CategorySelector from '@/components/event/modal/CategorySelector';
import EventTypeSelector from '@/components/event/modal/EventTypeSelector';
import GenreSelector from '@/components/event/modal/GenreSelector';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React, { useEffect, useState } from 'react';
import {
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
import { launchImageLibrary, MediaType } from 'react-native-image-picker';
import DateTimePickerModal from 'react-native-modal-datetime-picker';

const screenHeight = Dimensions.get("window").height;

interface CreateEditEventModalProps {
  visible: boolean;
  onClose: () => void;
  onCreate?: (eventData: any) => void;
  onUpdate?: (eventId: string, eventData: any) => void;
  groupName?: string;
  editMode?: boolean;
  initialData?: Event;
}

const CreateEditEventModal: React.FC<CreateEditEventModalProps> = ({ 
  visible, 
  onClose, 
  onCreate,
  onUpdate,
  groupName = 'Мега крутая группа',
  editMode = false,
  initialData
}) => {
  const [eventName, setEventName] = useState('');
  const [description, setDescription] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [eventType, setEventType] = useState<'online' | 'offline'>('online');
  const [selectedGenres, setSelectedGenres] = useState<string[]>([]);
  const [publisher, setPublisher] = useState('');
  const [publishYear, setPublishYear] = useState('');
  const [ageRating, setAgeRating] = useState('');
  const [eventDate, setEventDate] = useState<Date>(new Date());
  const [duration, setDuration] = useState('');
  const [eventPlace, setEventPlace] = useState('');
  const [maxParticipants, setMaxParticipants] = useState('');
  const [eventImage, setEventImage] = useState<string>('');
  const [isDatePickerVisible, setDatePickerVisibility] = useState(false);

  useEffect(() => {
    if (editMode && initialData) {
      setEventName(initialData.title);
      setDescription(initialData.description);
      setSelectedCategory(initialData.category);
      setEventType(initialData.typePlace);
      setSelectedGenres(initialData.genres);
      setPublisher(initialData.publisher);
      setPublishYear(initialData.publicationDate);
      setAgeRating(initialData.ageRating);
      setDuration(initialData.duration);
      setEventPlace(initialData.eventPlace);
      setMaxParticipants(initialData.maxParticipants.toString());
      setEventImage(initialData.imageUri);

      const dateParts = initialData.date.split(' ');
      if (dateParts.length === 2) {
        const [datePart, timePart] = dateParts;
        const [day, month, year] = datePart.split('.');
        const [hours, minutes] = timePart.split(':');
        setEventDate(new Date(
          parseInt(year),
          parseInt(month) - 1,
          parseInt(day),
          parseInt(hours),
          parseInt(minutes)
        ));
      }
    }
  }, [editMode, initialData, visible]);

  const categories = [
    { id: 'movie', label: 'Фильм', icon: require('@/assets/images/event_card/movie.png') },
    { id: 'game', label: 'Игра', icon: require('@/assets/images/event_card/game.png') },
    { id: 'table_game', label: 'Настолка', icon: require('@/assets/images/event_card/table_game.png') },
    { id: 'other', label: 'Другое', icon: require('@/assets/images/event_card/other.png') },
  ];

  useEffect(() => {
    const options = {
      mediaType: 'photo' as MediaType,
      includeBase64: false,
      maxHeight: 2000,
      maxWidth: 2000,
    };

    launchImageLibrary(options, (response) => {
      if (response.didCancel || response.errorMessage) {
        return;
      }

      if (response.assets && response.assets[0]) {
        setEventImage(response.assets[0].uri || '');
      }
    });
  });

  const handleGenreToggle = (genre: string) => {
    setSelectedGenres(prev => 
      prev.includes(genre) 
        ? prev.filter(g => g !== genre)
        : [...prev, genre]
    );
  };

  const handleMapPress = () => {
    Alert.alert('Карта', 'Здесь будет открыта карта для выбора места');
  };

  const handleDateConfirm = (date: Date) => {
    setEventDate(date);
    setDatePickerVisibility(false);
  };

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const handleSubmit = () => {
    if (!eventName.trim()) {
      Alert.alert('Ошибка', 'Введите название события');
      return;
    }

    const eventData = {
      title: eventName,
      description,
      category: selectedCategory,
      typePlace: eventType,
      genres: selectedGenres,
      publisher,
      publicationDate: publishYear,
      ageRating,
      date: formatDate(eventDate),
      duration,
      eventPlace,
      maxParticipants: parseInt(maxParticipants) || 10,
      imageUri: eventImage,
      currentParticipants: initialData?.currentParticipants || 0,
    };
    
    if (editMode && initialData) {
      onUpdate?.(initialData.id, eventData);
    } else {
      onCreate?.(eventData);
    }
    
    resetForm();
    onClose();
  };

  const resetForm = () => {
    setEventName('');
    setDescription('');
    setSelectedCategory('');
    setEventType('online');
    setSelectedGenres([]);
    setPublisher('');
    setPublishYear('');
    setAgeRating('');
    setEventDate(new Date());
    setDuration('');
    setEventPlace('');
    setMaxParticipants('');
    setEventImage('');
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <View style={styles.overlay}>
        <View style={styles.modal}>
          <ScrollView
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
            nestedScrollEnabled
            bounces={false}
          >
            <View style={styles.header}>
              <Text style={styles.title}>
                {editMode ? 'Редактирование события' : 'Основная информация'}
              </Text>
              <TouchableOpacity style={styles.closeButton} onPress={handleClose}>
                <Image
                  tintColor={Colors.black}
                  style={{ width: 35, height: 35, resizeMode: 'cover' }}
                  source={require('@/assets/images/event_card/back.png')}
                />
              </TouchableOpacity>
            </View>

            <View style={styles.content}>
              <TextInput
                style={styles.input}
                placeholder="Название"
                placeholderTextColor={Colors.grey}
                value={eventName}
                onChangeText={setEventName}
                maxLength={100}
              />

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
              />

              <CategorySelector
                categories={categories}
                selected={selectedCategory}
                onSelect={setSelectedCategory}
              />

              <EventTypeSelector
                selected={eventType}
                onSelect={setEventType}
              />

              <GenreSelector
                selected={selectedGenres}
                onToggle={handleGenreToggle}
              />

              <TextInput
                style={styles.input}
                placeholder="Издатель"
                placeholderTextColor={Colors.grey}
                value={publisher}
                onChangeText={setPublisher}
                maxLength={50}
              />

              <View style={styles.rowContainer}>
                <TextInput
                  style={[styles.input]}
                  placeholder="Год издания"
                  placeholderTextColor={Colors.grey}
                  value={publishYear}
                  onChangeText={setPublishYear}
                  keyboardType="numeric"
                  maxLength={4}
                />
                <TextInput
                  style={styles.input}
                  placeholder="Ограничение"
                  placeholderTextColor={Colors.grey}
                  value={ageRating}
                  onChangeText={setAgeRating}
                  maxLength={10}
                />
              </View>

              <View style={[styles.input, styles.disabledInput]}>
                <Text style={styles.disabledText}>{groupName}</Text>
              </View>

              <View style={styles.rowContainer}>
                <TouchableOpacity 
                  style={styles.dateButton}
                  onPress={() => setDatePickerVisibility(true)}
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
                />

                <TextInput
                  style={styles.input}
                  placeholder="Длительность"
                  placeholderTextColor={Colors.grey}
                  value={duration}
                  onChangeText={setDuration}
                  maxLength={20}
                />
              </View>
              
              <View style={styles.placeContainer}>
                <TextInput
                  style={[styles.input, styles.placeInput]}
                  placeholder="Место проведения"
                  placeholderTextColor={Colors.grey}
                  value={eventPlace}
                  onChangeText={setEventPlace}
                  maxLength={100}
                />
                {eventType === 'offline' && (
                  <TouchableOpacity style={styles.mapButton} onPress={handleMapPress}>
                    <Image 
                      source={require('@/assets/images/event_card/offline.png')} 
                      style={styles.mapIcon}
                    />
                  </TouchableOpacity>
                )}
              </View>

              <TextInput
                style={styles.input}
                placeholder="Кол-во участников"
                placeholderTextColor={Colors.grey}
                value={maxParticipants}
                onChangeText={setMaxParticipants}
                keyboardType="numeric"
                maxLength={3}
              />

              <TouchableOpacity style={styles.imageUpload}>
                <View style={styles.uploadPlaceholder}>
                  {eventImage ? (
                    <Image source={{ uri: eventImage }} style={styles.eventImage} />
                  ) : (
                    <>
                      <Image 
                        source={require('@/assets/images/groups/upload_image.png')} 
                        style={styles.uploadIcon}
                      />
                      <Text style={styles.uploadText}>Загрузите своё изображение</Text>
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
                <TouchableOpacity style={styles.createButton} onPress={handleSubmit}>
                  <Text style={styles.createButtonText}>
                    {editMode ? 'Сохранить изменения' : 'Создать событие'}
                  </Text>
                </TouchableOpacity>
              </View>
            </ImageBackground>
          </ScrollView>
        </View>
      </View>
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
    fontFamily: inter.black,
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
  input: {
    borderBottomWidth: 1,
    borderBottomColor: Colors.grey,
    paddingVertical: 8,
    paddingHorizontal: 0,
    marginBottom: 16,
    fontFamily: inter.regular,
    fontSize: 16,
    color: Colors.black,
  },
  textArea: {
    borderWidth: 1,
    borderColor: Colors.grey,
    borderRadius: 8,
    padding: 16,
    minHeight: 100,
  },
  sectionLabel: {
    fontFamily: inter.bold,
    fontSize: 16,
    color: Colors.black,
    marginBottom: 10,
  },
  rowContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  disabledInput: {
    backgroundColor: Colors.lightLightGrey,
    borderBottomColor: Colors.lightGrey,
  },
  disabledText: {
    fontFamily: inter.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  dateButton: {
    flexDirection: 'row',
    alignItems: 'center',
    borderBottomWidth: 1,
    borderBottomColor: Colors.grey,
    marginBottom: 16,
  },
  dateText: {
    fontFamily: inter.regular,
    fontSize: 16,
    color: Colors.black,
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
    fontFamily: inter.bold,
    fontSize: 14,
    color: Colors.grey,
  },
  bottomBackground: {
    width: "100%",
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
  },
  createButtonText: {
    fontFamily: inter.bold,
    fontSize: 16,
    color: Colors.blue3,
  },
});

export default CreateEditEventModal;