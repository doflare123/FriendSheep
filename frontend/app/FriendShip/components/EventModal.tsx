import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
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
import { Event } from './EventCard';
import Toast from './Toast';

const screenHeight = Dimensions.get("window").height;

const categoryIcons: Record<Event["category"], any> = {
  movie: require("../assets/images/event_card/movie.png"),
  game: require("../assets/images/event_card/game.png"),
  table_game: require("../assets/images/event_card/table_game.png"),
  other: require("../assets/images/event_card/other.png"),
};

const placeIcons: Record<Event["typePlace"], any> = {
  online: require("../assets/images/event_card/online.png"),
  offline: require("../assets/images/event_card/offline.png"),
};

interface EventModalProps {
  visible: boolean;
  onClose: () => void;
  event: Event;
  onJoin?: () => void;
}

const formatTitle = (title: string) => {
  const maxLength = 40;
  const firstLineLimit = 19;
  const secondLineLimit = 21;

  if (!title || title.trim().length < 5) {
    return "Без названия";
  }

  let trimmed = title.slice(0, maxLength);

  let firstLine = trimmed.slice(0, firstLineLimit);
  let secondLine = trimmed.slice(firstLineLimit, firstLineLimit + secondLineLimit);

  return firstLine + (secondLine ? "\n" + secondLine : "");
};

const getFontSize = (title: string) => {
  const formatted = formatTitle(title);
  const isTwoLines = formatted.includes("\n");

  const letters = title.replace(/[^a-zA-ZА-Яа-яЁё]/g, "");
  const upperLetters = letters.replace(/[^A-ZА-ЯЁ]/g, "");

  const upperRatio = letters.length > 0 ? upperLetters.length / letters.length : 0;
  const wideLetters = title.match(/[ЖШЩМW]/g) || [];

  if (wideLetters.length > title.length * 0.3) return 8.5;

  if (upperRatio >= 0.4) {
    return 11;
  }

  if (title === title.toUpperCase()) {
    return 13;
  }

  if (title.length > 35) return 16;
  if (title.length > 25) return 17;
  return 19;
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
  const [toastVisible, setToastVisible] = React.useState(false);

  const isTwoLines = formatTitle(event.title).includes("\n");
  const rectangleHeight = isTwoLines ? 75 : 55;

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
                <Image
                  source={require('../assets/images/event_card/rectangle.png')}
                  style={[styles.rectangle, { height: rectangleHeight }]}
                />
                <Text
                  style={[styles.title, { fontSize: getFontSize(event.title) }]}
                  numberOfLines={2}
                  ellipsizeMode="tail"
                >
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

              <View style={styles.content}>
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
                      source={require('../assets/images/event_card/duration.png')}
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

                <Text style={[styles.label, {marginTop: 2}]}>Место проведения:</Text>
                <Text
                  style={[styles.value, { color: Colors.lightBlue, marginTop: 0 }]}
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
                      source={require('../assets/images/event_card/person.png')}
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

              <TouchableOpacity style={styles.closeButton} onPress={onClose}>
                <Image
                  tintColor={Colors.black}
                  style={{ width: 35, height: 35, resizeMode: 'cover' }}
                  source={require('../assets/images/event_card/back.png')}
                />
              </TouchableOpacity>

              <ImageBackground
                source={require('../assets/images/event_card/bottom_rectangle.png')}
                style={styles.bottomBackground}
                resizeMode="stretch"
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
                    onPress={() => setToastVisible(true)}
                  >
                    <Text style={styles.joinButtonText}>Присоединиться</Text>
                  </TouchableOpacity>
                </View>
              </ImageBackground>

              {toastVisible && (
                <Toast
                  visible={toastVisible}
                  type="success"
                  title="Успешно!"
                  message={`Вы зарегистрированы на событие "${event.title}" ${event.date}`}
                  onHide={() => setToastVisible(false)}
                />
              )}
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
  rectangle: {
    position: 'absolute',
    left: 0,
    bottom: -20,
    width: 300,
    height: 75,
    resizeMode: 'stretch',
  },
  title: {
    position: 'absolute',
    maxWidth: 240,
    overflow: "hidden",
    left: 16,
    right: 16,
    bottom: 0,
    fontFamily: inter.black,
    color: Colors.black,
    textAlign: 'left',
    width: 240,
    lineHeight: 22,
  },
  iconsRow: {
    position: 'absolute',
    top: 140,
    right: 10,
    flexDirection: 'row',
    gap: 6,
  },
  iconOverlay: {
    backgroundColor: Colors.white,
    borderRadius: 20,
    padding: 6,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
  },
  content: {
    backgroundColor: Colors.white,
    padding: 16,
  },
  description: {
    fontFamily: inter.regular,
    fontSize: 14,
    marginBottom: 16,
    lineHeight: 18,
  },
  label: {
    fontFamily: inter.bold,
    fontSize: 14,
    marginEnd: 6,
    marginTop: 6,
  },
  value: {
    fontFamily: inter.regular,
    fontSize: 14,
    color: Colors.black,
    marginTop: 6,
  },
  genres: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginVertical: 6,
  },
  genreBadge: {
    marginRight: 6,
    backgroundColor: Colors.lightBlue,
    borderRadius: 20,
    paddingHorizontal: 8,
    paddingVertical: 2,
    marginBottom: 8,
  },
  genreText: { fontFamily: inter.regular, color: Colors.black, fontSize: 12 },
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
  bottomWave: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    width: '100%',
    height: 160,
    resizeMode: 'stretch',
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
    fontFamily: inter.bold,
    fontSize: 16,
    color: Colors.blue,
  },
  closeButton: { position: 'absolute', top: 5, right: 10 },
});

export default EventModal;
