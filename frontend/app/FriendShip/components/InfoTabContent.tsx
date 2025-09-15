import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React from 'react';
import {
    Image,
    ImageBackground,
    ScrollView,
    StyleSheet,
    Text,
    TextInput,
    TouchableOpacity,
    View,
} from 'react-native';
import { MediaType, launchImageLibrary } from 'react-native-image-picker';

interface InfoTabContentProps {
  groupName: string;
  setGroupName: (value: string) => void;
  shortDescription: string;
  setShortDescription: (value: string) => void;
  fullDescription: string;
  setFullDescription: (value: string) => void;
  country: string;
  setCountry: (value: string) => void;
  city: string;
  setCity: (value: string) => void;
  isPrivate: boolean;
  setIsPrivate: (value: boolean) => void;
  selectedCategories: string[];
  toggleCategory: (categoryId: string) => void;
  groupImage: string;
  setGroupImage: (value: string) => void;
  onContactsPress: () => void;
  onSaveChanges: () => void;
}

const InfoTabContent: React.FC<InfoTabContentProps> = ({
  groupName,
  setGroupName,
  shortDescription,
  setShortDescription,
  fullDescription,
  setFullDescription,
  country,
  setCountry,
  city,
  setCity,
  isPrivate,
  setIsPrivate,
  selectedCategories,
  toggleCategory,
  groupImage,
  setGroupImage,
  onContactsPress,
  onSaveChanges,
}) => {
  const categories = [
    { id: 'movie', icon: require('../assets/images/event_card/movie.png') },
    { id: 'game', icon: require('../assets/images/event_card/game.png') },
    { id: 'table_game', icon: require('../assets/images/event_card/table_game.png') },
    { id: 'other', icon: require('../assets/images/event_card/other.png') },
  ];

  const contacts = [
    { id: 'add_contact', icon: require('../assets/images/groups/contacts/add_contact.png') },
  ];

  const handleImagePicker = () => {
    const options = {
      mediaType: 'photo' as MediaType,
      includeBase64: false,
      maxHeight: 2000,
      maxWidth: 2000,
    };

    launchImageLibrary(options, (response) => {
      if (response.didCancel || response.errorMessage) {
        return;
      }

      if (response.assets && response.assets[0]) {
        setGroupImage(response.assets[0].uri || '');
      }
    });
  };

  return (
    <ScrollView
      style={styles.tabContent}
      showsVerticalScrollIndicator={false}
      keyboardShouldPersistTaps="handled"
    >
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

        <View style={styles.locationRow}>
          <TextInput
            style={[styles.input, styles.locationInput]}
            placeholder="Страна"
            placeholderTextColor={Colors.grey}
            value={country}
            onChangeText={setCountry}
            maxLength={50}
          />
          <TextInput
            style={[styles.input, styles.locationInput]}
            placeholder="Город"
            placeholderTextColor={Colors.grey}
            value={city}
            onChangeText={setCity}
            maxLength={50}
          />
        </View>

        <View style={styles.checkboxContainer}>
          <TouchableOpacity 
            style={styles.checkbox}
            onPress={() => setIsPrivate(!isPrivate)}
          >
            <View style={styles.checkboxRow}>
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
                  onPress={onContactsPress}
                >
                  <Image source={contact.icon} style={styles.contactIcon} />
                </TouchableOpacity>
              ))}
            </View>
          </View>
          
          <View style={styles.imageUpload}>
            <TouchableOpacity onPress={handleImagePicker}>
              <View style={styles.uploadPlaceholder}>
                {groupImage ? (
                  <Image source={{ uri: groupImage }} style={styles.groupImage} />
                ) : (
                  <Image 
                    source={require('../assets/images/groups/upload_image.png')} 
                    style={styles.uploadIcon}
                  />
                )}
              </View>
            </TouchableOpacity>
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
            style={styles.saveButton}
            onPress={onSaveChanges}
          >
            <Text style={styles.saveButtonText}>Сохранить изменения</Text>
          </TouchableOpacity>
        </View>
      </ImageBackground>
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  tabContent: {
    flex: 1,
  },
  content: {
    backgroundColor: Colors.white,
    paddingHorizontal: 16,
    paddingTop: 20,
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
  locationRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 16,
  },
  locationInput: {
    flex: 1,
  },
  checkboxContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 20,
  },
  checkbox: {
    marginRight: 8,
  },
  checkboxRow: {
    flexDirection: 'row',
    alignItems: 'center',
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
    width: 120,
    height: 120,
    borderRadius: 60,
    backgroundColor: Colors.lightLightGrey,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: Colors.lightGrey,
    borderStyle: 'dashed',
    overflow: 'hidden',
  },
  groupImage: {
    width: '100%',
    height: '100%',
    borderRadius: 60,
  },
  uploadIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
    tintColor: Colors.grey,
  },
  bottomBackground: {
    width: "100%",
    height: 80,
    marginBottom: -10
  },
  bottomContent: {
    padding: 20,
  },
  saveButton: {
    backgroundColor: Colors.white,
    marginHorizontal: 60,
    paddingVertical: 6,
    borderRadius: 20,
    alignItems: 'center',
  },
  saveButtonText: {
    fontFamily: inter.bold,
    fontSize: 16,
    color: Colors.blue,
  },
});

export default InfoTabContent;