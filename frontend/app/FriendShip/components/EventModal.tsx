import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React from 'react';
import {
  Image,
  Linking,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View
} from 'react-native';
import { Event } from './EventCard';
import Toast from './Toast';

interface EventModalProps {
  visible: boolean;
  onClose: () => void;
  event: Event;
  onJoin?: () => void;
}

const EventModal: React.FC<EventModalProps> = ({ visible, onClose, event }) => {

  const [toastVisible, setToastVisible] = React.useState(false);

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <TouchableWithoutFeedback onPress={onClose}>
        <View style={styles.overlay}>
          <TouchableWithoutFeedback onPress={() => {}}>
            <View style={styles.modal}>
                <View style={styles.header}>
                  <Image source={{ uri: event.imageUri }} style={styles.image} />
                  <Image
                    source={require('../assets/images/event_card/rectangle.png')}
                    style={styles.rectangle}
                  />
                  <Text style={styles.title}>{event.title}</Text>
                  <View style={styles.iconsRow}>
                      <View style={styles.iconOverlay}>
                        <Image
                          source={require("../assets/images/event_card/movie.png")}
                          style={{ resizeMode: 'contain', width: 20, height: 20 }}
                        />
                      </View>
                      <View style={styles.iconOverlay}>
                        <Image
                          source={require("../assets/images/event_card/online.png")}
                          style={{ resizeMode: 'contain', width: 20, height: 20 }}
                        />
                      </View>
                    </View>
                </View>

                <View style={styles.content}>
                  <Text style={styles.description}>
                    {event.description ||
                      'Описание отсутствует'}
                  </Text>

                  <View style={[styles.row, { marginBottom: 8}]}>
                    <View style={{flexDirection: 'row'}}>
                      <Text style={styles.label}>Дата:</Text>
                      <Text style={styles.value}>{event.date}</Text>
                    </View>
                    <View style={{flexDirection: 'row'}}>
                      <Text style={styles.value}>{event.duration}</Text>
                      <Image source={require('../assets/images/event_card/duration.png')} style={{width: 20, height: 20, resizeMode: 'contain', marginStart: 2, marginTop: 6}}/>
                    </View>
                  </View>

                  <Text style={styles.label}>Жанры:</Text>
                  <View style={styles.genres}>
                    {event.genres.map((g) => (
                      <View key={g} style={styles.genreBadge}>
                        <Text style={styles.genreText}>{g}</Text>
                      </View>
                    ))}
                  </View>

                  <Text style={styles.label}>Место проведения:</Text>
                  <Text
                    style={[styles.value, { color: Colors.lightBlue, marginTop: 0 }]}
                    onPress={() => Linking.openURL(event.eventPlace)}
                  >
                    {event.eventPlace}
                  </Text>

                  <View style={[styles.row, {marginBottom: 150}]}>
                    <View style={{flexDirection: 'row'}}>
                      <Text style={styles.label}>Участников:</Text>
                      <Text style={styles.value}>{event.participants}</Text>
                      <Image source={require('../assets/images/event_card/person.png')} style={{width: 20, height: 20, resizeMode: 'contain', marginStart: 2, marginTop: 6}}/>
                    </View>
                  </View> 
                </View>

                <Image
                  source={require('../assets/images/event_card/bottom_rectangle.png')}
                  style={styles.bottomWave}
                />

                <View style={styles.bottomContent}>
                  <View style={styles.rowBetween}>
                    <Text style={styles.label}>
                      Издатель: <Text style={styles.value}>{event.publisher}</Text>
                    </Text>
                  </View>

                  <Text style={styles.label}>
                      Год издания: <Text style={styles.value}>{event.publicationDate}</Text>
                  </Text>

                  <Text style={styles.label}>
                    Возрастное ограничение:{' '}
                    <Text style={styles.value}>{event.ageRating}</Text>
                  </Text>

                  <TouchableOpacity style={styles.joinButton} onPress={() => setToastVisible(true)}>
                    <Text style={styles.joinButtonText}>Присоединиться</Text>
                  </TouchableOpacity>
                </View>
              <TouchableOpacity style={styles.closeButton} onPress={onClose}>
                <Image tintColor={Colors.black} style={{width: 35, height: 35, resizeMode: 'cover'}} source={require('../assets/images/event_card/back.png')}/>
              </TouchableOpacity>
              
              {toastVisible && <Toast
                    visible={toastVisible}
                    type="success"
                    title="Успешно!"
                    message={`Вы зарегистрированы на событие "${event.title}" ${event.date}`}
                    onHide={() => setToastVisible(false)}
              />}
              
            </View>              
          </TouchableWithoutFeedback>
        </View>
      </TouchableWithoutFeedback>
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
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: Colors.white,
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
    bottom: -25,
    left: 0,
    width: 300,
    height: 65,
    resizeMode: 'stretch'
  },
  title: {
    position: 'absolute',
    bottom: 0,
    left: 16,
    fontFamily: inter.black,
    fontSize: 20,
    color: Colors.black,
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
    marginTop: 6
  },
  value: { fontFamily: inter.regular, fontSize: 14, color: Colors.black, marginTop: 6 },
  genres: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginVertical: 6,
  },
  genreBadge: { marginRight: 6, backgroundColor: Colors.lightBlue, borderRadius: 20, paddingHorizontal: 8, paddingVertical: 2 },
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
  inline: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  bottomWave: {
    position: 'absolute',
    bottom: -3,
    left: 0,
    width: '100%',
    height: 160,
    resizeMode: 'stretch',
  },
  bottomContent: {
    position: 'absolute',
    bottom: 14,
    left: 16,
    right: 16,
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
