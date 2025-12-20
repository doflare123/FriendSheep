import permissionsService from '@/api/services/permissionsService';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import { validateUserDisplayName, validateUserUsername } from '@/utils/validators';
import * as ImagePicker from 'expo-image-picker';
import React, { useEffect, useMemo, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Image,
  ImageBackground,
  ImageSourcePropType,
  Linking,
  Modal,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View
} from 'react-native';

const screenHeight = Dimensions.get("window").height;

interface EditProfileModalProps {
  visible: boolean;
  onClose: () => void;
  onSave?: (profileData: ProfileData) => void;
  currentProfile: {
    avatar: ImageSourcePropType;
    name: string;
    username: string;
    description: string;
  };
}

interface ProfileData {
  avatar: string | ImageSourcePropType;
  name: string;
  username: string;
  description: string;
}

const EditProfileModal: React.FC<EditProfileModalProps> = ({ 
  visible, 
  onClose, 
  onSave,
  currentProfile 
}) => {
  const colors = useThemedColors();
  const [avatar, setAvatar] = useState<string | ImageSourcePropType>(currentProfile.avatar);
  const [name, setName] = useState(currentProfile.name);
  const [username, setUsername] = useState(currentProfile.username);
  const [description, setDescription] = useState(currentProfile.description);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (visible) {
      setAvatar(currentProfile.avatar);
      setName(currentProfile.name);
      setUsername(currentProfile.username);
      setDescription(currentProfile.description);
    }
  }, [visible, currentProfile]);

  const nameValidation = useMemo(() => validateUserDisplayName(name), [name]);
  const usernameValidation = useMemo(() => validateUserUsername(username), [username]);

  const handleSave = async () => {
    if (!nameValidation.isValid) {
      Alert.alert('Ошибка', `Имя: ${nameValidation.missingRequirements.join(', ')}`);
      return;
    }

    if (!usernameValidation.isValid) {
      Alert.alert('Ошибка', `Юзернейм: ${usernameValidation.missingRequirements.join(', ')}`);
      return;
    }

    setLoading(true);

    try {
      const profileData: ProfileData = {
        avatar,
        name: name.trim(),
        username: username.trim(),
        description: description.trim(),
      };
      
      await onSave?.(profileData);
      onClose();
    } catch (error) {
      console.error('[EditProfileModal] Ошибка сохранения:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setAvatar(currentProfile.avatar);
    setName(currentProfile.name);
    setUsername(currentProfile.username);
    setDescription(currentProfile.description);
    onClose();
  };

  const showPermissionAlert = (onRetry: () => void) => {
    Alert.alert(
      'Разрешение на доступ к фото',
      'Для загрузки аватара необходимо разрешить доступ к галерее. Хотите предоставить разрешение?',
      [
        {
          text: 'Отмена',
          style: 'cancel',
        },
        {
          text: 'Настройки',
          onPress: () => {
            Linking.openSettings();
          },
        },
        {
          text: 'Разрешить',
          onPress: onRetry,
        },
      ]
    );
  };

  const handleAvatarPress = async () => {
    try {
      console.log('[EditProfileModal] 🖼️ Запрос на выбор аватара');

      const hasPermission = await permissionsService.checkMediaPermission();
      
      if (!hasPermission) {
        console.log('[EditProfileModal] ⚠️ Нет разрешения на медиатеку');

        const granted = await permissionsService.requestMediaPermission();
        
        if (!granted) {
          console.log('[EditProfileModal] ❌ Пользователь отклонил разрешение');
          showPermissionAlert(handleAvatarPress);
          return;
        }
      }

      console.log('[EditProfileModal] ✅ Разрешение получено, открываем галерею');

      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: ['images'],
        allowsEditing: true,
        aspect: [1, 1],
        quality: 0.8,
      });

      console.log('[EditProfileModal] 📦 Результат выбора:', result);

      if (!result.canceled && result.assets && result.assets[0]) {
        console.log('[EditProfileModal] ✅ Аватар выбран:', result.assets[0].uri);
        setAvatar({ uri: result.assets[0].uri });
      } else {
        console.log('[EditProfileModal] ℹ️ Пользователь отменил выбор аватара');
      }
    } catch (error) {
      console.error('[EditProfileModal] ❌ Ошибка выбора изображения:', error);
      Alert.alert('Ошибка', 'Не удалось загрузить изображение');
    }
  };

  const getAvatarSource = (): ImageSourcePropType => {
    if (typeof avatar === 'string') {
      return { uri: avatar };
    }
    return avatar;
  };

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <View style={styles.overlay}>
        <View style={[styles.modal, { backgroundColor: colors.white }]}>
          <ScrollView
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
            nestedScrollEnabled
            bounces={false}
            alwaysBounceVertical={false}
          >
            <View style={[styles.header, { backgroundColor: colors.white }]}>
              <Text style={[styles.title, { color: colors.black }]}>Редактирование профиля</Text>
              <TouchableOpacity 
                style={styles.closeButton} 
                onPress={handleClose}
                disabled={loading}
              >
                <Image
                  tintColor={colors.black}
                  style={{ width: 35, height: 35, resizeMode: 'cover' }}
                  source={require('@/assets/images/event_card/back.png')}
                />
              </TouchableOpacity>
            </View>

            <View style={[styles.content, { backgroundColor: colors.white }]}>
              <TouchableOpacity 
                style={[styles.avatarContainer, { borderColor: colors.lightGrey }]}
                onPress={handleAvatarPress}
                disabled={loading}
              >
                <Image source={getAvatarSource()} style={[styles.avatar, { borderColor: colors.lightGrey }]} />
              </TouchableOpacity>

              <View style={styles.inputWrapper}>
                <View style={styles.labelRow}>
                  <Text style={[styles.label, { color: colors.black }]}>Имя</Text>
                  {name.length > 0 && (
                    <Text style={[
                      styles.charCount,
                      { color: nameValidation.length > nameValidation.maxLength ? Colors.red : colors.grey }
                    ]}>
                      {nameValidation.length}/{nameValidation.maxLength}
                    </Text>
                  )}
                </View>
                <TextInput
                  style={[
                    styles.input,
                    { 
                      borderBottomColor: colors.grey,
                      color: colors.black
                    },
                    name.length > 0 && !nameValidation.isValid && styles.inputError
                  ]}
                  placeholder="Имя"
                  placeholderTextColor={colors.grey}
                  value={name}
                  onChangeText={setName}
                  maxLength={50}
                  editable={!loading}
                />
                {name.length > 0 && !nameValidation.isValid && (
                  <Text style={styles.errorText}>
                    {nameValidation.missingRequirements.join(', ')}
                  </Text>
                )}
              </View>

              <View style={styles.inputWrapper}>
                <View style={styles.labelRow}>
                  <Text style={[styles.label, { color: colors.black }]}>Юзернейм</Text>
                  {username.length > 0 && (
                    <Text style={[
                      styles.charCount,
                      { color: usernameValidation.length > usernameValidation.maxLength ? Colors.red : colors.grey }
                    ]}>
                      {usernameValidation.length}/{usernameValidation.maxLength}
                    </Text>
                  )}
                </View>
                <TextInput
                  style={[
                    styles.input,
                    { 
                      borderBottomColor: colors.grey,
                      color: colors.black
                    },
                    username.length > 0 && !usernameValidation.isValid && styles.inputError
                  ]}
                  placeholder="Юзернейм"
                  placeholderTextColor={colors.grey}
                  value={username}
                  onChangeText={setUsername}
                  maxLength={30}
                  editable={!loading}
                  autoCapitalize="none"
                />
                {username.length > 0 && !usernameValidation.isValid && (
                  <Text style={styles.errorText}>
                    {usernameValidation.missingRequirements.join(', ')}
                  </Text>
                )}
              </View>

              <View style={styles.inputWrapper}>
                <Text style={[styles.label, { color: colors.black }]}>Описание</Text>
                <TextInput
                  style={[
                    styles.input, 
                    styles.textArea,
                    {
                      borderColor: colors.grey,
                      color: colors.black
                    }
                  ]}
                  placeholder="Описание"
                  placeholderTextColor={colors.grey}
                  value={description}
                  onChangeText={setDescription}
                  multiline
                  numberOfLines={3}
                  textAlignVertical="top"
                  maxLength={40}
                  editable={!loading}
                />
              </View>
            </View>

            <ImageBackground
              source={require('@/assets/images/event_card/bottom_rectangle.png')}
              style={styles.bottomBackground}
              resizeMode="stretch"
              tintColor={Colors.lightBlue}
            >
              <View style={styles.bottomContent}>
                <TouchableOpacity
                  style={[
                    styles.saveButton,
                    { backgroundColor: Colors.white },
                    (loading || !nameValidation.isValid || !usernameValidation.isValid) && styles.saveButtonDisabled
                  ]}
                  onPress={handleSave}
                  disabled={loading || !nameValidation.isValid || !usernameValidation.isValid}
                >
                  {loading ? (
                    <ActivityIndicator color={Colors.blue3} />
                  ) : (
                    <Text style={styles.saveButtonText}>Сохранить</Text>
                  )}
                </TouchableOpacity>
              </View>
            </ImageBackground>
          </ScrollView>
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
  },
  modal: {
    marginHorizontal: 12,
    borderRadius: 25,
    overflow: "hidden",
    maxHeight: screenHeight * 0.85,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 14,
    paddingHorizontal: 16,
    position: 'relative',
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    textAlign: 'center',
  },
  closeButton: {
    position: 'absolute',
    right: 10,
    zIndex: 1,
  },
  content: {
    paddingHorizontal: 16,
    paddingBottom: 16,
    alignItems: 'center',
  },
  avatarContainer: {
    marginBottom: 8,
    position: 'relative',
    alignItems: 'center',
    width: 156,
    height: 156,
    borderRadius: 78,
    borderWidth: 3,
    borderStyle: 'dashed',
  },
  avatar: {
    width: 150,
    height: 150,
    borderRadius: 75,
  },
  inputWrapper: {
    width: '100%',
    marginBottom: 16,
  },
  labelRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  label: {
    fontFamily: Montserrat.medium,
    fontSize: 14,
  },
  charCount: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
  },
  input: {
    width: '100%',
    borderBottomWidth: 1,
    paddingVertical: 8,
    paddingHorizontal: 0,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  inputError: {
    borderBottomColor: Colors.red,
  },
  textArea: {
    borderWidth: 1,
    borderRadius: 8,
    padding: 16,
    minHeight: 60,
    maxHeight: 120,
  },
  errorText: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.red,
    marginTop: 4,
  },
  bottomBackground: {
    width: "100%",
    marginTop: -2
  },
  bottomContent: {
    padding: 16,
  },
  saveButton: {
    marginHorizontal: 60,
    paddingVertical: 10,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 40,
  },
  saveButtonDisabled: {
    opacity: 0.6,
  },
  saveButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.blue3,
  },
});

export default EditProfileModal;