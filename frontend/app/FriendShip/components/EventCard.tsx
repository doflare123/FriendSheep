import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

interface EventCardProps {
  title: string;
  date: string;
  genres: string[];
  participants: string;
  duration: string;
  imageUri: string;
  onPress?: () => void;
}

const EventCard: React.FC<EventCardProps> = ({ title, date, genres, participants, duration, imageUri, onPress }) => {
  return (
    <View style={styles.card}>
      <View style={styles.imageWrapper}>
        <Image source={{ uri: imageUri }} style={styles.image} />
        <View style={styles.dateBadge}>
          <Text style={styles.dateText}>{date}</Text>
        </View>
        <View style={styles.iconOverlay}>
          <Image source={require("../assets/images/event_card/movie.png")} style={{ resizeMode: 'contain', width: 20, height: 20 }} />
        </View>
      </View>

      <View style={styles.content}>
        <Text style={styles.title}>{title}</Text>
        <View style={styles.genres}>
          <Text style={[styles.metaText, { alignSelf: 'center' }]}>Жанры:</Text>
          {genres.map((genre) => (
            <View key={genre} style={styles.genreBadge}>
              <Text style={styles.genreText}>{genre}</Text>
            </View>
          ))}
        </View>

        <View style={styles.metaContainer}>
          <View style={styles.metaRow}>
            <Text style={styles.metaText}>{participants}</Text>
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

        <TouchableOpacity style={styles.button} onPress={onPress}>
          <Text style={styles.buttonText}>Узнать подробнее</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  card: {
    width: 300,
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
    top: 15,
    left: 15,
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
    top: 15,
    right: 15,
    backgroundColor: Colors.white,
    borderRadius: 20,
    padding: 6,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
  },
  title: { fontFamily: inter.black, fontSize: 18, marginTop: -4, marginBottom: 8 },
  genres: { flexDirection: 'row', flexWrap: 'wrap', marginBottom: 8 },
  genreBadge: { marginRight: 6, marginBottom: 6, backgroundColor: Colors.lightBlue, borderRadius: 20, paddingHorizontal: 8, paddingVertical: 2 },
  genreText: { color: Colors.black, fontSize: 10 },
  metaContainer: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 8 },
  metaRow: { flexDirection: 'row', alignItems: 'center', marginRight: 6 },
  metaIcon: { resizeMode: 'contain', width: 16, height: 16 },
  metaText: { fontSize: 14, color: Colors.black },
  button: { marginTop: 4, backgroundColor: Colors.lightBlue, borderRadius: 20, paddingVertical: 4, alignItems: 'center' },
  buttonText: { color: Colors.white, fontSize: 16 },
});

export default EventCard;
