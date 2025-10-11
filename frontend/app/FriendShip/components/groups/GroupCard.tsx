import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

export interface Group {
  id: string;
  name: string;
  participantsCount: number;
  description: string;
  imageUri: string | any;
  onPress?: () => void;
  highlightedName?: {
    before: string;
    match: string;
    after: string;
  };
  highlightedDescription?: {
    before: string;
    match: string;
    after: string;
  };
}

interface GroupCardProps extends Group {
  actionText?: string;
  actionColor?: string[];
}

const GroupCard: React.FC<GroupCardProps> = ({
  name,
  participantsCount,
  description,
  imageUri,
  onPress,
  highlightedName,
  highlightedDescription,
}) => {
  const renderHighlightedText = (
    normalText: string,
    highlighted: { before: string; match: string; after: string } | undefined,
    style: any
  ) => {
    if (!highlighted) {
      return <Text style={style}>{normalText}</Text>;
    }

    return (
      <Text style={style}>
        {highlighted.before}
        <Text style={styles.highlight}>{highlighted.match}</Text>
        {highlighted.after}
      </Text>
    );
  };

  return (
    <View style={styles.shadowWrapper}>
      <TouchableOpacity 
        style={styles.card}
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
              {renderHighlightedText(
                name,
                highlightedName,
                [styles.groupName, { numberOfLines: 2, ellipsizeMode: 'tail' }]
              )}
              
              <View style={styles.iconsContainer}>
                <Image
                  source={require('@/assets/images/event_card/movie.png')}
                  style={styles.icon}
                />
                <Image
                  source={require('@/assets/images/event_card/game.png')}
                  style={styles.icon}
                />
                <Image
                  source={require('@/assets/images/event_card/table_game.png')}
                  style={styles.icon}
                />
              </View>
              
              {renderHighlightedText(
                description,
                highlightedDescription,
                [styles.description, { numberOfLines: 3, ellipsizeMode: 'tail' }]
              )}
              
              <View style={styles.participantsRow}>
                <Text style={styles.participantsText}>
                  Участники:  {participantsCount}
                </Text>
                <Image
                  source={require('@/assets/images/event_card/person2.png')}
                  style={styles.personIcon}
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
    width: 360,
    alignSelf: 'center',
    borderRadius: 20,
    backgroundColor: Colors.white,
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
    width: 130,
    height: 130,
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
    color: Colors.black,
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
    color: Colors.black,
    marginRight: 2,
  },
  personIcon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
  },
  description: {
    width: 200,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    lineHeight: 18,
    marginBottom: 8,
    flex: 1,
  },
  highlight: {
    backgroundColor: Colors.lightBlue,
    color: Colors.black,
  },
});

export default GroupCard;