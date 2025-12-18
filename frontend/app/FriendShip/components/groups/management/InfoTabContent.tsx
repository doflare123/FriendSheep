import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import { getContactIconByLink } from '@/utils/contactHelpers';
import * as ImagePicker from 'expo-image-picker';
import React from 'react';
import {
  ActivityIndicator,
  Image,
  ImageBackground,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';

interface InfoTabContentProps {
  groupName: string;
  setGroupName: (value: string) => void;
  shortDescription: string;
  setShortDescription: (value: string) => void;
  fullDescription: string;
  setFullDescription: (value: string) => void;
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
  isSaving?: boolean;
  selectedContacts?: any[];
}

const InfoTabContent: React.FC<InfoTabContentProps> = ({
  groupName,
  setGroupName,
  shortDescription,
  setShortDescription,
  fullDescription,
  setFullDescription,
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
  isSaving = false,
  selectedContacts = [],
}) => {
  const colors = useThemedColors();

  const categories = [
    { id: 'movie', icon: require('@/assets/images/event_card/movie.png') },
    { id: 'game', icon: require('@/assets/images/event_card/game.png') },
    { id: 'table_game', icon: require('@/assets/images/event_card/table_game.png') },
    { id: 'other', icon: require('@/assets/images/event_card/other.png') },
  ];

  const handleImagePicker = async () => {
    const { status } = await ImagePicker.requestMediaLibraryPermissionsAsync();
    
    if (status !== 'granted') {
      alert('Необходимо разрешение на доступ к галерее');
      return;
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsEditing: true,
      aspect: [1, 1],
      quality: 0.8,
    });

    if (!result.canceled && result.assets[0]) {
      setGroupImage(result.assets[0].uri);
    }
  };

  return (
    <ScrollView
      style={styles.tabContent}
      showsVerticalScrollIndicator={false}
      keyboardShouldPersistTaps="handled"
    >
      <View style={[styles.content, { backgroundColor: colors.white }]}>
        <TextInput
          style={[
            styles.input,
            {
              borderBottomColor: colors.grey,
              color: colors.black
            }
          ]}
          placeholder="Название"
          placeholderTextColor={colors.grey}
          value={groupName}
          onChangeText={setGroupName}
          maxLength={50}
          editable={!isSaving}
        />

        <TextInput
          style={[
            styles.input,
            {
              borderBottomColor: colors.grey,
              color: colors.black
            }
          ]}
          placeholder="Краткое описание"
          placeholderTextColor={colors.grey}
          value={shortDescription}
          onChangeText={setShortDescription}
          maxLength={100}
          editable={!isSaving}
        />

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
          value={fullDescription}
          onChangeText={setFullDescription}
          multiline
          numberOfLines={6}
          textAlignVertical="top"
          maxLength={500}
          editable={!isSaving}
        />

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
          editable={!isSaving}
        />

        <View style={styles.checkboxContainer}>
          <TouchableOpacity 
            style={styles.checkbox}
            onPress={() => setIsPrivate(!isPrivate)}
            disabled={isSaving}
          >
            <View style={styles.checkboxRow}>
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
            <Text style={[styles.sectionLabel, { color: colors.black }]}>
              Категории:
            </Text>
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
                  disabled={isSaving}
                >
                  <Image source={category.icon} style={[styles.categoryIcon, {tintColor: colors.black}]} />
                </TouchableOpacity>
              ))}
            </View>

            <Text style={[styles.sectionLabel, { color: colors.black }]}>
              Контакты:
            </Text>
              <View style={styles.contactsContainer}>
                <TouchableOpacity
                  style={styles.contactButton}
                  onPress={onContactsPress}
                  disabled={isSaving}
                >
                  <Image 
                    source={require('@/assets/images/groups/contacts/add_contact.png')} 
                    style={styles.contactIcon} 
                  />
                </TouchableOpacity>
                
                {selectedContacts && selectedContacts.length > 0 && 
                  selectedContacts.map((contact, index) => {
                    const icon = getContactIconByLink(contact.link);
                    return (
                      <View key={`contact-${index}`} style={styles.contactButton}>
                        <Image 
                          source={icon} 
                          style={styles.contactIcon} 
                        />
                      </View>
                    );
                  })
                }
              </View>
          </View>
          
          <View style={styles.imageUpload}>
            <TouchableOpacity onPress={handleImagePicker} disabled={isSaving}>
              <View style={[
                styles.uploadPlaceholder,
                {
                  backgroundColor: colors.veryLightGrey,
                  borderColor: colors.lightGrey
                }
              ]}>
                {groupImage ? (
                  <Image source={{ uri: groupImage }} style={styles.groupImage} />
                ) : (
                  <Image 
                    source={require('@/assets/images/groups/upload_image.png')} 
                    style={[styles.uploadIcon, { tintColor: colors.grey }]}
                  />
                )}
              </View>
            </TouchableOpacity>
          </View>
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
              isSaving && styles.saveButtonDisabled
            ]}
            onPress={onSaveChanges}
            disabled={isSaving}
          >
            {isSaving ? (
              <ActivityIndicator color={Colors.blue} />
            ) : (
              <Text style={styles.saveButtonText}>Сохранить изменения</Text>
            )}
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
    paddingHorizontal: 16,
    paddingTop: 20,
    paddingBottom: 16,
  },
  input: {
    borderBottomWidth: 1,
    paddingVertical: 4,
    paddingHorizontal: 0,
    marginBottom: 16,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  textArea: {
    borderWidth: 1,
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
  checkboxRow: {
    flexDirection: 'row',
    alignItems: 'center',
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
    width: 140,
    height: 140,
    borderRadius: 100,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
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
    marginHorizontal: 60,
    paddingVertical: 6,
    borderRadius: 20,
    alignItems: 'center',
  },
  saveButtonDisabled: {
    opacity: 0.6,
  },
  saveButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.blue,
  },
});

export default InfoTabContent;