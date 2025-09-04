import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React, { useState } from 'react';
import {
    Dimensions,
    Image,
    ImageBackground,
    Modal,
    ScrollView,
    StyleSheet,
    Text,
    TextInput,
    TouchableOpacity,
    TouchableWithoutFeedback,
    View
} from 'react-native';
import ContactsModal from './ContactsModal';

const screenHeight = Dimensions.get("window").height;

interface CreateGroupModalProps {
  visible: boolean;
  onClose: () => void;
  onCreate?: (groupData: any) => void;
}

const CreateGroupModal: React.FC<CreateGroupModalProps> = ({ visible, onClose, onCreate }) => {
  const [groupName, setGroupName] = useState('');
  const [shortDescription, setShortDescription] = useState('');
  const [fullDescription, setFullDescription] = useState('');
  const [city, setCity] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [selectedContacts, setSelectedContacts] = useState<string[]>([]);
  const [contactsModalVisible, setContactsModalVisible] = useState(false);

  const categories = [
    { id: 'movie', icon: require('../assets/images/event_card/movie.png') },
    { id: 'game', icon: require('../assets/images/event_card/game.png') },
    { id: 'table_game', icon: require('../assets/images/event_card/table_game.png') },
    { id: 'other', icon: require('../assets/images/event_card/other.png') },
  ];

  const contacts = [
    { id: 'add_contact', icon: require('../assets/images/groups/contacts/add_contact.png') },
  ];

  const toggleCategory = (categoryId: string) => {
    setSelectedCategories(prev => 
      prev.includes(categoryId) 
        ? prev.filter(id => id !== categoryId)
        : [...prev, categoryId]
    );
  };

    const toggleContact = (contactId: string) => {
    if (contactId === 'add_contact') {
        setContactsModalVisible(true);
    }
    };

  const handleCreate = () => {
    const groupData = {
      name: groupName,
      shortDescription,
      fullDescription,
      city,
      isPrivate,
      categories: selectedCategories,
      contacts: selectedContacts,
    };
    
    onCreate?.(groupData);
    onClose();
  };

  const resetForm = () => {
    setGroupName('');
    setShortDescription('');
    setFullDescription('');
    setCity('');
    setIsPrivate(false);
    setSelectedCategories([]);
    setSelectedContacts([]);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleContactsSave = (contacts: any[]) => {
  setSelectedContacts(contacts);
  setContactsModalVisible(false);
};

  return (
    <>
    <Modal visible={visible} animationType="fade" transparent>
      <View style={styles.overlay}>
        <TouchableWithoutFeedback onPress={handleClose}>
          <View style={StyleSheet.absoluteFill} />
        </TouchableWithoutFeedback>

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
                <TouchableOpacity style={styles.closeButton} onPress={handleClose}>
                    <Image
                    tintColor={Colors.black}
                    style={{ width: 35, height: 35, resizeMode: 'cover' }}
                    source={require('../assets/images/event_card/back.png')}
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
              />

              <TextInput
                style={styles.input}
                placeholder="Краткое описание"
                placeholderTextColor={Colors.grey}
                value={shortDescription}
                onChangeText={setShortDescription}
                maxLength={100}
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
              />

              <TextInput
                style={styles.input}
                placeholder="Город"
                placeholderTextColor={Colors.grey}
                value={city}
                onChangeText={setCity}
                maxLength={50}
              />

              <View style={styles.checkboxContainer}>
                <TouchableOpacity 
                  style={styles.checkbox}
                  onPress={() => setIsPrivate(!isPrivate)}
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
                      >
                        <Image source={category.icon} style={styles.categoryIcon} />
                      </TouchableOpacity>
                    ))}
                  </View>

                  <Text style={styles.sectionLabel}>Контакты:</Text>
                  <View style={styles.contactsContainer}>
                    {contacts.map((contact) => (
                      <TouchableOpacity
                        key={contact.id}
                        style={styles.contactButton}
                        onPress={() => toggleContact(contact.id)}
                      >
                        <Image source={contact.icon} style={styles.contactIcon} />
                      </TouchableOpacity>
                    ))}
                  </View>
                </View>
                
                <View style={styles.imageUpload}>
                  <View style={styles.uploadPlaceholder}>
                    <Image 
                      source={require('../assets/images/groups/upload_image.png')} 
                      style={styles.uploadIcon}
                    />
                  </View>
                </View>
              </View>
            </View>

            <ImageBackground
              source={require('../assets/images/event_card/bottom_rectangle.png')}
              style={styles.bottomBackground}
              resizeMode="stretch"
            >
              <View style={styles.bottomContent}>
                <TouchableOpacity
                  style={styles.createButton}
                  onPress={handleCreate}
                >
                  <Text style={styles.createButtonText}>Создать группу</Text>
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
    paddingVertical: 15,
    paddingHorizontal: 16,
    backgroundColor: Colors.white,
    position: 'relative',
  },
  title: {
    fontFamily: inter.black,
    fontSize: 20,
    color: Colors.black,
    textAlign: 'center',
  },
  closeButton: {
    position: 'absolute',
    right: 16,
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
    fontFamily: inter.regular,
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
    backgroundColor: Colors.lightBlue,
  },
  checkboxLabel: {
    fontFamily: inter.regular,
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
    fontFamily: inter.bold,
    fontSize: 16,
    color: Colors.black,
    marginBottom: 8,
  },
  categoriesContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
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
    backgroundColor: Colors.lightBlue,
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
  createButtonText: {
    fontFamily: inter.bold,
    fontSize: 16,
    color: Colors.blue,
  },
});

export default CreateGroupModal;