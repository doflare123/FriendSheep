import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { LinearGradient } from 'expo-linear-gradient';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

export interface Event {
  id: string;
  title: string;
  date: string;
  genres: string[];
  participants: string;
  duration: string;
  imageUri: string;
  description: string;
  typeEvent: string;
  typePlace: string;
  eventPlace: string;
  publisher: string;
  publicationDate: number;
  ageRating: string;
  onPress?: () => void;
}

const EventCard: React.FC<Event> = ({ title, date, genres, participants, duration, imageUri, onPress }) => {
  return (
    <View style={styles.card}>
      <View style={styles.imageWrapper}>
        <Image source={{ uri: imageUri }} style={styles.image} />
        <View style={styles.dateBadge}>
          <Text style={styles.dateText}>{date}</Text>
        </View>
        <View style={styles.iconOverlay}>
          <Image source={require("../assets/images/event_card/movie.png")} style={{ resizeMode: 'contain', width: 25, height: 25 }} />
        </View>
      </View>

      <View style={styles.content}>
        <Text style={styles.title}>{title}</Text>
        <View style={styles.genres}>
          <Text style={[styles.metaText, { marginRight: 6 }]}>Жанры:</Text>
          {genres.map((genre) => (
            <View key={genre} style={styles.genreBadge}>
              <Text style={styles.genreText}>{genre}</Text>
            </View>
          ))}
        </View>

        <View style={styles.metaContainer}>
          <View style={styles.metaRow}>
            <Text style={styles.metaText}>Участников: {participants}</Text>
            <Image source={require("../assets/images/event_card/person.png")} style={styles.metaIcon} />
          </View>

          <View style={styles.metaRow}>
            <Text style={styles.metaText}>{duration}</Text>
            <Image source={require("../assets/images/event_card/duration.png")} style={styles.metaIcon} />
          </View>
        </View>

        <View style={styles.iconOverlay}>
          <Image source={require("../assets/images/event_card/online.png")} style={{ resizeMode: 'contain', width: 20, height: 20 }} />
        </View>

        <TouchableOpacity onPress={onPress}>
          <LinearGradient
            colors={[Colors.lightBlue, Colors.blue]}
            start={{ x: 0, y: 0 }}
            end={{ x: 0, y: 1 }}
            style={styles.button}
          >
            <Text style={styles.buttonText}>Узнать подробнее</Text>
          </LinearGradient>
        </TouchableOpacity>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  card: {
    width: 320,
    alignSelf: 'center',
    borderRadius: 16,
    backgroundColor: Colors.white,
    borderColor: Colors.blue,
    borderWidth: 4,
    overflow: 'hidden',
    elevation: 4,
    minHeight: 200,
  },
  imageWrapper: {
    minHeight: 100,
    height: 120,
  },
  image: {
    width: '100%',
    height: 120,
    resizeMode: 'cover',
  },
  content: {
    minHeight: 120,
    padding: 12,
  },
  dateBadge: {
    position: 'absolute',
    top: 10,
    left: 10,
    backgroundColor: Colors.white,
    borderRadius: 12,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
  },
  dateText: { fontFamily: inter.regular, fontSize: 12, margin: -4, color: Colors.black },
  iconOverlay: {
    position: 'absolute',
    top: 10,
    right: 10,
    backgroundColor: Colors.white,
    borderRadius: 20,
    padding: 6,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
  },
  title: { fontFamily: inter.black, fontSize: 17, marginTop: -4, marginBottom: 8 },
  genres: { flexDirection: 'row', flexWrap: 'wrap', marginBottom: 8 },
  genreBadge: { marginRight: 6, backgroundColor: Colors.lightBlue, borderRadius: 20, paddingHorizontal: 8, paddingVertical: 2 },
  genreText: { fontFamily: inter.regular, color: Colors.black, fontSize: 10 },
  metaContainer: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 8 },
  metaRow: { flexDirection: 'row', alignItems: 'center', marginRight: 6 },
  metaIcon: { resizeMode: 'contain', width: 16, height: 16, marginLeft: 6 },
  metaText: { fontFamily: inter.regular, fontSize: 12, color: Colors.black },
  button: { marginTop: 4, backgroundColor: Colors.lightBlue, borderRadius: 20, paddingVertical: 4, alignItems: 'center' },
  buttonText: { fontFamily: inter.regular, color: Colors.white, fontSize: 16 },
});

export default EventCard;
