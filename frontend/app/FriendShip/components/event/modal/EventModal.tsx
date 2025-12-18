import groupService from '@/api/services/group/groupService';
import sessionService from '@/api/services/session/sessionService';
import ConfirmationModal from '@/components/ConfirmationModal';
import { useToast } from '@/components/ToastContext';
import { Event } from '@/components/event/EventCard';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import { addEventToCalendar, removeEventFromCalendar } from '@/utils/calendarHelper';
import { formatDuration } from '@/utils/formatDuration';
 
import { useNavigation } from '@react-navigation/native';
import React, { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Dimensions,
  Image,
  ImageBackground,
  Linking,
  Modal,
  ScrollView,
  StyleSheet,
  Switch,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View
} from 'react-native';

const screenHeight = Dimensions.get("window").height;

const categoryIcons: Record<Event["category"], any> = {
  movie: require("@/assets/images/event_card/movie.png"),
  game: require("@/assets/images/event_card/game.png"),
  table_game: require("@/assets/images/event_card/table_game.png"),
  other: require("@/assets/images/event_card/other.png"),
};

const placeIcons: Record<Event["typePlace"], any> = {
  online: require("@/assets/images/event_card/online.png"),
  offline: require("@/assets/images/event_card/offline.png"),
};

interface EventModalProps {
  visible: boolean;
  onClose: () => void;
  event: Event;
  onSessionUpdate?: () => void;
}

const formatTitle = (title: string) => {
  if (!title || title.trim().length < 5) {
    return "Без названия";
  }
  return title;
};

const formatDescription = (description?: string) => {
  if (!description || description.trim().length < 5) {
    return "Описание отсутствует";
  }

  if (description.length > 300) {
    return description.slice(0, 300) + "...";
  }

  return description;
};

const formatGenres = (genres: string[]) => {
  if (!genres || genres.length < 1) return ["Жанр отсутствует"];
  return genres.slice(0, 9);
};

const formatEventPlace = (place?: string) => {
  if (!place || place.trim().length < 5) return "Место не указано";
  if (place.length > 200) return place.slice(0, 200) + "...";
  return place;
};

const formatPublisher = (publisher?: string) => {
  if (!publisher || publisher.trim().length < 5) return null;
  if (publisher.length > 40) return publisher.slice(0, 40) + "...";
  return publisher;
};

const EventModal: React.FC<EventModalProps> = ({ 
  visible, 
  onClose, 
  event,
  onSessionUpdate 
}) => {
  const colors = useThemedColors();
  const { showToast } = useToast();
  const [isLoading, setIsLoading] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [sessionData, setSessionData] = useState<any>(null);
  const [isParticipant, setIsParticipant] = useState(false);
  const [isCreator, setIsCreator] = useState(false);
  const [showLeaveConfirmation, setShowLeaveConfirmation] = useState(false);
  const [currentParticipants, setCurrentParticipants] = useState(event.currentParticipants);
  const [maxParticipants, setMaxParticipants] = useState(event.maxParticipants);
  
  const [showLinkConfirmation, setShowLinkConfirmation] = useState(false);
  const [pendingLink, setPendingLink] = useState<string>('');

  const [addToCalendar, setAddToCalendar] = useState(false);
  const [calendarEventId, setCalendarEventId] = useState<string | undefined>(event.calendarEventId);
  

  useEffect(() => {
    if (visible) {
      loadSessionDetail();
    } else {
      setSessionData(null);
      setIsParticipant(false);
      setShowLeaveConfirmation(false);
      setAddToCalendar(false);
    }
  }, [visible, event.id]);

  const loadSessionDetail = async () => {
    try {
      setIsLoading(true);
      console.log('[EventModal] Загружаем детальные данные сессии');

      const data = await sessionService.getSessionDetail(parseInt(event.id));

      setSessionData(data);
      setCurrentParticipants(data.session.current_users);
      setMaxParticipants(data.session.count_users_max);

      const userIsParticipant = data.session.is_sub === true;
      setIsParticipant(userIsParticipant);

      if (userIsParticipant) {
        try {
          const savedCalendarEventId = await sessionService.getCalendarEventId(parseInt(event.id));
          if (savedCalendarEventId) {
            setCalendarEventId(savedCalendarEventId);
            setAddToCalendar(true);
            console.log('[EventModal] ✅ Найден сохранённый calendarEventId:', savedCalendarEventId);
          }
        } catch (calError) {
          console.log('[EventModal] ⚠️ Не удалось загрузить calendarEventId:', calError);
        }
      }

      try {
        await groupService.getGroupDetail(data.session.group_id);
        setIsCreator(userIsParticipant);
        console.log('[EventModal] ✅ Админ группы, считаем создателем события');
      } catch (error) {
        console.log('[EventModal] ⚠️ Не админ группы');
        setIsCreator(false);
      }
    } catch (error: any) {
      console.error('[EventModal] Ошибка загрузки:', error);
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось загрузить сессию',
      });
    } finally {
      setIsLoading(false);
    }
  };

  const parseEventDate = (dateString: string): Date => {
    const [datePart, timePart] = dateString.split(' ');
    const [day, month, year] = datePart.split('.');
    const [hours, minutes] = timePart ? timePart.split(':') : ['0', '0'];
    
    return new Date(
      parseInt(year),
      parseInt(month) - 1,
      parseInt(day),
      parseInt(hours),
      parseInt(minutes)
    );
  };

  const handleJoinLeave = async () => {
    if (isCreator) {
      showToast({
        type: 'error',
        title: 'Действие недоступно',
        message: 'Вы создатель события. Удалите событие, если оно больше не нужно.',
      });
      return;
    }

    if (isParticipant) {
      setShowLeaveConfirmation(true);
    } else {
      await handleJoin();
    }
  };

  const handleAddToCalendarToggle = async (value: boolean) => {
    if (!sessionData) return;

    try {
      if (value) {
        console.log('[EventModal] 📅 Добавление события в календарь');
        
        const startDate = new Date(sessionData.session.start_time);
        const durationMinutes = sessionData.session.duration;
        const endDate = new Date(startDate.getTime() + durationMinutes * 60000);

        console.log('[EventModal] 📅 Дата начала:', startDate.toISOString());
        console.log('[EventModal] 📅 Дата окончания:', endDate.toISOString());

        const eventId = await addEventToCalendar({
          title: sessionData.session.title,
          location: sessionData.session.session_place === 'Оффлайн' 
            ? (sessionData.metadata?.Location || event.eventPlace)
            : undefined,
          startDate,
          endDate,
          notes: sessionData.metadata?.Notes || event.description,
          groupName: event.group,
        });

        setCalendarEventId(eventId);
        setAddToCalendar(true);

        await sessionService.saveCalendarEventId(parseInt(event.id), eventId);

        showToast({
          type: 'success',
          title: 'Успешно',
          message: 'Событие добавлено в календарь',
        });

      } else {
        if (calendarEventId) {
          console.log('[EventModal] 🗑️ Удаление события из календаря');
          await removeEventFromCalendar(calendarEventId);
          await sessionService.removeCalendarEventId(parseInt(event.id));
          
          setCalendarEventId(undefined);
          setAddToCalendar(false);

          showToast({
            type: 'success',
            title: 'Готово',
            message: 'Событие удалено из календаря',
          });
        }
      }
    } catch (error: any) {
      console.error('[EventModal] ❌ Ошибка работы с календарём:', error);
      setAddToCalendar(!value);
      
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось обновить календарь',
      });
    }
  };

  const handleJoin = async () => {
    try {
      setIsProcessing(true);
      console.log('[EventModal] 🚪 Присоединение к сессии');

      await sessionService.joinSession({
        group_id: sessionData.session.group_id,
        session_id: parseInt(event.id),
      });

      setIsParticipant(true);
      setCurrentParticipants(prev => prev + 1);

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: `Вы зарегистрированы на событие "${event.title}"`,
      });

      await loadSessionDetail();
      onSessionUpdate?.();
    } catch (error: any) {
      console.error('[EventModal] ❌ Ошибка присоединения:', error);

      if (error.message?.includes('уже присоединились')) {
        setIsParticipant(true);
        await loadSessionDetail();
      }
      
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось присоединиться к сессии',
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const handleLeave = async () => {
    try {
      setIsProcessing(true);
      setShowLeaveConfirmation(false);
      console.log('[EventModal] 🚪 Выход из сессии');

      await sessionService.leaveSession(parseInt(event.id));

      if (calendarEventId) {
        try {
          await removeEventFromCalendar(calendarEventId);
          await sessionService.removeCalendarEventId(parseInt(event.id));
          setCalendarEventId(undefined);
          setAddToCalendar(false);
        } catch (calError) {
          console.error('[EventModal] Ошибка удаления из календаря:', calError);
        }
      }

      setIsParticipant(false);
      setCurrentParticipants(prev => Math.max(0, prev - 1));

      showToast({
        type: 'success',
        title: 'Успешно',
        message: 'Вы покинули сессию',
      });

      await loadSessionDetail();
      onSessionUpdate?.();

      onClose();
    } catch (error: any) {
      console.error('[EventModal] ❌ Ошибка выхода:', error);
      
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось покинуть сессию',
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const getButtonText = () => {
    if (isProcessing) return 'Загрузка...';
    if (isCreator) return 'Вы создатель';
    return isParticipant ? 'Покинуть' : 'Присоединиться';
  };

  const getButtonStyle = () => {
    if (isCreator) return styles.disabledButton;
    return isParticipant ? styles.leaveButton : styles.joinButton;
  };

  const formatDateTime = (isoDate?: string) => {
    if (!isoDate) return 'Дата не указана';
    
    const date = new Date(isoDate);
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    
    return `${day}.${month}.${year} ${hours}:${minutes}`;
  };

  const navigation = useNavigation<any>();

  const handleGroupPress = () => {
    if (sessionData?.session?.group_id) {
      onClose();
      navigation.navigate('GroupPage', { 
        groupId: sessionData.session.group_id 
      });
    }
  }
  
  const handleLocationPress = () => {
    const location = sessionData?.metadata?.Location || event.eventPlace;
    
    if (!location || location === 'Место не указано') {
      return;
    }

    const isLink = location.startsWith('http://') || location.startsWith('https://');
    
    if (isLink) {
      setPendingLink(location);
      setShowLinkConfirmation(true);
    } else {
      const mapsUrl = `https://yandex.ru/maps/?text=${encodeURIComponent(location)}`;
      Linking.openURL(mapsUrl);
    }
  };

  const handleConfirmOpenLink = () => {
    if (pendingLink) {
      Linking.openURL(pendingLink);
      setShowLinkConfirmation(false);
      setPendingLink('');
    }
  };

  return (
    <>
      <Modal visible={visible} animationType="fade" transparent>
        <View style={styles.overlay}>
          <TouchableWithoutFeedback onPress={onClose}>
            <View style={StyleSheet.absoluteFill} />
          </TouchableWithoutFeedback>

          <View style={[styles.modal, { backgroundColor: colors.white }]}>
            {isLoading ? (
              <View style={styles.loadingContainer}>
                <ActivityIndicator size="large" color={Colors.lightBlue} />
                <Text style={styles.loadingText}>Загрузка...</Text>
              </View>
            ) : (
              <ScrollView
                showsVerticalScrollIndicator={false}
                keyboardShouldPersistTaps="handled"
                nestedScrollEnabled
                bounces={false}
                alwaysBounceVertical={false}
              >
                <View style={styles.header}>
                  <Image 
                    source={{ uri: sessionData?.session?.image_url || event.imageUri }} 
                    style={styles.image} 
                  />
                </View>

                <TouchableOpacity style={styles.closeButton} onPress={onClose}>
                  <Image
                    tintColor={colors.black}
                    style={{ width: 35, height: 35, resizeMode: 'cover' }}
                    source={require('@/assets/images/event_card/back.png')}
                  />
                </TouchableOpacity>

                <View style={[styles.content, { backgroundColor: colors.white }]}>
                  <View style={styles.titleRow}>
                    <Text style={[styles.title, { color: colors.black }]} numberOfLines={2} ellipsizeMode="tail">
                      {formatTitle(sessionData?.session?.title || event.title)}
                    </Text>
                    <View style={styles.iconsRow}>
                      <View style={[styles.iconOverlay, { backgroundColor: colors.white }]}>
                        <Image
                          source={categoryIcons[event.category]}
                          style={{ resizeMode: 'contain', width: 20, height: 20, tintColor: colors.black }}
                        />
                      </View>
                      <View style={[styles.iconOverlay, { backgroundColor: colors.white }]}>
                        <Image
                          source={placeIcons[event.typePlace]}
                          style={{ resizeMode: 'contain', width: 20, height: 20, tintColor: colors.black }}
                        />
                      </View>
                    </View>
                  </View>

                  <Text style={[styles.description, { color: colors.black }]}>
                    {formatDescription(sessionData?.metadata?.Notes || event.description)}
                  </Text>

                  <View style={[styles.row, { marginBottom: 8 }]}>
                    <View style={{ flexDirection: 'row', flex: 1 }}>
                      <Text style={[styles.label, { color: colors.black }]}>Дата:</Text>
                      <Text style={[styles.value, { color: colors.black }]} numberOfLines={1}>
                        {formatDateTime(sessionData?.session?.start_time) || event.date}
                      </Text>
                    </View>
                    <View style={{ flexDirection: 'row' }}>
                      <Text style={[styles.value, { color: colors.black }]}>
                        {formatDuration(sessionData?.session?.duration || event.duration.replace(' мін', ''))}
                      </Text>
                      <Image
                        source={require('@/assets/images/event_card/duration.png')}
                        style={{
                          width: 20,
                          height: 20,
                          resizeMode: 'contain',
                          marginStart: 2,
                          marginTop: 6,
                          tintColor: colors.black
                        }}
                      />
                    </View>
                  </View>

                  <Text style={[styles.hint, {color: colors.darkGrey}]}>(указано местное время)</Text>

                  <Text style={[styles.label, { color: colors.black }]}>Жанры:</Text>
                  <View style={styles.genres}>
                    {formatGenres(sessionData?.metadata?.Genres || event.genres).map((g, index) => (
                      <View key={`${g}-${index}`} style={styles.genreBadge}>
                        <Text style={[styles.genreText, { color: colors.black }]}>{g}</Text>
                      </View>
                    ))}
                  </View>

                  {event.group && (
                    <View style={{ marginBottom: 4 }}>
                      <Text style={[styles.label, { color: colors.black }]}>
                        Организатор:{' '}
                        <Text
                          style={[styles.value, styles.clickableText]}
                          onPress={handleGroupPress}
                        >
                          {event.group}
                        </Text>
                      </Text>
                    </View>
                  )}

                  <Text style={[styles.label, { marginTop: 2, color: colors.black }]}>Место проведения:</Text>
                  <Text
                    style={[styles.value, styles.clickableText, {marginTop: 2}]}
                    onPress={handleLocationPress}
                  >
                    {formatEventPlace(sessionData?.metadata?.Location || event.eventPlace)}
                  </Text>

                  <View style={styles.row}>
                    <View style={{ flexDirection: 'row' }}>
                      <Text style={[styles.label, { color: colors.black }]}>Участников:</Text>
                      <Text style={[styles.value, { color: colors.black }]}>
                        {currentParticipants}/{maxParticipants}
                      </Text>
                      <Image
                        source={require('@/assets/images/event_card/person.png')}
                        style={{
                          width: 20,
                          height: 20,
                          resizeMode: 'contain',
                          marginStart: 2,
                          marginTop: 6,
                          tintColor: colors.black
                        }}
                      />
                    </View>
                  </View>

                  {isParticipant && (
                    <View style={styles.calendarSection}>
                      <View style={styles.calendarRow}>
                        <View style={styles.calendarInfo}>
                          <Text style={[styles.calendarLabel, { color: colors.black }]}>Добавить в календарь</Text>
                          <Text style={[styles.calendarHint, {color: colors.darkGrey}]}>
                            Событие будет добавлено в ваш календарь
                          </Text>
                        </View>
                        <Switch
                          value={addToCalendar}
                          onValueChange={handleAddToCalendarToggle}
                          trackColor={{ false: Colors.lightGrey, true: Colors.lightBlue }}
                          thumbColor={addToCalendar ? Colors.white : Colors.white}
                        />
                      </View>
                    </View>
                  )}
                </View>

                <ImageBackground
                  source={require('@/assets/images/event_card/bottom_rectangle.png')}
                  style={styles.bottomBackground}
                  resizeMode="stretch"
                  tintColor={colors.lightBlue}
                >
                  <View style={styles.bottomContent}>
                    {formatPublisher(sessionData?.metadata?.Country || event.publisher) && (
                      <Text style={[styles.label, { color: colors.black }]}>
                        Издатель: <Text style={[styles.value, { color: colors.black }]}>
                          {formatPublisher(sessionData?.metadata?.Country || event.publisher)}
                        </Text>
                      </Text>
                    )}

                    {(sessionData?.metadata?.Year || event.publicationDate) && (
                      <Text style={[styles.label, { color: colors.black }]}>
                        Год издания: <Text style={[styles.value, { color: colors.black }]}>
                          {sessionData?.metadata?.Year || event.publicationDate}
                        </Text>
                      </Text>
                    )}

                    {(sessionData?.metadata?.AgeLimit || event.ageRating) && (
                      <Text style={[styles.label, { color: colors.black }]}>
                        Возрастное ограничение: <Text style={[styles.value, { color: colors.black }]}>
                          {sessionData?.metadata?.AgeLimit || event.ageRating}
                        </Text>
                      </Text>
                    )}

                    <TouchableOpacity
                      style={[styles.actionButton, getButtonStyle(), { backgroundColor: Colors.white }]}
                      onPress={handleJoinLeave}
                      disabled={isProcessing}
                    >
                      {isProcessing ? (
                        <ActivityIndicator color={Colors.white} size="small" />
                      ) : (
                        <Text style={styles.actionButtonText}>{getButtonText()}</Text>
                      )}
                    </TouchableOpacity>
                  </View>
                </ImageBackground>
              </ScrollView>
            )}
          </View>
        </View>
      </Modal>

      <ConfirmationModal
        visible={showLeaveConfirmation}
        title="Покинуть сессию?"
        message="Вы уверены, что хотите покинуть эту сессию? Событие также будет удалено из календаря."
        onConfirm={handleLeave}
        onCancel={() => setShowLeaveConfirmation(false)}
      />

      <ConfirmationModal
        visible={showLinkConfirmation}
        title="Открыть ссылку?"
        message={`Вы будете перенаправлены на:\n${pendingLink}`}
        onConfirm={handleConfirmOpenLink}
        onCancel={() => {
          setShowLinkConfirmation(false);
          setPendingLink('');
        }}
      />
    </>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
  },
  modal: {
    marginHorizontal: 8,
    borderRadius: 30,
    overflow: "hidden",
    maxHeight: screenHeight * 0.85,
  },
  loadingContainer: {
    padding: 60,
    alignItems: 'center',
    justifyContent: 'center',
  },
  loadingText: {
    marginTop: 12,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  header: {
    alignItems: 'center',
    position: 'relative',
  },
  image: {
    width: '100%',
    height: 180,
    resizeMode: 'cover',
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  iconsRow: {
    flexDirection: 'row',
    gap: 8,
    marginLeft: 6,
  },
  iconOverlay: {
    borderRadius: 20,
    padding: 6,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
  },
  title: {
    flex: 1,
    fontFamily: Montserrat.bold,
    fontSize: 20,
    lineHeight: 24,
  },
  content: {
    padding: 16,
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    marginBottom: 16,
    lineHeight: 18,
  },
  label: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    marginEnd: 6,
    marginTop: 6,
  },
  value: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    marginTop: 6,
  },
  genres: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginVertical: 6,
    marginBottom: 0
  },
  genreBadge: {
    marginRight: 6,
    backgroundColor: Colors.lightBlue,
    borderRadius: 20,
    paddingHorizontal: 8,
    paddingVertical: 2,
    marginBottom: 8,
  },
  genreText: { 
    fontFamily: Montserrat.regular, 
    fontSize: 12 
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: 8,
  },
  calendarSection: {
    marginTop: 16,
    paddingTop: 16,
    borderTopWidth: 1,
    borderTopColor: Colors.lightGrey,
  },
  calendarRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  calendarInfo: {
    flex: 1,
    marginRight: 12,
  },
  calendarLabel: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    marginBottom: 4,
  },
  calendarHint: {
    fontFamily: Montserrat.regular,
    fontSize: 11,
  },
  bottomBackground: {
    width: "100%",
  },
  bottomContent: {
    padding: 16,
  },
  actionButton: {
    marginTop: 16,
    marginHorizontal: 60,
    paddingVertical: 10,
    borderRadius: 20,
    alignItems: 'center',
  },
  joinButton: {
  },
  leaveButton: {
  },
  actionButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.red,
  },
  closeButton: { 
    position: 'absolute', 
    top: 5, 
    right: 10, 
    zIndex: 10 
  },
  clickableText: {
    color: Colors.lightBlue,
  },
  disabledButton: {
  },
  hint: {
    fontFamily: Montserrat.regular,
    fontSize: 10,
    marginTop: -10,
    marginBottom: 4
  }
});

export default EventModal;