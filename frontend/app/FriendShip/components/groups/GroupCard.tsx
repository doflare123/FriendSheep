import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

export type GroupCategory = 'movie' | 'game' | 'table_game' | 'other';

export interface Group {
  id: string;
  name: string;
  participantsCount: number;
  description: string;
  imageUri: string | any;
  categories: GroupCategory[];
  onPress?: () => void;
  highlightedName?: {
    before: string;
    match: string;
    after: string;
  };
  isPrivate?: boolean;
}

interface GroupCardProps extends Group {
  actionText?: string;
  actionColor?: string[];
}

const categoryIcons: Record<GroupCategory, any> = {
  movie: require('@/assets/images/event_card/movie.png'),
  game: require('@/assets/images/event_card/game.png'),
  table_game: require('@/assets/images/event_card/table_game.png'),
  other: require('@/assets/images/event_card/other.png'),
};

const GroupCard: React.FC<GroupCardProps> = ({
  name,
  participantsCount,
  description,
  imageUri,
  categories = [],
  onPress,
  highlightedName,
}) => {
  const colors = useThemedColors();

  const renderHighlightedText = () => {
    if (!highlightedName) {
      return (
        <Text 
          style={[styles.groupName, { color: colors.black }]}
          numberOfLines={2}
          ellipsizeMode='tail'
        >
          {name}
        </Text>
      );
    }

    return (
      <Text 
        style={[styles.groupName, { color: colors.black }]}
        numberOfLines={2}
        ellipsizeMode='tail'
      >
        {highlightedName.before}
        <Text style={{ backgroundColor: colors.lightBlue, color: colors.black }}>
          {highlightedName.match}
        </Text>
        {highlightedName.after}
      </Text>
    );
  };

  console.log('GroupCard получил:', {
    name,
    imageUri,
    categories,
    imageType: typeof imageUri,
    highlightedName,
  });

  return (
    <View style={styles.shadowWrapper}>
      <TouchableOpacity 
        style={[styles.card, { backgroundColor: colors.card }]}
        onPress={onPress}
        activeOpacity={0.7}
      >
        <View style={styles.header}>
          <View style={styles.content}>
            <Image 
              source={typeof imageUri === 'string' ? { uri: imageUri } : imageUri} 
              style={styles.groupImage}
            />
          </View>
          <View style={styles.content}>
            <View style={styles.headerContent}>
              {renderHighlightedText()}
              
              {categories.length > 0 && (
                <View style={styles.iconsContainer}>
                  {categories.map((category, index) => (
                    <Image
                      key={`${category}-${index}`}
                      source={categoryIcons[category]}
                      style={[styles.icon, {tintColor: colors.black}]}
                    />
                  ))}
                </View>
              )}
              
              <Text 
                style={[styles.description, { color: colors.black }]} 
                numberOfLines={3} 
                ellipsizeMode='tail'
              >
                {description}
              </Text>
              
              <View style={styles.participantsRow}>
                <Text style={[styles.participantsText, { color: colors.black }]}>
                  Участники:  {participantsCount}
                </Text>
                <Image
                  source={require('@/assets/images/event_card/person.png')}
                  style={[styles.personIcon, {tintColor: colors.black}]}
                />
              </View>
            </View>
          </View> 
        </View>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  card: {
    width: 340,
    alignSelf: 'center',
    borderRadius: 20,
    overflow: 'hidden',
    elevation: 4,
    padding: 8,
  },
  shadowWrapper:{
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.1,
    shadowRadius: 3,
  },
  header: {
    flexDirection: 'row',
    marginBottom: 12,
  },
  content:{
    flexDirection: 'column',
    justifyContent: 'center',
  },
  groupImage: {
    width: 110,
    height: 110,
    borderRadius: 100,
    resizeMode: 'contain',
    marginRight: 12,
  },
  headerContent: {
    flex: 1,
    justifyContent: 'flex-start',
  },
  groupName: {
    width: 200,
    fontFamily: Montserrat.bold,
    fontSize: 18,
  },
  iconsContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8
  },
  icon: {
    width: 16,
    height: 16,
    resizeMode: 'contain',
    marginRight: 8,
  },
  participantsRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 'auto',
  },
  participantsText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    marginRight: 2,
  },
  personIcon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
    marginTop: -4
  },
  description: {
    width: 200,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    lineHeight: 18,
    marginBottom: 8,
    flex: 1,
  },
});

export default GroupCard;