import { useToast } from '@/components/ToastContext';
import { Event } from '@/components/event/EventCard';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import {
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
  onJoin?: () => void;
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

const EventModal: React.FC<EventModalProps> = ({ visible, onClose, event }) => {
  const { showToast } = useToast();

  const handleJoin = () => {
    onClose();
    showToast({
      type: 'success',
      title: 'Успешно!',
      message: `Вы зарегистрированы на событие "${event.title}" ${event.date}`,
    });
  };

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <View style={styles.overlay}>
        <TouchableWithoutFeedback onPress={onClose}>
          <View style={StyleSheet.absoluteFill} />
        </TouchableWithoutFeedback>

        <View style={styles.modal}>
          <ScrollView
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
            nestedScrollEnabled
            bounces={false}
            alwaysBounceVertical={false}
          >
              <View style={styles.header}>
                <Image source={{ uri: event.imageUri }} style={styles.image} />
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
                    {formatTitle(event.title)}
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
                  {formatDescription(event.description)}
                </Text>

                <View style={[styles.row, { marginBottom: 8 }]}>
                  <View style={{ flexDirection: 'row' }}>
                    <Text style={styles.label}>Дата:</Text>
                    <Text style={styles.value}>{event.date}</Text>
                  </View>
                  <View style={{ flexDirection: 'row' }}>
                    <Text style={styles.value}>{event.duration}</Text>
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

                <Text style={styles.label}>Жанры:</Text>
                <View style={styles.genres}>
                  {formatGenres(event.genres).map((g) => (
                    <View key={g} style={styles.genreBadge}>
                      <Text style={styles.genreText}>{g}</Text>
                    </View>
                  ))}
                </View>

                {event.group && (
                  <Text style={[styles.label, {marginBottom: 4}]}>
                    Организатор: <Text style={styles.value}>{event.group}</Text>
                  </Text>
                )}

                <Text style={[styles.label, {marginTop: 2}]}>Место проведения:</Text>
                <Text
                  style={[styles.value, { color: Colors.lightBlue3, marginTop: 0 }]}
                  onPress={() => Linking.openURL(event.eventPlace)}
                >
                  {formatEventPlace(event.eventPlace)}
                </Text>

                <View style={styles.row}>
                  <View style={{ flexDirection: 'row' }}>
                    <Text style={styles.label}>Участников:</Text>
                    <Text style={styles.value}>
                      {event.currentParticipants}/{event.maxParticipants}
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
                tintColor={Colors.lightBlue3}
              >
                <View style={styles.bottomContent}>
                  {formatPublisher(event.publisher) && (
                    <Text style={styles.label}>
                      Издатель: <Text style={styles.value}>{formatPublisher(event.publisher)}</Text>
                    </Text>
                  )}

                  {event.publicationDate && (
                    <Text style={styles.label}>
                      Год издания: <Text style={styles.value}>{event.publicationDate}</Text>
                    </Text>
                  )}

                  {event.ageRating && (
                    <Text style={styles.label}>
                      Возрастное ограничение: <Text style={styles.value}>{event.ageRating}</Text>
                    </Text>
                  )}

                  <TouchableOpacity
                    style={styles.joinButton}
                    onPress={handleJoin}
                  >
                    <Text style={styles.joinButtonText}>Присоединиться</Text>
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
    marginHorizontal: 8,
    borderRadius: 30,
    overflow: "hidden",
    backgroundColor: Colors.white,
    maxHeight: screenHeight * 0.85,
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
    borderColor: Colors.lightBlue3,
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
    backgroundColor: Colors.lightBlue3,
    borderRadius: 20,
    paddingHorizontal: 8,
    paddingVertical: 2,
    marginBottom: 8,
  },
  genreText: { fontFamily: Montserrat.regular, color: Colors.black, fontSize: 12 },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: 8,
  },
  rowBetween: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: 8,
  },
  bottomBackground: {
    width: "100%",
  },
  bottomContent: {
    padding: 16,
  },
  joinButton: {
    marginTop: 16,
    backgroundColor: Colors.white,
    marginHorizontal: 60,
    paddingVertical: 6,
    borderRadius: 20,
    alignItems: 'center',
  },
  joinButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.blue3,
  },
  closeButton: { position: 'absolute', top: 5, right: 10, zIndex: 10 },
});

export default EventModal;