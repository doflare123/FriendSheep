import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { LinearGradient } from 'expo-linear-gradient';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

export interface Group {
  id: string;
  name: string;
  participantsCount: number;
  description: string;
  imageUri: string | any;
  onPress?: () => void;
}

interface GroupCardProps extends Group {
  actionText: string;
  actionColor?: string[];
}

const GroupCard: React.FC<GroupCardProps> = ({
  name,
  participantsCount,
  description,
  imageUri,
  onPress,
  actionText,
  actionColor = [Colors.lightBlue, Colors.blue],
}) => {
  return (
    <View style={styles.card}>
      <View style={styles.header}>
        <Image 
          source={typeof imageUri === 'string' ? { uri: imageUri } : imageUri} 
          style={styles.groupImage} 
        />
        <View style={styles.headerContent}>
          <Text style={styles.groupName} numberOfLines={1} ellipsizeMode="tail">
            {name}
          </Text>
          <View style={styles.iconsContainer}>
            <Image
              source={require('../assets/images/event_card/movie.png')}
              style={styles.icon}
            />
            <Image
              source={require('../assets/images/event_card/game.png')}
              style={styles.icon}
            />
            <Image
              source={require('../assets/images/event_card/table_game.png')}
              style={styles.icon}
            />
          </View>
          <View style={styles.participantsRow}>
            <Text style={styles.participantsText}>
              Участников: {participantsCount}
            </Text>
            <Image
              source={require('../assets/images/event_card/person.png')}
              style={styles.personIcon}
            />
          </View>
        </View>
      </View>

      <Text style={styles.description} numberOfLines={2} ellipsizeMode="tail">
        {description}
      </Text>

      <TouchableOpacity onPress={onPress}>
        <LinearGradient
          colors={[Colors.blue, Colors.lightBlue]}
          start={{ x: 0, y: 0 }}
          end={{ x: 0, y: 1 }}
          style={styles.button}
        >
          <Text style={styles.buttonText}>{actionText}</Text>
        </LinearGradient>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  card: {
    width: 320,
    alignSelf: 'center',
    borderRadius: 16,
    backgroundColor: Colors.white,
    borderColor: Colors.lightBlue,
    borderWidth: 3,
    overflow: 'hidden',
    elevation: 4,
    padding: 12,
    minHeight: 180,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  groupImage: {
    width: 80,
    height: 80,
    borderRadius: 40,
    resizeMode: 'cover',
    marginRight: 12,
  },
  headerContent: {
    flex: 1,
    justifyContent: 'space-between',
    height: 80,
  },
  groupName: {
    fontFamily: inter.black,
    fontSize: 18,
    color: Colors.black,
    marginBottom: 4,
  },
  iconsContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  icon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
    marginRight: 8,
  },
  participantsRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  participantsText: {
    fontFamily: inter.regular,
    fontSize: 16,
    color: Colors.black,
  },
  personIcon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
  },
  description: {
    fontFamily: inter.regular,
    fontSize: 14,
    color: Colors.black,
    lineHeight: 16,
    marginBottom: 16,
    flex: 1,
  },
  button: {
    backgroundColor: Colors.lightBlue,
    borderRadius: 20,
    paddingVertical: 4,
    alignItems: 'center',
    marginTop: 'auto',
  },
  buttonText: {
    fontFamily: inter.regular,
    color: Colors.white,
    fontSize: 18,
  },
});

export default GroupCard;