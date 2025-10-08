import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { formatTitle } from "../utils/formatTitle";

export interface Event {
  id: string;
  title: string;
  date: string;
  genres: string[];
  currentParticipants: number;
  maxParticipants: number;
  duration: string;
  imageUri: string;
  description: string;
  typeEvent: string;
  typePlace: 'online' | 'offline';
  eventPlace: string;
  publisher: string;
  publicationDate: string;
  ageRating: string;
  category: 'movie' | 'game' | 'table_game' | 'other';
  onPress?: () => void;
  highlightedTitle?: {
    before: string;
    match: string;
    after: string;
  };
}

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

const getVisibleGenres = (genres: string[], maxBadges = 3, maxChars = 20) => {
  const visible: string[] = [];
  let totalChars = 0;

  for (let i = 0; i < genres.length; i++) {
    const g = genres[i];
    totalChars += g.length;

    if (visible.length >= maxBadges || totalChars > maxChars) {
      visible.push("…");
      break;
    }

    visible.push(g);
  }

  return visible;
};

const EventCard: React.FC<Event> = ({
  title,
  date,
  genres,
  currentParticipants,
  maxParticipants,
  duration,
  imageUri,
  onPress,
  category,
  typePlace,
  highlightedTitle,
}) => {
  const renderTitle = () => {
    if (highlightedTitle) {
      return (
        <Text style={styles.title}>
          {highlightedTitle.before}
          <Text style={styles.highlightedText}>{highlightedTitle.match}</Text>
          {highlightedTitle.after}
        </Text>
      );
    } else {
      return <Text
        style={styles.title}
        numberOfLines={1}
        ellipsizeMode="tail"
      >
        {formatTitle(title)}
      </Text>
      ;
    }
  };

  return (
    <TouchableOpacity onPress={onPress} style={styles.card}>
      <View style={styles.imageWrapper}>
        <Image source={{ uri: imageUri }} style={styles.image} />
        <View style={styles.dateBadge}>
          <Text style={styles.dateText}>{date}</Text>
        </View>
      </View>

      <View style={styles.content}>
        <View style={{flexDirection: 'row', justifyContent: 'space-between'}}>
          {renderTitle()}
          <View style={{flexDirection: 'row', justifyContent: 'space-between'}}>
            <View style={styles.iconOverlay}>
                <Image
                  source={categoryIcons[category]}
                  style={{ resizeMode: 'contain', width: 16, height: 16 }}
                />
            </View>
            <View style={styles.iconOverlay}>
                <Image
                  source={placeIcons[typePlace]}
                  style={{ resizeMode: 'contain', width: 16, height: 16 }}
                />
            </View>
          </View>
        </View>
        <View style={styles.genres}>
          <Text style={[styles.metaText, { marginRight: 6, fontFamily: Montserrat.regular }]}>Жанры:</Text>
          {getVisibleGenres(genres).map((genre, index) => (
            <View key={index} style={styles.genreBadge}>
              <Text style={styles.genreText}>{genre}</Text>
            </View>
          ))}
        </View>

        <View style={styles.metaContainer}>
          <View style={styles.metaRow}>
            <Text style={styles.metaText}>
              Участники: {currentParticipants}/{maxParticipants}
            </Text>
            <Image
              source={require("../assets/images/event_card/person2.png")}
              style={styles.metaIcon}
            />
          </View>

          <View style={styles.metaRow}>
            <Text style={styles.metaText}>{duration}</Text>
            <Image
              source={require("../assets/images/event_card/duration.png")}
              style={styles.metaIcon}
            />
          </View>
        </View>
      </View>
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  card: {
    width: 320,
    alignSelf: 'center',
    borderRadius: 26,
    backgroundColor: Colors.white,
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
    height: 130,
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
    borderColor: Colors.lightBlue3,
  },
  dateText: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    margin: -4,
    color: Colors.black,
  },
  iconOverlay: {
    width: 30,
    height: 30,
    backgroundColor: Colors.white,
    borderRadius: 20,
    borderWidth: 2,
    marginLeft: 6,
    borderColor: Colors.lightBlue3,
    alignItems: 'center',
    alignSelf: 'center',
    justifyContent: 'center'
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    maxWidth: '85%',
    marginTop: 4,
    marginBottom: 4,
    color: Colors.black,
  },
  highlightedText: {
    backgroundColor: Colors.lightBlue,
    color: Colors.blue,
    fontFamily: inter.black,
  },
  genres: { 
    flexDirection: 'row', 
    flexWrap: 'wrap', 
    marginBottom: 8 
  },
  genreBadge: {
    justifyContent: 'center',
    marginRight: 6,
    borderColor: Colors.lightBlue3,
    borderWidth: 2,
    borderRadius: 30,
    paddingHorizontal: 8,
  },
  genreText: { 
    fontFamily: Montserrat.regular, 
    color: Colors.black, 
    fontSize: 10 
  },
  metaContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  metaRow: { 
    flexDirection: 'row', 
    alignItems: 'center', 
    marginRight: 6 
  },
  metaIcon: {
    marginTop: -4,
    resizeMode: 'contain',
    width: 18,
    height: 18,
    marginLeft: 2,
  },
  metaText: { 
    fontFamily: Montserrat.regular, 
    fontSize: 14, 
    color: Colors.black 
  },
});

export default EventCard;