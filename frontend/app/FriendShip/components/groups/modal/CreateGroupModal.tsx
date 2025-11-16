import groupService from '@/api/services/groupService';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import * as ImagePicker from 'expo-image-picker';
import React, { useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Image,
  ImageBackground,
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
  const [groupName, setGroupName] = useState('');
  const [shortDescription, setShortDescription] = useState('');
  const [fullDescription, setFullDescription] = useState('');
  const [city, setCity] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [selectedContacts, setSelectedContacts] = useState<Contact[]>([]);
  const [contactsModalVisible, setContactsModalVisible] = useState(false);
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

  const pickImage = async () => {
    const { status } = await ImagePicker.requestMediaLibraryPermissionsAsync();
    
    if (status !== 'granted') {
      Alert.alert('Ошибка', 'Необходимо разрешение на доступ к галерее');
      return;
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsEditing: true,
      aspect: [1, 1],
      quality: 0.8,
    });

    if (!result.canceled && result.assets[0]) {
      const asset = result.assets[0];
      const filename = asset.uri.split('/').pop() || 'group_image.jpg';
      const fileType = filename.split('.').pop()?.toLowerCase();
      
      setSelectedImage({
        uri: asset.uri,
        name: filename,
        type: `image/${fileType === 'jpg' ? 'jpeg' : fileType}`,
      });
    }
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
        name: groupName.trim(),
        description: fullDescription.trim(),
        smallDescription: shortDescription.trim(),
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

  const resetForm = () => {
    setGroupName('');
    setShortDescription('');
    setFullDescription('');
    setCity('');
    setIsPrivate(false);
    setSelectedCategories([]);
    setSelectedContacts([]);
    setSelectedImage(null);
  };

  const handleClose = () => {
    if (!isLoading) {
      resetForm();
      onClose();
    }
  };

  const handleContactsSave = (contacts: Contact[]) => {
    setSelectedContacts(contacts);
    setContactsModalVisible(false);
  };

  return (
    <>
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
                <Text style={styles.title}>Основная информация</Text>
                <TouchableOpacity 
                  style={styles.closeButton} 
                  onPress={handleClose}
                  disabled={isLoading}
                >
                  <Image
                    tintColor={Colors.black}
                    style={{ width: 35, height: 35, resizeMode: 'cover' }}
                    source={require('@/assets/images/event_card/back.png')}
                  />
                </TouchableOpacity>
              </View>

              <View style={styles.content}>
                <TextInput
                  style={styles.input}
                  placeholder="Название"
                  placeholderTextColor={Colors.grey}
                  value={groupName}
                  onChangeText={setGroupName}
                  maxLength={50}
                  editable={!isLoading}
                />

                <TextInput
                  style={styles.input}
                  placeholder="Краткое описание"
                  placeholderTextColor={Colors.grey}
                  value={shortDescription}
                  onChangeText={setShortDescription}
                  maxLength={100}
                  editable={!isLoading}
                />

                <TextInput
                  style={[styles.input, styles.textArea]}
                  placeholder="Описание"
                  placeholderTextColor={Colors.grey}
                  value={fullDescription}
                  onChangeText={setFullDescription}
                  multiline
                  numberOfLines={6}
                  textAlignVertical="top"
                  maxLength={500}
                  editable={!isLoading}
                />

                <TextInput
                  style={styles.input}
                  placeholder="Город (необязательно)"
                  placeholderTextColor={Colors.grey}
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
                      <View style={[styles.checkboxCircle, isPrivate && styles.checkboxSelected]} />
                      <Text style={styles.checkboxLabel}>Эта группа приватная?</Text>
                    </View>
                  </TouchableOpacity>
                </View>

                <View style={styles.categoriesAndImageSection}>
                  <View style={styles.leftSection}>
                    <Text style={styles.sectionLabel}>Категории:</Text>
                    <View style={styles.categoriesContainer}>
                      {categories.map((category) => (
                        <TouchableOpacity
                          key={category.id}
                          style={[
                            styles.categoryButton,
                            selectedCategories.includes(category.id) && styles.categoryButtonSelected
                          ]}
                          onPress={() => toggleCategory(category.id)}
                          disabled={isLoading}
                        >
                          <Image source={category.icon} style={styles.categoryIcon} />
                        </TouchableOpacity>
                      ))}
                    </View>

                  <Text style={styles.sectionLabel}>Контакты:</Text>
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
                      <View style={styles.uploadPlaceholder}>
                        <Image 
                          source={require('@/assets/images/groups/upload_image.png')} 
                          style={styles.uploadIcon}
                        />
                      </View>
                    )}
                  </TouchableOpacity>
                </View>
              </View>

              <ImageBackground
                source={require('@/assets/images/event_card/bottom_rectangle.png')}
                style={styles.bottomBackground}
                resizeMode="stretch"
                tintColor={Colors.lightBlue3}
              >
                <View style={styles.bottomContent}>
                  <TouchableOpacity
                    style={[styles.createButton, isLoading && styles.createButtonDisabled]}
                    onPress={handleCreate}
                    disabled={isLoading}
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
    fontSize: 20,
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
    paddingTop: 8,
    paddingBottom: 16,
  },
  input: {
    borderBottomWidth: 1,
    borderBottomColor: Colors.grey,
    paddingVertical: 4,
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
    minHeight: 120,
  },
  checkboxContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 20,
  },
  checkbox: {
    marginRight: 8,
  },
  checkboxCircle: {
    width: 20,
    height: 20,
    borderRadius: 10,
    borderWidth: 2,
    borderColor: Colors.grey,
    backgroundColor: Colors.white,
  },
  checkboxSelected: {
    backgroundColor: Colors.lightBlue3,
  },
  checkboxLabel: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    marginLeft: 12,
    color: Colors.black,
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
    color: Colors.black,
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
    backgroundColor: Colors.lightGrey,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
  },
  categoryButtonSelected: {
    backgroundColor: Colors.lightBlue3,
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
    borderRadius: 120,
    backgroundColor: Colors.lightLightGrey,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: Colors.lightGrey,
    borderStyle: 'dashed',
  },
  uploadIcon: {
    width: 40,
    height: 40,
    resizeMode: 'contain',
    tintColor: Colors.grey,
  },
  selectedImage: {
    width: 150,
    height: 150,
    borderRadius: 75,
    resizeMode: 'cover',
  },
  bottomBackground: {
    width: "100%",
  },
  bottomContent: {
    padding: 16,
  },
  createButton: {
    backgroundColor: Colors.white,
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