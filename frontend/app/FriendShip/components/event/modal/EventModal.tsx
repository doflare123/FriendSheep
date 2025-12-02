import sessionService from '@/api/services/session/sessionService';
import ConfirmationModal from '@/components/ConfirmationModal';
import { useToast } from '@/components/ToastContext';
import { Event } from '@/components/event/EventCard';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
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

  useEffect(() => {
    if (visible) {
      loadSessionDetail();
    } else {
      setSessionData(null);
      setIsParticipant(false);
      setShowLeaveConfirmation(false);
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
      
      console.log('[EventModal] Данные загружены');
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

  const handleJoinLeave = async () => {
    if (isParticipant && currentParticipants === 1) {
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
    if (isParticipant && currentParticipants === 1) return 'Вы создатель';
    return isParticipant ? 'Покинуть' : 'Присоединиться';
  };

  const getButtonStyle = () => {
    if (isParticipant && currentParticipants === 1) return styles.disabledButton;
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

          <View style={styles.modal}>
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
                    tintColor={Colors.black}
                    style={{ width: 35, height: 35, resizeMode: 'cover' }}
                    source={require('@/assets/images/event_card/back.png')}
                  />
                </TouchableOpacity>

                <View style={styles.content}>
                  <View style={styles.titleRow}>
                    <Text style={styles.title} numberOfLines={2} ellipsizeMode="tail">
                      {formatTitle(sessionData?.session?.title || event.title)}
                    </Text>
                    <View style={styles.iconsRow}>
                      <View style={styles.iconOverlay}>
                        <Image
                          source={categoryIcons[event.category]}
                          style={{ resizeMode: 'contain', width: 20, height: 20 }}
                        />
                      </View>
                      <View style={styles.iconOverlay}>
                        <Image
                          source={placeIcons[event.typePlace]}
                          style={{ resizeMode: 'contain', width: 20, height: 20 }}
                        />
                      </View>
                    </View>
                  </View>

                  <Text style={styles.description}>
                    {formatDescription(sessionData?.metadata?.Notes || event.description)}
                  </Text>

                  <View style={[styles.row, { marginBottom: 8 }]}>
                    <View style={{ flexDirection: 'row', flex: 1 }}>
                      <Text style={styles.label}>Дата:</Text>
                      <Text style={styles.value} numberOfLines={1}>
                        {formatDateTime(sessionData?.session?.start_time) || event.date}
                      </Text>
                    </View>
                    <View style={{ flexDirection: 'row' }}>
                      <Text style={styles.value}>
                        {formatDuration(sessionData?.session?.duration || event.duration.replace(' мин', ''))}
                      </Text>
                      <Image
                        source={require('@/assets/images/event_card/duration.png')}
                        style={{
                          width: 20,
                          height: 20,
                          resizeMode: 'contain',
                          marginStart: 2,
                          marginTop: 6,
                        }}
                      />
                    </View>
                  </View>

                  <Text style={styles.hint}>(указано местное время)</Text>

                  <Text style={styles.label}>Жанры:</Text>
                  <View style={styles.genres}>
                    {formatGenres(sessionData?.metadata?.Genres || event.genres).map((g, index) => (
                      <View key={`${g}-${index}`} style={styles.genreBadge}>
                        <Text style={styles.genreText}>{g}</Text>
                      </View>
                    ))}
                  </View>

                  {event.group && (
                    <View style={{ marginBottom: 4 }}>
                      <Text style={styles.label}>
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

                  <Text style={[styles.label, { marginTop: 2 }]}>Место проведения:</Text>
                  <Text
                    style={[styles.value, styles.clickableText, {marginTop: 2}]}
                    onPress={handleLocationPress}
                  >
                    {formatEventPlace(sessionData?.metadata?.Location || event.eventPlace)}
                  </Text>

                  <View style={styles.row}>
                    <View style={{ flexDirection: 'row' }}>
                      <Text style={styles.label}>Участников:</Text>
                      <Text style={styles.value}>
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
                        }}
                      />
                    </View>
                  </View>
                </View>

                <ImageBackground
                  source={require('@/assets/images/event_card/bottom_rectangle.png')}
                  style={styles.bottomBackground}
                  resizeMode="stretch"
                  tintColor={Colors.lightBlue}
                >
                  <View style={styles.bottomContent}>
                    {formatPublisher(sessionData?.metadata?.Country || event.publisher) && (
                      <Text style={styles.label}>
                        Издатель: <Text style={styles.value}>
                          {formatPublisher(sessionData?.metadata?.Country || event.publisher)}
                        </Text>
                      </Text>
                    )}

                    {(sessionData?.metadata?.Year || event.publicationDate) && (
                      <Text style={styles.label}>
                        Год издания: <Text style={styles.value}>
                          {sessionData?.metadata?.Year || event.publicationDate}
                        </Text>
                      </Text>
                    )}

                    {(sessionData?.metadata?.AgeLimit || event.ageRating) && (
                      <Text style={styles.label}>
                        Возрастное ограничение: <Text style={styles.value}>
                          {sessionData?.metadata?.AgeLimit || event.ageRating}
                        </Text>
                      </Text>
                    )}

                    <TouchableOpacity
                      style={[styles.actionButton, getButtonStyle()]}
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
        message="Вы уверены, что хотите покинуть эту сессию?"
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
    backgroundColor: Colors.white,
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
    backgroundColor: Colors.white,
    borderRadius: 20,
    padding: 6,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
  },
  title: {
    flex: 1,
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.black,
    lineHeight: 24,
  },
  content: {
    backgroundColor: Colors.white,
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
    color: Colors.black,
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
    color: Colors.black, 
    fontSize: 12 
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: 8,
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
    backgroundColor: Colors.white,
  },
  leaveButton: {
    backgroundColor: Colors.white,
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
    backgroundColor: Colors.white,
  },
  hint: {
    fontFamily: Montserrat.regular,
    fontSize: 10,
    marginTop: -10,
    color: Colors.darkGrey,
    marginBottom: 4
  }
});

export default EventModal;