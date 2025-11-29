import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import * as ImagePicker from 'expo-image-picker';
import React, { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Image,
  ImageBackground,
  ImageSourcePropType,
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

  const handleSave = async () => {
    if (!name.trim()) {
      Alert.alert('Ошибка', 'Введите имя');
      return;
    }

    if (!username.trim() || username.length < 2) {
      Alert.alert('Ошибка', 'Юзернейм должен содержать минимум 2 символа');
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

  const handleAvatarPress = async () => {
    try {
      const permissionResult = await ImagePicker.requestMediaLibraryPermissionsAsync();
      
      if (!permissionResult.granted) {
        Alert.alert(
          'Требуется разрешение',
          'Для загрузки фото необходимо разрешение на доступ к галерее'
        );
        return;
      }

      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: ['images'],
        allowsEditing: true,
        aspect: [1, 1],
        quality: 0.8,
      });

      if (!result.canceled && result.assets[0]) {
        setAvatar({ uri: result.assets[0].uri });
      }
    } catch (error) {
      console.error('[EditProfileModal] Ошибка выбора изображения:', error);
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
        <View style={styles.modal}>
          <ScrollView
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
            nestedScrollEnabled
            bounces={false}
            alwaysBounceVertical={false}
          >
            <View style={styles.header}>
              <Text style={styles.title}>Редактирование профиля</Text>
              <TouchableOpacity 
                style={styles.closeButton} 
                onPress={handleClose}
                disabled={loading}
              >
                <Image
                  tintColor={Colors.black}
                  style={{ width: 35, height: 35, resizeMode: 'cover' }}
                  source={require('@/assets/images/event_card/back.png')}
                />
              </TouchableOpacity>
            </View>

            <View style={styles.content}>
              <TouchableOpacity 
                style={styles.avatarContainer}
                onPress={handleAvatarPress}
                disabled={loading}
              >
                <Image source={getAvatarSource()} style={styles.avatar} />
              </TouchableOpacity>

              <TextInput
                style={styles.input}
                placeholder="Имя"
                placeholderTextColor={Colors.grey}
                value={name}
                onChangeText={setName}
                maxLength={50}
                editable={!loading}
              />

              <TextInput
                style={styles.input}
                placeholder="Юзернейм"
                placeholderTextColor={Colors.grey}
                value={username}
                onChangeText={setUsername}
                maxLength={30}
                editable={!loading}
                autoCapitalize="none"
              />

              <TextInput
                style={[styles.input, styles.textArea]}
                placeholder="Описание"
                placeholderTextColor={Colors.grey}
                value={description}
                onChangeText={setDescription}
                multiline
                numberOfLines={3}
                textAlignVertical="top"
                maxLength={40}
                editable={!loading}
              />
            </View>

            <ImageBackground
              source={require('@/assets/images/event_card/bottom_rectangle.png')}
              style={styles.bottomBackground}
              resizeMode="stretch"
              tintColor={Colors.lightBlue}
            >
              <View style={styles.bottomContent}>
                <TouchableOpacity
                  style={[styles.saveButton, loading && styles.saveButtonDisabled]}
                  onPress={handleSave}
                  disabled={loading}
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
    backgroundColor: Colors.white,
    maxHeight: screenHeight * 0.85,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 14,
    paddingHorizontal: 16,
    backgroundColor: Colors.white,
    position: 'relative',
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
  },
  closeButton: {
    position: 'absolute',
    right: 10,
    zIndex: 1,
  },
  content: {
    backgroundColor: Colors.white,
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
    borderColor: Colors.lightGrey
  },
  avatar: {
    width: 150,
    height: 150,
    borderRadius: 75,
    borderColor: Colors.lightGrey,
  },
  input: {
    width: '100%',
    borderBottomWidth: 1,
    borderBottomColor: Colors.grey,
    paddingVertical: 8,
    paddingHorizontal: 0,
    marginBottom: 16,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
  },
  textArea: {
    borderWidth: 1,
    borderColor: Colors.grey,
    borderRadius: 8,
    padding: 16,
    minHeight: 60,
    maxHeight: 120,
  },
  bottomBackground: {
    width: "100%",
    marginTop: -1
  },
  bottomContent: {
    padding: 16,
  },
  saveButton: {
    backgroundColor: Colors.white,
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