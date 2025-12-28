import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useRef, useState } from 'react';
import {
  Image,
  ImageSourcePropType,
  Linking,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View
} from 'react-native';

interface ProfileHeaderProps {
  avatar: ImageSourcePropType;
  name: string;
  username: string;
  description: string;
  registrationDate: string;
  telegramLink?: string;
  isOwnProfile: boolean;
  isEnterprise?: boolean;
  onEditProfile?: () => void;
  onChangeTiles?: () => void;
  onInviteToGroup?: () => void;
}

const ProfileHeader: React.FC<ProfileHeaderProps> = ({
  avatar,
  name,
  username,
  description,
  registrationDate,
  telegramLink,
  isOwnProfile,
  isEnterprise,
  onEditProfile,
  onChangeTiles,
  onInviteToGroup,
}) => {
  const colors = useThemedColors();
  const [menuVisible, setMenuVisible] = useState(false);
  const [menuPosition, setMenuPosition] = useState({ top: 0, left: 0 });
  const [verifiedTooltipVisible, setVerifiedTooltipVisible] = useState(false);
  const settingsIconRef = useRef<View>(null);

  const handleTelegramPress = () => {
    if (telegramLink) {
      Linking.openURL(telegramLink);
    } else {
      Linking.openURL('https://t.me/FriendShipNotify_bot');
    }
  };

  const handleSettingsPress = () => {
    settingsIconRef.current?.measureInWindow((x, y, width, height) => {
      setMenuPosition({
        top: y,
        left: x,
      });
      setMenuVisible(true);
    });
  };

  const handleMenuItemPress = (action: 'edit' | 'tiles') => {
    setMenuVisible(false);
    if (action === 'edit' && onEditProfile) {
      onEditProfile();
    } else if (action === 'tiles' && onChangeTiles) {
      onChangeTiles();
    }
  };

  return (
    <>
      <View style={styles.header}>
        <View style={styles.avatarContainer}>
          <Image source={avatar} style={styles.profileImage} />
          
          {isOwnProfile && (
            <TouchableOpacity 
              ref={settingsIconRef}
              style={styles.settingsIcon}
              onPress={handleSettingsPress}
            >
              <Image 
                source={require('@/assets/images/profile/settings.png')} 
                style={[styles.settingsImage, {tintColor: Colors.lightGreyBlue}]}
              />
            </TouchableOpacity>
          )}

          {!isOwnProfile && (
            <TouchableOpacity 
              style={styles.inviteIcon}
              onPress={onInviteToGroup}
            >
              <Image 
                source={require('@/assets/images/profile/group_add.png')} 
                style={styles.inviteImage}
              />
            </TouchableOpacity>
          )}
        </View>
        
        <View style={styles.headerInfo}>
          <View style={styles.nameRow}>
            <Text style={[styles.profileName, { color: colors.black }]}>{name}</Text>
            {isEnterprise && (
              <TouchableWithoutFeedback onPress={() => setVerifiedTooltipVisible(true)}>
                <View>
                  <Image 
                    source={require('@/assets/images/profile/verified.png')} 
                    style={styles.verifiedIcon}
                  />
                </View>
              </TouchableWithoutFeedback>
            )}
            {isOwnProfile && (
              <TouchableOpacity onPress={handleTelegramPress}>
                <Image 
                  source={require('@/assets/images/profile/telegram.png')} 
                  style={[
                    styles.telegramIcon,
                    !telegramLink && styles.telegramIconInactive
                  ]}
                />
              </TouchableOpacity>
            )}
          </View>
          <Text style={[styles.profileUs, { color: colors.black }]}>{username}</Text>
          <Text style={[styles.description, { color: colors.black }]}>{description}</Text>
          <Text style={styles.registrationDate}>{registrationDate}</Text>
        </View>
      </View>

      {isOwnProfile && (
        <Modal
          visible={menuVisible}
          transparent
          animationType="fade"
          onRequestClose={() => setMenuVisible(false)}
        >
          <TouchableWithoutFeedback onPress={() => setMenuVisible(false)}>
            <View style={styles.modalOverlay}>
              <TouchableWithoutFeedback>
                <View 
                  style={[
                    styles.menuContainer,
                    { 
                      top: menuPosition.top, 
                      left: menuPosition.left,
                      backgroundColor: colors.white
                    }
                  ]}
                >
                  <TouchableOpacity
                    style={styles.menuItem}
                    onPress={() => handleMenuItemPress('tiles')}
                  >
                    <Text style={[styles.menuItemText, { color: colors.black }]}>Сменить плитки</Text>
                  </TouchableOpacity>
                  
                  <View style={styles.menuDivider} />
                  
                  <TouchableOpacity
                    style={styles.menuItem}
                    onPress={() => handleMenuItemPress('edit')}
                  >
                    <Text style={[styles.menuItemText, { color: colors.black }]} numberOfLines={2}>Редактировать профиль</Text>
                  </TouchableOpacity>
                </View>
              </TouchableWithoutFeedback>
            </View>
          </TouchableWithoutFeedback>
        </Modal>
      )}

      {isEnterprise && (
        <Modal
          visible={verifiedTooltipVisible}
          transparent
          animationType="fade"
          onRequestClose={() => setVerifiedTooltipVisible(false)}
        >
          <TouchableWithoutFeedback onPress={() => setVerifiedTooltipVisible(false)}>
            <View style={styles.tooltipOverlay}>
              <View style={[styles.tooltipContainer, { backgroundColor: colors.white }]}>
                <Image 
                  source={require('@/assets/images/profile/verified.png')} 
                  style={styles.tooltipIcon}
                />
                <Text style={[styles.tooltipTitle, { color: colors.black }]}>
                  Верифицированный{'\n'}аккаунт
                </Text>
                <Text style={[styles.tooltipText, { color: colors.black }]}>
                  Этот аккаунт подтверждён администрацией FriendShip
                </Text>
                <TouchableOpacity 
                  style={styles.tooltipButton}
                  onPress={() => setVerifiedTooltipVisible(false)}
                >
                  <Text style={styles.tooltipButtonText}>Понятно</Text>
                </TouchableOpacity>
              </View>
            </View>
          </TouchableWithoutFeedback>
        </Modal>
      )}
    </>
  );
};

const styles = StyleSheet.create({
  header: {
    marginVertical: -10,
    flexDirection: 'row',
    padding: 20,
    alignItems: 'flex-start',
  },
  avatarContainer: {
    position: 'relative',
    marginRight: 16,
  },
  profileImage: {
    width: 120,
    height: 120,
    borderRadius: 100,
  },
  settingsIcon: {
    position: 'absolute',
    bottom: 0,
    right: 0,
  },
  settingsImage: {
    width: 30,
    height: 30,
  },
  inviteIcon: {
    position: 'absolute',
    bottom: 0,
    right: 0,
  },
  inviteImage: {
    width: 35,
    height: 35,
  },
  headerInfo: {
    flex: 1,
  },
  nameRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  profileName: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    marginRight: 4,
  },
  verifiedIcon: {
    width: 20,
    height: 20,
    resizeMode: 'contain',
  },
  telegramIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
  },
  telegramIconInactive: {
    tintColor: Colors.lightGrey,
    opacity: 0.5,
  },
  profileUs: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    marginTop: -6,
    marginBottom: 8,
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    marginBottom: 8,
  },
  registrationDate: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.lightGrey,
    alignSelf: 'flex-end',
  },
  modalOverlay: {
    flex: 1,
  },
  menuContainer: {
    position: 'absolute',
    borderRadius: 12,
    borderTopLeftRadius: 0,
    paddingVertical: 8,
    minWidth: 100,
    elevation: 5,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 3.84,
  },
  menuItem: {
    paddingVertical: 8,
    paddingHorizontal: 16,
    alignItems: 'center'
  },
  menuItemText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  menuDivider: {
    height: 1,
    backgroundColor: Colors.veryLightGrey,
    marginHorizontal: 8,
  },
  tooltipOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  tooltipContainer: {
    borderRadius: 16,
    padding: 24,
    marginHorizontal: 32,
    alignItems: 'center',
    elevation: 5,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 3.84,
  },
  tooltipIcon: {
    width: 48,
    height: 48,
    marginBottom: 16,
    resizeMode: 'contain',
  },
  tooltipTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    marginBottom: 8,
    textAlign: 'center',
  },
  tooltipText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    marginBottom: 20,
    textAlign: 'center',
    lineHeight: 20,
  },
  tooltipButton: {
    backgroundColor: Colors.lightBlue,
    paddingVertical: 8,
    paddingHorizontal: 24,
    borderRadius: 20,
  },
  tooltipButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
});

export default ProfileHeader;