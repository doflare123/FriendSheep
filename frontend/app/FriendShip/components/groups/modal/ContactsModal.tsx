import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React, { useEffect, useMemo, useState } from 'react';
import {
  Dimensions,
  Image,
  Modal,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View
} from 'react-native';

const screenHeight = Dimensions.get("window").height;

export interface Contact {
  id: string;
  name: string;
  icon: any;
  description: string;
  link: string;
}

interface GroupContact {
  name: string;
  link: string;
}

interface ContactsModalProps {
  visible: boolean;
  onClose: () => void;
  onSave?: (contacts: Contact[]) => void;
  initialContacts?: GroupContact[];
}

const contactIcons = {
  discord: require('@/assets/images/groups/contacts/discord.png'),
  vk: require('@/assets/images/groups/contacts/vk.png'),
  telegram: require('@/assets/images/groups/contacts/telegram.png'),
  twitch: require('@/assets/images/groups/contacts/twitch.png'),
  youtube: require('@/assets/images/groups/contacts/youtube.png'),
  whatsapp: require('@/assets/images/groups/contacts/whatsapp.png'),
  max: require('@/assets/images/groups/contacts/max.png'),
  default: require('@/assets/images/groups/contacts/default.png'),
};

const CONTACT_TYPES = [
  { id: 'discord', name: 'Discord', icon: contactIcons.discord, placeholder: 'https://discord.gg/...' },
  { id: 'vk', name: 'VK', icon: contactIcons.vk, placeholder: 'https://vk.com/...' },
  { id: 'telegram', name: 'Telegram', icon: contactIcons.telegram, placeholder: 'https://t.me/...' },
  { id: 'twitch', name: 'Twitch', icon: contactIcons.twitch, placeholder: 'https://twitch.tv/...' },
  { id: 'youtube', name: 'YouTube', icon: contactIcons.youtube, placeholder: 'https://youtube.com/...' },
  { id: 'whatsapp', name: 'WhatsApp', icon: contactIcons.whatsapp, placeholder: 'https://wa.me/...' },
  { id: 'max', name: 'Max', icon: contactIcons.max, placeholder: 'https://max.ru/...'},
  { id: 'default', name: 'Ваш выбор', icon: contactIcons.default, placeholder: 'https://...'},
];

const getContactType = (link: string) => {
  const lowerStr = link.toLowerCase();

  if (lowerStr.includes('discord.gg') || lowerStr.includes('discord.com')) return 'discord';
  if (lowerStr.includes('vk.com') || lowerStr.includes('vk.ru')) return 'vk';
  if (lowerStr.includes('t.me') || lowerStr.includes('telegram')) return 'telegram';
  if (lowerStr.includes('twitch.tv')) return 'twitch';
  if (lowerStr.includes('youtube.com') || lowerStr.includes('youtu.be')) return 'youtube';
  if (lowerStr.includes('wa.me') || lowerStr.includes('whatsapp')) return 'whatsapp';
  if (lowerStr.includes('max.ru')) return 'max';
  
  return 'default';
};

const ContactsModal: React.FC<ContactsModalProps> = ({ 
  visible, 
  onClose, 
  onSave,
  initialContacts = []
}) => {
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [isInitialized, setIsInitialized] = useState(false);

  const memoizedInitialContacts = useMemo(() => {
    return initialContacts;
  }, [JSON.stringify(initialContacts)]);

  useEffect(() => {
    if (visible && !isInitialized) {
      if (memoizedInitialContacts && memoizedInitialContacts.length > 0) {
        const existingContacts = memoizedInitialContacts.map(contact => {
          const contactType = getContactType(contact.link);
          const contactInfo = CONTACT_TYPES.find(ct => ct.id === contactType) || { id: 'default', name: 'Ваш выбор', icon: contactIcons.default, placeholder: 'https://...' };
          
          return {
            id: contactType,
            name: contactInfo.name,
            icon: contactInfo.icon,
            description: contact.name,
            link: contact.link,
          };
        });
        
        setContacts(existingContacts);
      } else {
        setContacts([
          { id: '', name: '', icon: contactIcons.max, description: '', link: '' }
        ]);
      }
      setIsInitialized(true);
    } else if (!visible && isInitialized) {
      setIsInitialized(false);
      setContacts([]);
    }
  }, [visible, isInitialized, memoizedInitialContacts]);

  const selectContactType = (index: number, typeId: string) => {
    const contactType = CONTACT_TYPES.find(ct => ct.id === typeId);
    if (!contactType) return;

    setContacts(prev => prev.map((contact, i) => 
      i === index 
        ? { 
            ...contact, 
            id: typeId,
            name: contactType.name,
            description: contact.description || contactType.name
          }
        : contact
    ));
  };

  const updateContactLink = (index: number, value: string) => {
    setContacts(prev => prev.map((contact, i) => {
      if (i !== index) return contact;

      const detectedType = getContactType(value);
      const contactType = CONTACT_TYPES.find(ct => ct.id === detectedType);
      
      return {
        ...contact,
        link: value,
        ...(contact.id === '' && contactType ? {
          id: detectedType,
          name: contactType.name,
          icon: contactType.icon,
          description: contact.description || contactType.name
        } : {})
      };
    }));
  };

  const updateContactDescription = (index: number, value: string) => {
    setContacts(prev => prev.map((contact, i) => 
      i === index ? { ...contact, description: value } : contact
    ));
  };

  const addContact = () => {
    setContacts(prev => [...prev, { 
      id: '', 
      name: '', 
      icon: contactIcons.max, 
      description: '', 
      link: '' 
    }]);
  };

  const removeContact = (index: number) => {
    if (contacts.length > 1) {
      setContacts(prev => prev.filter((_, i) => i !== index));
    }
  };

  const handleSave = () => {
    const filledContacts = contacts.filter(contact => 
      contact.link.trim() !== ''
    );
    
    const processedContacts = filledContacts.map(contact => ({
      ...contact,
      description: contact.description || contact.name || contact.id
    }));
    
    console.log('Сохранение контактов:', processedContacts);
    onSave?.(processedContacts);
    handleClose();
  };

  const handleClose = () => {
    setIsInitialized(false);
    setContacts([]);
    onClose();
  };

  const hasEmptyContact = contacts.some(c => !c.link || c.link.trim() === '');

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <View style={styles.overlay}>
        <TouchableWithoutFeedback onPress={handleClose}>
          <View style={StyleSheet.absoluteFill} />
        </TouchableWithoutFeedback>

        <View style={styles.modal}>
          <View style={styles.header}>
            <Text style={styles.title}>Контакты группы</Text>
            <Text style={styles.subtitle}>Добавьте ссылки на социальные сети</Text>
          </View>

          <ScrollView
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
            bounces={false}
            style={styles.scrollView}
          >
            <View style={styles.content}>
              {contacts.map((contact, index) => (
                <View key={`contact-item-${index}`} style={styles.contactItem}>
                  <View style={styles.contactHeader}>
                    <ScrollView 
                      horizontal 
                      showsHorizontalScrollIndicator={false}
                      style={styles.typesScroll}
                    >
                      {CONTACT_TYPES.map(type => (
                        <TouchableOpacity
                          key={type.id}
                          style={[
                            styles.typeButton,
                            contact.id === type.id && styles.typeButtonSelected
                          ]}
                          onPress={() => selectContactType(index, type.id)}
                        >
                          <Image source={type.icon} style={styles.typeIcon} />
                          <Text style={[
                            styles.typeText,
                            contact.id === type.id && styles.typeTextSelected
                          ]}>
                            {type.name}
                          </Text>
                        </TouchableOpacity>
                      ))}
                    </ScrollView>

                    {contacts.length > 1 && (
                      <TouchableOpacity 
                        style={styles.removeButton}
                        onPress={() => removeContact(index)}
                      >
                        <Text style={styles.removeButtonText}>✕</Text>
                      </TouchableOpacity>
                    )}
                  </View>

                  <View style={styles.inputsContainer}>
                    <TextInput
                      style={styles.input}
                      placeholder="Название (необязательно)"
                      placeholderTextColor={Colors.grey}
                      value={contact.description}
                      onChangeText={(text) => updateContactDescription(index, text)}
                      maxLength={50}
                    />
                    <TextInput
                      style={styles.input}
                      placeholder={
                        contact.id 
                          ? CONTACT_TYPES.find(t => t.id === contact.id)?.placeholder 
                          : "Ссылка"
                      }
                      placeholderTextColor={Colors.grey}
                      value={contact.link}
                      onChangeText={(text) => updateContactLink(index, text)}
                      maxLength={200}
                      keyboardType="url"
                      autoCapitalize="none"
                    />
                  </View>
                </View>
              ))}

              {!hasEmptyContact && (
                <TouchableOpacity style={styles.addButton} onPress={addContact}>
                  <Text style={styles.addButtonText}>+ Добавить контакт</Text>
                </TouchableOpacity>
              )}
            </View>
          </ScrollView>

          <View style={styles.footer}>
            <TouchableOpacity style={styles.cancelButton} onPress={handleClose}>
              <Text style={styles.cancelButtonText}>Отмена</Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.saveButton} onPress={handleSave}>
              <Text style={styles.saveButtonText}>Сохранить</Text>
            </TouchableOpacity>
          </View>
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
    marginHorizontal: 20,
    borderRadius: 20,
    backgroundColor: Colors.white,
    maxHeight: screenHeight * 0.8,
    overflow: 'hidden',
  },
  header: {
    paddingVertical: 16,
    paddingHorizontal: 20,
    backgroundColor: Colors.white,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
    marginBottom: 4,
  },
  subtitle: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.grey,
    textAlign: 'center',
  },
  scrollView: {
    flexGrow: 1,
  },
  content: {
    padding: 20,
  },
  contactItem: {
    marginBottom: 24,
    paddingBottom: 20,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightLightGrey,
  },
  contactHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  typesScroll: {
    flex: 1,
  },
  typeButton: {
    flexDirection: 'column',
    alignItems: 'center',
    width: 80,
    marginRight: 8,
    paddingVertical: 8,
    borderRadius: 16,
    backgroundColor: Colors.veryLightGrey,
  },
  typeButtonSelected: {
    backgroundColor: Colors.lightBlue3,
  },
  typeIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
    marginBottom: 4,
  },
  typeText: {
    fontFamily: Montserrat.regular,
    fontSize: 10,
    color: Colors.black,
  },
  typeTextSelected: {
    fontFamily: Montserrat.bold,
    color: Colors.white,
  },
  inputsContainer: {
    flex: 1,
  },
  input: {
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
    paddingVertical: 8,
    paddingHorizontal: 0,
    marginBottom: 12,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
  },
  removeButton: {
    position: 'absolute',
    right: 0,
    top: -12,
    width: 25,
    height: 25,
    borderRadius: 15,
    backgroundColor: Colors.red,
    justifyContent: 'center',
    alignItems: 'center',
  },
  removeButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
  addButton: {
    backgroundColor: Colors.lightBlue3,
    paddingVertical: 8,
    borderRadius: 20,
    alignItems: 'center',
    marginTop: 10,
  },
  addButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.white,
  },
  footer: {
    flexDirection: 'row',
    padding: 20,
    backgroundColor: Colors.white,
    borderTopWidth: 1,
    borderTopColor: Colors.lightGrey,
  },
  cancelButton: {
    flex: 1,
    backgroundColor: Colors.veryLightGrey,
    paddingVertical: 10,
    borderRadius: 25,
    alignItems: 'center',
    marginRight: 10,
  },
  cancelButtonText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
  },
  saveButton: {
    flex: 1,
    backgroundColor: Colors.lightBlue3,
    paddingVertical: 10,
    borderRadius: 25,
    alignItems: 'center',
    marginLeft: 10,
  },
  saveButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
});

export default ContactsModal;