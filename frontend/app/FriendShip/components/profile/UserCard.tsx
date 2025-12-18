import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
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

const PLACEHOLDER_IMAGE = 'https://via.placeholder.com/150/CCCCCC/FFFFFF?text=User';

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
  const colors = useThemedColors();
  const getImageSource = () => {
    if (!imageUri || (typeof imageUri === 'string' && imageUri.trim() === '')) {
      return { uri: PLACEHOLDER_IMAGE };
    }
    return typeof imageUri === 'string' ? { uri: imageUri } : imageUri;
  };

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
        <Text style={{backgroundColor: colors.lightBlue2}}>{highlighted.match}</Text>
        {highlighted.after}
      </Text>
    );
  };

  return (
    <View style={styles.shadowWrapper}>
      <TouchableOpacity 
        style={[styles.card, {backgroundColor: colors.card}]}
        onPress={onPress}
        activeOpacity={0.7}
      >
        <Image 
          source={getImageSource()}
          style={styles.userImage}
          defaultSource={{ uri: PLACEHOLDER_IMAGE }}
        />
        
        <View style={styles.content}>
          {renderHighlightedText(
            name,
            highlightedName,
            [styles.userName, {color: colors.black}]
          )}
          
          {renderHighlightedText(
            username,
            highlightedUsername,
            [styles.userUsername, {color: colors.black}]
          )}
          
          {description && renderHighlightedText(
            description,
            highlightedDescription,
            [styles.description, {color: colors.black}]
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
    marginBottom: 2,
  },
  userUsername: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    marginBottom: 6,
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    lineHeight: 18,
  },
});

export default UserCard;