import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { Image, ImageSourcePropType, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

export interface User {
  id: string;
  name: string;
  username: string;
  description: string;
  imageUri: string | ImageSourcePropType;
  onPress?: () => void;
  highlightedName?: {
    before: string;
    match: string;
    after: string;
  };
  highlightedUsername?: {
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

const UserCard: React.FC<User> = ({
  name,
  username,
  description,
  imageUri,
  onPress,
  highlightedName,
  highlightedUsername,
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
        <Image 
          source={typeof imageUri === 'string' ? { uri: imageUri } : imageUri} 
          style={styles.userImage}
        />
        
        <View style={styles.content}>
          {renderHighlightedText(
            name,
            highlightedName,
            styles.userName
          )}
          
          {renderHighlightedText(
            username,
            highlightedUsername,
            styles.userUsername
          )}
          
          {renderHighlightedText(
            description,
            highlightedDescription,
            styles.description
          )}
        </View>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  shadowWrapper: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 6,
  },
  card: {
    width: '100%',
    maxWidth: 500,
    alignSelf: 'center',
    flexDirection: 'row',
    borderRadius: 20,
    backgroundColor: Colors.white,
    padding: 16,
    elevation: 4,
    alignItems: 'center',
  },
  userImage: {
    width: 90,
    height: 90,
    borderRadius: 45,
    marginRight: 16,
    resizeMode: 'cover',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
  },
  userName: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.black,
    marginBottom: 2,
  },
  userUsername: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    marginBottom: 6,
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    lineHeight: 18,
  },
  highlight: {
    backgroundColor: Colors.lightBlue,
    color: Colors.black,
  },
});

export default UserCard;