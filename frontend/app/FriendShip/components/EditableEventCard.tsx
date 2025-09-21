import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { LinearGradient } from 'expo-linear-gradient';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { formatTitle } from "../utils/formatTitle";
import { Event } from './EventCard';

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

interface EditableEventCardProps extends Event {
  onEdit: () => void;
}

const EditableEventCard: React.FC<EditableEventCardProps> = ({
  title,
  date,
  genres,
  currentParticipants,
  maxParticipants,
  duration,
  imageUri,
  category,
  typePlace,
  highlightedTitle,
  onEdit,
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
      return (
        <Text
          style={styles.title}
          numberOfLines={1}
          ellipsizeMode="tail"
        >
          {formatTitle(title)}
        </Text>
      );
    }
  };

  return (
    <View style={styles.card}>
      <View style={styles.imageWrapper}>
        <Image source={{ uri: imageUri }} style={styles.image} />
        <View style={styles.dateBadge}>
          <Text style={styles.dateText}>{date}</Text>
        </View>
        <View style={styles.iconOverlay}>
          <Image
            source={categoryIcons[category]}
            style={{ resizeMode: 'contain', width: 25, height: 25 }}
          />
        </View>
      </View>

      <View style={styles.content}>
        {renderTitle()}
        <View style={styles.genres}>
          <Text style={[styles.metaText, { marginRight: 6, fontFamily: inter.medium }]}>
            Жанры:
          </Text>
          {getVisibleGenres(genres).map((genre, index) => (
            <View key={index} style={styles.genreBadge}>
              <Text style={styles.genreText}>{genre}</Text>
            </View>
          ))}
        </View>

        <View style={styles.metaContainer}>
          <View style={styles.metaRow}>
            <Text style={styles.metaText}>
              Участников: {currentParticipants}/{maxParticipants}
            </Text>
            <Image
              source={require("../assets/images/event_card/person.png")}
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

        <View style={styles.iconOverlay}>
          <Image
            source={placeIcons[typePlace]}
            style={{ resizeMode: 'contain', width: 20, height: 20 }}
          />
        </View>

        <TouchableOpacity onPress={onEdit}>
          <LinearGradient
            colors={[Colors.lightBlue, Colors.blue]}
            start={{ x: 0, y: 0 }}
            end={{ x: 0, y: 1 }}
            style={styles.button}
          >
            <Text style={styles.buttonText}>Редактировать</Text>
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
  dateText: {
    fontFamily: inter.regular,
    fontSize: 12,
    margin: -4,
    color: Colors.black,
  },
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
  title: {
    fontFamily: inter.black,
    fontSize: 17,
    maxWidth: '85%',
    marginTop: -4,
    marginBottom: 8,
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
    marginRight: 6,
    backgroundColor: Colors.lightBlue,
    borderRadius: 20,
    paddingHorizontal: 8,
    paddingVertical: 2,
  },
  genreText: { 
    fontFamily: inter.regular, 
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
    resizeMode: 'contain',
    width: 20,
    height: 20,
    marginLeft: 2,
  },
  metaText: { 
    fontFamily: inter.regular, 
    fontSize: 12, 
    color: Colors.black 
  },
  button: {
    marginTop: 4,
    backgroundColor: Colors.lightBlue,
    borderRadius: 20,
    paddingVertical: 4,
    alignItems: 'center',
  },
  buttonText: {
    fontFamily: inter.regular,
    color: Colors.white,
    fontSize: 16,
  },
});

export default EditableEventCard;