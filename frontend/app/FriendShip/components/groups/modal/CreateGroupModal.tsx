import groupService from '@/api/services/group/groupService';
import permissionsService from '@/api/services/permissionsService';
import DescriptionModal from '@/components/event/modal/DescriptionModal';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import profanityFilter from '@/utils/profanityFilter';
import { validateFullDescription, validateGroupName, validateShortDescription } from '@/utils/validators';
import * as ImagePicker from 'expo-image-picker';
import React, { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  BackHandler,
  Dimensions,
  Image,
  ImageBackground,
  Linking,
  Modal,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View
} from 'react-native';
import ContactsModal, { Contact } from './ContactsModal';

const screenHeight = Dimensions.get("window").height;

interface CreateGroupModalProps {
  visible: boolean;
  onClose: () => void;
  onCreate?: (groupData: any) => void;
}

const CATEGORY_IDS: { [key: string]: number } = {
  movie: 1,
  game: 2,
  table_game: 3,
  other: 4,
};

const CreateGroupModal: React.FC<CreateGroupModalProps> = ({ visible, onClose, onCreate }) => {
  const colors = useThemedColors();
  const [groupName, setGroupName] = useState('');
  const [shortDescription, setShortDescription] = useState('');
  const [fullDescription, setFullDescription] = useState('');
  const [city, setCity] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [selectedContacts, setSelectedContacts] = useState<Contact[]>([]);
  const [contactsModalVisible, setContactsModalVisible] = useState(false);
  const [shortDescModalVisible, setShortDescModalVisible] = useState(false);
  const [fullDescModalVisible, setFullDescModalVisible] = useState(false);
  const [selectedImage, setSelectedImage] = useState<{ uri: string; name: string; type: string } | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const categories = [
    { id: 'movie', icon: require('@/assets/images/event_card/movie.png') },
    { id: 'game', icon: require('@/assets/images/event_card/game.png') },
    { id: 'table_game', icon: require('@/assets/images/event_card/table_game.png') },
    { id: 'other', icon: require('@/assets/images/event_card/other.png') },
  ];

  const toggleCategory = (categoryId: string) => {
    setSelectedCategories(prev => 
      prev.includes(categoryId) 
        ? prev.filter(id => id !== categoryId)
        : [...prev, categoryId]
    );
  };

  const showPermissionAlert = (onRetry: () => void) => {
    Alert.alert(
      'Разрешение на доступ к фото',
      'Для загрузки изображения группы необходимо разрешить доступ к галерее. Хотите предоставить разрешение?',
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

  const pickImage = async () => {
    try {
      console.log('[CreateGroupModal] 🖼️ Запрос на выбор изображения группы');

      const hasPermission = await permissionsService.checkMediaPermission();
      
      if (!hasPermission) {
        console.log('[CreateGroupModal] ⚠️ Нет разрешения на медиатеку');

        const granted = await permissionsService.requestMediaPermission();
        
        if (!granted) {
          console.log('[CreateGroupModal] ❌ Пользователь отклонил разрешение');
          showPermissionAlert(pickImage);
          return;
        }
      }

      console.log('[CreateGroupModal] ✅ Разрешение получено, открываем галерею');

      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: ['images'],
        allowsEditing: true,
        aspect: [1, 1],
        quality: 0.8,
      });

      console.log('[CreateGroupModal] 📦 Результат выбора:', result);

      if (!result.canceled && result.assets && result.assets[0]) {
        const asset = result.assets[0];
        const filename = asset.uri.split('/').pop() || 'group_image.jpg';
        const fileType = filename.split('.').pop()?.toLowerCase();
        
        console.log('[CreateGroupModal] ✅ Изображение выбрано:', asset.uri);
        
        setSelectedImage({
          uri: asset.uri,
          name: filename,
          type: `image/${fileType === 'jpg' ? 'jpeg' : fileType}`,
        });
      } else {
        console.log('[CreateGroupModal] ℹ️ Пользователь отменил выбор изображения');
      }
    } catch (error) {
      console.error('[CreateGroupModal] ❌ Ошибка выбора изображения:', error);
      Alert.alert('Ошибка', 'Не удалось загрузить изображение');
    }
  };

  const isFormValid = (): boolean => {
    const nameValidation = validateGroupName(groupName);
    const shortDescValidation = validateShortDescription(shortDescription);
    const fullDescValidation = validateFullDescription(fullDescription);
    
    return !!(
      nameValidation.isValid &&
      shortDescValidation.isValid &&
      fullDescValidation.isValid &&
      selectedCategories.length > 0 &&
      selectedImage
    );
  };

  const handleCreate = async () => {
    if (!groupName.trim()) {
      Alert.alert('Ошибка', 'Введите название группы');
      return;
    }
    if (!shortDescription.trim()) {
      Alert.alert('Ошибка', 'Введите краткое описание');
      return;
    }
    if (!fullDescription.trim()) {
      Alert.alert('Ошибка', 'Введите полное описание');
      return;
    }
    if (selectedCategories.length === 0) {
      Alert.alert('Ошибка', 'Выберите хотя бы одну категорию');
      return;
    }
    if (!selectedImage) {
      Alert.alert('Ошибка', 'Загрузите изображение группы');
      return;
    }

    setIsLoading(true);

    try {
      const categoryIds = selectedCategories.map(catId => CATEGORY_IDS[catId]);

      const groupData = {
        name: profanityFilter.clean(groupName.trim()),
        description: profanityFilter.clean(fullDescription.trim()),
        smallDescription: profanityFilter.clean(shortDescription.trim()),
        city: city.trim() || undefined,
        categories: categoryIds,
        isPrivate,
        image: selectedImage,
        contacts: selectedContacts.map(contact => ({
          name: contact.description || contact.id,
          link: contact.link
        })).filter(c => c.link.trim() !== ''),
      };

      const result = await groupService.createGroup(groupData);
      
      Alert.alert('Успех', 'Группа успешно создана!');
      onCreate?.(result);
      resetForm();
      onClose();
    } catch (error: any) {
      Alert.alert('Ошибка', error.message || 'Не удалось создать группу');
    } finally {
      setIsLoading(false);
    }
  };

  const resetForm = useCallback(() => {
    setGroupName('');
    setShortDescription('');
    setFullDescription('');
    setCity('');
    setIsPrivate(false);
    setSelectedCategories([]);
    setSelectedContacts([]);
    setSelectedImage(null);
  }, []);

  const handleClose = useCallback(() => {
    if (!isLoading) {
      resetForm();
      onClose();
    }
  }, [isLoading, resetForm, onClose]);

  useEffect(() => {
    if (!visible) return;

    const backHandler = BackHandler.addEventListener('hardwareBackPress', () => {
      if (isLoading) return true;
      handleClose();
      return true;
    });

    return () => backHandler.remove();
  }, [visible, isLoading, handleClose]);

  const handleContactsSave = (contacts: Contact[]) => {
    setSelectedContacts(contacts);
    setContactsModalVisible(false);
  };

  return (
    <>
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
                <Text style={[styles.title, { color: colors.black }]}>Основная информация</Text>
                <TouchableOpacity 
                  style={styles.closeButton} 
                  onPress={handleClose}
                  disabled={isLoading}
                >
                  <Image
                    tintColor={colors.black}
                    style={{ width: 35, height: 35, resizeMode: 'cover' }}
                    source={require('@/assets/images/event_card/back.png')}
                  />
                </TouchableOpacity>
              </View>

              <View style={[styles.content, { backgroundColor: colors.white }]}>
                <View style={styles.fieldContainer}>
                  <TextInput
                    style={[
                      styles.input,
                      { 
                        borderBottomColor: colors.grey,
                        color: colors.black
                      }
                    ]}
                    placeholder="Название *"
                    placeholderTextColor={colors.grey}
                    value={groupName}
                    onChangeText={setGroupName}
                    maxLength={40}
                    editable={!isLoading}
                  />
                  {groupName.length > 0 && (
                    <Text style={[styles.charCounter, { color: colors.grey }]}>
                      {groupName.length} / 40 {groupName.length < 5 && '(мин. 5)'}
                    </Text>
                  )}
                </View>

                <View style={styles.fieldContainer}>
                  <View style={styles.inputWithButtonContainer}>
                    <TextInput
                      style={[
                        styles.input,
                        { 
                          borderBottomColor: colors.grey,
                          color: colors.black
                        }
                      ]}
                      placeholder="Краткое описание *"
                      placeholderTextColor={colors.grey}
                      value={shortDescription}
                      onChangeText={setShortDescription}
                      maxLength={50}
                      editable={!isLoading}
                    />
                    <TouchableOpacity
                      style={styles.expandButton}
                      onPress={() => setShortDescModalVisible(true)}
                      disabled={isLoading}
                    >
                      <Image
                        source={require('@/assets/images/event_card/back.png')}
                        style={[styles.expandIcon, { tintColor: colors.grey }]}
                      />
                    </TouchableOpacity>
                  </View>
                  {shortDescription.length > 0 && (
                    <Text style={[styles.charCounter, { color: colors.grey }]}>
                      {shortDescription.length} / 50 {shortDescription.length < 5 && '(мин. 5)'}
                    </Text>
                  )}
                </View>

                <View style={styles.fieldContainer}>
                  <View style={styles.textAreaContainer}>
                    <TextInput
                      style={[
                        styles.input, 
                        styles.textArea,
                        {
                          borderColor: colors.grey,
                          color: colors.black
                        }
                      ]}
                      placeholder="Полное описание *"
                      placeholderTextColor={colors.grey}
                      value={fullDescription}
                      onChangeText={setFullDescription}
                      multiline
                      numberOfLines={6}
                      textAlignVertical="top"
                      maxLength={300}
                      editable={!isLoading}
                    />
                    <TouchableOpacity
                      style={styles.expandButtonTextArea}
                      onPress={() => setFullDescModalVisible(true)}
                      disabled={isLoading}
                    >
                      <Image
                        source={require('@/assets/images/event_card/back.png')}
                        style={[styles.expandIcon, { tintColor: colors.grey }]}
                      />
                    </TouchableOpacity>
                  </View>
                  {fullDescription.length > 0 && (
                    <Text style={[styles.charCounter, { color: colors.grey }]}>
                      {fullDescription.length} / 300 {fullDescription.length < 5 && '(мин. 5)'}
                    </Text>
                  )}
                </View>

                <TextInput
                  style={[
                    styles.input,
                    { 
                      borderBottomColor: colors.grey,
                      color: colors.black
                    }
                  ]}
                  placeholder="Город"
                  placeholderTextColor={colors.grey}
                  value={city}
                  onChangeText={setCity}
                  maxLength={50}
                  editable={!isLoading}
                />

                <View style={styles.checkboxContainer}>
                  <TouchableOpacity 
                    style={styles.checkbox}
                    onPress={() => setIsPrivate(!isPrivate)}
                    disabled={isLoading}
                  >
                    <View style={{flexDirection: 'row', alignItems: 'center'}}>
                      <View style={[
                        styles.checkboxCircle,
                        { borderColor: colors.grey },
                        isPrivate && { backgroundColor: colors.lightBlue }
                      ]} />
                      <Text style={[styles.checkboxLabel, { color: colors.black }]}>
                        Приватная группа
                      </Text>
                    </View>
                  </TouchableOpacity>
                </View>

                <View style={styles.categoriesAndImageSection}>
                  <View style={styles.leftSection}>
                    <Text style={[styles.sectionLabel, { color: colors.black }]}>Категории *</Text>
                    <View style={styles.categoriesContainer}>
                      {categories.map((category) => (
                        <TouchableOpacity
                          key={category.id}
                          style={[
                            styles.categoryButton,
                            { backgroundColor: colors.veryLightGrey },
                            selectedCategories.includes(category.id) && { backgroundColor: colors.lightBlue }
                          ]}
                          onPress={() => toggleCategory(category.id)}
                          disabled={isLoading}
                        >
                          <Image source={category.icon} style={styles.categoryIcon} />
                        </TouchableOpacity>
                      ))}
                    </View>

                    <Text style={[styles.sectionLabel, { color: colors.black }]}>Контакты:</Text>
                    <View style={styles.contactsContainer}>
                      <TouchableOpacity
                        style={styles.contactButton}
                        onPress={() => setContactsModalVisible(true)}
                        disabled={isLoading}
                      >
                        <Image 
                          source={require('@/assets/images/groups/contacts/add_contact.png')} 
                          style={styles.contactIcon} 
                        />
                      </TouchableOpacity>
                      
                      {selectedContacts
                        .filter(c => c.link && c.link.trim() !== '')
                        .map((contact, index) => (
                          <View key={`${contact.id}-${index}`} style={styles.selectedContactItem}>
                            <Image 
                              source={contact.icon} 
                              style={styles.contactIcon} 
                            />
                          </View>
                        ))}
                    </View>
                  </View>
                              
                  <TouchableOpacity 
                    style={styles.imageUpload}
                    onPress={pickImage}
                    disabled={isLoading}
                  >
                    {selectedImage ? (
                      <Image 
                        source={{ uri: selectedImage.uri }} 
                        style={styles.selectedImage}
                      />
                    ) : (
                      <View style={[
                        styles.uploadPlaceholder,
                        { 
                          backgroundColor: colors.veryLightGrey,
                          borderColor: colors.lightGrey
                        }
                      ]}>
                        <Image 
                          source={require('@/assets/images/groups/upload_image.png')} 
                          style={[styles.uploadIcon, { tintColor: colors.grey }]}
                        />
                        <Text style={[styles.uploadText, { color: colors.grey }]}>
                          Загрузите{'\n'}изображение *
                        </Text>
                      </View>
                    )}
                  </TouchableOpacity>
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
                      styles.createButton,
                      { backgroundColor: colors.white },
                      (!isFormValid() || isLoading) && styles.createButtonDisabled
                    ]}
                    onPress={handleCreate}
                    disabled={!isFormValid() || isLoading}
                  >
                    {isLoading ? (
                      <ActivityIndicator color={Colors.blue3} />
                    ) : (
                      <Text style={styles.createButtonText}>Создать группу</Text>
                    )}
                  </TouchableOpacity>
                </View>
              </ImageBackground>
            </ScrollView>
          </View>
        </View>
      </Modal>

      <ContactsModal
        visible={contactsModalVisible}
        onClose={() => setContactsModalVisible(false)}
        onSave={handleContactsSave}
      />

      <DescriptionModal
        visible={shortDescModalVisible}
        onClose={() => setShortDescModalVisible(false)}
        description={shortDescription}
        onChangeDescription={setShortDescription}
        maxLength={50}
        title="Краткое описание"
        placeholder="Введите краткое описание группы..."
      />

      <DescriptionModal
        visible={fullDescModalVisible}
        onClose={() => setFullDescModalVisible(false)}
        description={fullDescription}
        onChangeDescription={setFullDescription}
        maxLength={300}
        title="Полное описание"
        placeholder="Введите полное описание группы..."
      />
    </>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
  },
  modal: {
    marginHorizontal: 6,
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
    fontSize: 20,
    textAlign: 'center',
  },
  closeButton: {
    position: 'absolute',
    right: 10,
    zIndex: 1,
  },
  content: {
    paddingHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 16,
  },
  fieldContainer: {
    marginBottom: 16,
  },
  inputWithButtonContainer: {
    position: 'relative',
  },
  input: {
    borderBottomWidth: 1,
    paddingVertical: 4,
    paddingHorizontal: 0,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  textAreaContainer: {
    position: 'relative',
  },
  textArea: {
    borderWidth: 1,
    borderRadius: 8,
    padding: 16,
    paddingRight: 48,
    minHeight: 120,
    borderBottomWidth: 1,
  },
  charCounter: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    marginTop: 4,
  },
  expandButton: {
    position: 'absolute',
    right: 0,
    bottom: 4,
    width: 32,
    height: 32,
    justifyContent: 'center',
    alignItems: 'center',
  },
  expandButtonTextArea: {
    position: 'absolute',
    right: 8,
    bottom: 8,
    width: 32,
    height: 32,
    justifyContent: 'center',
    alignItems: 'center',
  },
  expandIcon: {
    width: 18,
    height: 18,
    transform: [{ rotate: '90deg' }],
  },
  checkboxContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 10,
  },
  checkbox: {
    marginRight: 8,
  },
  checkboxCircle: {
    width: 20,
    height: 20,
    borderRadius: 10,
    borderWidth: 2,
  },
  checkboxLabel: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    marginLeft: 12,
  },
  categoriesAndImageSection: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
  },
  leftSection: {
    flex: 1,
    paddingRight: 16,
  },
  sectionLabel: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    marginBottom: 8,
  },
  categoriesContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 6,
    marginBottom: 16,
  },
  categoryButton: {
    width: 35,
    height: 35,
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
  },
  categoryIcon: {
    width: 25,
    height: 25,
    resizeMode: 'contain',
  },
  contactsContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    alignItems: 'center',
  },
  contactButton: {
    width: 35,
    height: 35,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
  },
  contactIcon: {
    width: 35,
    height: 35,
    resizeMode: 'contain',
  },
  selectedContactItem: {
    width: 35,
    height: 35,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
  },
  imageUpload: {
    alignItems: 'center',
  },
  uploadPlaceholder: {
    width: 150,
    height: 150,
    borderRadius: 75,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderStyle: 'dashed',
  },
  uploadIcon: {
    width: 40,
    height: 40,
    resizeMode: 'contain',
  },
  uploadText: {
    fontFamily: Montserrat.bold,
    fontSize: 12,
    marginTop: 8,
    textAlign: 'center',
  },
  selectedImage: {
    width: 150,
    height: 150,
    borderRadius: 75,
    resizeMode: 'cover',
  },
  bottomBackground: {
    width: "100%",
    marginTop: -2
  },
  bottomContent: {
    padding: 16,
  },
  createButton: {
    marginHorizontal: 60,
    paddingVertical: 6,
    borderRadius: 20,
    alignItems: 'center',
  },
  createButtonDisabled: {
    opacity: 0.6,
  },
  createButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.blue3,
  },
});

export default CreateGroupModal;