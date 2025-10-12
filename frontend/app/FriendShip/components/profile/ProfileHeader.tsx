import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import {
    Image,
    ImageSourcePropType,
    Linking,
    StyleSheet,
    Text,
    TouchableOpacity,
    View
} from 'react-native';

interface ProfileHeaderProps {
  avatar: ImageSourcePropType;
  name: string;
  username: string;
  description: string;
  registrationDate: string;
  telegramLink?: string;
}

const ProfileHeader: React.FC<ProfileHeaderProps> = ({
  avatar,
  name,
  username,
  description,
  registrationDate,
  telegramLink,
}) => {
  const handleTelegramPress = () => {
    if (telegramLink) {
      Linking.openURL(telegramLink);
    }
  };

  return (
    <View style={styles.header}>
      <Image source={avatar} style={styles.profileImage} />
      <View style={styles.headerInfo}>
        <View style={styles.nameRow}>
          <Text style={styles.profileName}>{name}</Text>
          {telegramLink && (
            <TouchableOpacity onPress={handleTelegramPress}>
              <Image 
                source={require('@/assets/images/profile/telegram.png')} 
                style={styles.telegramIcon}
              />
            </TouchableOpacity>
          )}
        </View>
        <Text style={styles.profileUs}>{username}</Text>
        <Text style={styles.description}>{description}</Text>
        <Text style={styles.registrationDate}>{registrationDate}</Text>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  header: {
    marginVertical: -10,
    flexDirection: 'row',
    padding: 20,
    alignItems: 'flex-start',
  },
  headerInfo: {
    flex: 1,
  },
  profileImage: {
    width: 120,
    height: 120,
    borderRadius: 100,
    marginRight: 16,
  },
  nameRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  profileName: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.black,
    marginRight: 4,
  },
  telegramIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
  },
  profileUs: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    marginTop: -6,
    marginBottom: 8,
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    marginBottom: 8,
  },
  registrationDate: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.lightGrey,
    alignSelf: 'flex-end',
  },
});

export default ProfileHeader;