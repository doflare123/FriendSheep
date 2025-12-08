import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React, { useEffect, useMemo, useState } from 'react';
import {
  Dimensions,
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
  max: require('@/assets/images/groups/contacts/max.png'),
  default: require('@/assets/images/groups/contacts/default.png'),
};

const getContactType = (link: string) => {
  const lowerStr = link.toLowerCase();

  if (lowerStr.includes('discord.gg') || lowerStr.includes('discord.com')) return 'discord';
  if (lowerStr.includes('vk.com') || lowerStr.includes('vk.ru')) return 'vk';
  if (lowerStr.includes('t.me') || lowerStr.includes('telegram')) return 'telegram';
  if (lowerStr.includes('twitch.tv')) return 'twitch';
  if (lowerStr.includes('youtube.com') || lowerStr.includes('youtu.be')) return 'youtube';
  if (lowerStr.includes('max.ru')) return 'max';
  
  return 'default';
};

const getContactName = (type: string): string => {
  const names: Record<string, string> = {
    discord: 'Discord',
    vk: 'VK',
    telegram: 'Telegram',
    twitch: 'Twitch',
    youtube: 'YouTube',
    max: 'Max',
    default: 'Ссылка',
  };
  return names[type] || 'Ссылка';
};

const ContactsModal: React.FC<ContactsModalProps> = ({ 
  visible, 
  onClose, 
  onSave,
  initialContacts = []
}) => {
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [isInitialized, setIsInitialized] = useState(false);
  
  const [deleteConfirmVisible, setDeleteConfirmVisible] = useState(false);
  const [contactToDelete, setContactToDelete] = useState<number | null>(null);

  const memoizedInitialContacts = useMemo(() => {
    return initialContacts;
  }, [JSON.stringify(initialContacts)]);

  useEffect(() => {
    if (visible && !isInitialized) {
      if (memoizedInitialContacts && memoizedInitialContacts.length > 0) {
        const existingContacts = memoizedInitialContacts.map(contact => {
          const contactType = getContactType(contact.link);
          
          return {
            id: contactType,
            name: getContactName(contactType),
            icon: contactIcons[contactType as keyof typeof contactIcons] || contactIcons.default,
            description: contact.name,
            link: contact.link,
          };
        });
        
        console.log('[ContactsModal] 📋 Загружены контакты:', existingContacts);
        setContacts(existingContacts);
    } else {
      setContacts([
        { id: '', name: '', icon: contactIcons.default, description: '', link: '' }
      ]);
    }
      setIsInitialized(true);
    } else if (!visible && isInitialized) {
      setIsInitialized(false);
      setContacts([]);
    }
  }, [visible, isInitialized, memoizedInitialContacts]);

  const updateContactLink = (index: number, value: string) => {
    setContacts(prev => prev.map((contact, i) => {
      if (i !== index) return contact;

      const detectedType = getContactType(value);
      
      return {
        ...contact,
        link: value,
        id: detectedType,
        name: getContactName(detectedType),
        icon: contactIcons[detectedType as keyof typeof contactIcons] || contactIcons.default,
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
      icon: contactIcons.default, 
      description: '', 
      link: '' 
    }]);
  };

  const handleRemoveContactPress = (index: number) => {
    setContactToDelete(index);
    setDeleteConfirmVisible(true);
  };

  const confirmRemoveContact = () => {
    if (contactToDelete !== null) {
      setContacts(prev => prev.filter((_, i) => i !== contactToDelete));
    }
    setDeleteConfirmVisible(false);
    setContactToDelete(null);
  };

  const cancelRemoveContact = () => {
    setDeleteConfirmVisible(false);
    setContactToDelete(null);
  };

  const handleSave = () => {
    const filledContacts = contacts.filter(contact => 
      contact.link.trim() !== ''
    );

    const processedContacts = filledContacts.map(contact => ({
      ...contact,
      description: contact.description.trim() || contact.name || 'Ссылка'
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
    <>
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
                      <Text style={styles.contactNumber}>Контакт {index + 1}</Text>
                      <TouchableOpacity 
                        style={styles.removeButton}
                        onPress={() => handleRemoveContactPress(index)}
                      >
                        <Text style={styles.removeButtonText}>✕</Text>
                      </TouchableOpacity>
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
                        placeholder="Ссылка"
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
                    <Text style={styles.addButtonText}>Добавить контакт</Text>
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

      <Modal
        visible={deleteConfirmVisible}
        transparent={true}
        animationType="fade"
      >
        <View style={styles.confirmModalOverlay}>
          <View style={styles.confirmModalContent}>
            <Text style={styles.confirmModalTitle}>Удалить контакт?</Text>
            <Text style={styles.confirmModalMessage}>
              Контакт будет удалён из списка
            </Text>
            
            <View style={styles.confirmModalButtons}>
              <TouchableOpacity 
                style={[styles.confirmModalButton, styles.confirmCancelButton]}
                onPress={cancelRemoveContact}
              >
                <Text style={styles.confirmCancelButtonText}>Отмена</Text>
              </TouchableOpacity>
              
              <TouchableOpacity 
                style={[styles.confirmModalButton, styles.confirmDeleteButton]}
                onPress={confirmRemoveContact}
              >
                <Text style={styles.confirmModalButtonText}>Удалить</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
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
    justifyContent: 'space-between',
    marginBottom: 4,
  },
  contactNumber: {
    fontFamily: Montserrat.semibold,
    fontSize: 14,
    color: Colors.black,
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
    width: 30,
    height: 30,
    borderRadius: 15,
    justifyContent: 'center',
    alignItems: 'center',
  },
  removeButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 24,
    color: Colors.red,
  },
  addButton: {
    backgroundColor: Colors.lightBlue,
    paddingVertical: 12,
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
    paddingVertical: 12,
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
    backgroundColor: Colors.lightBlue,
    paddingVertical: 12,
    borderRadius: 25,
    alignItems: 'center',
    marginLeft: 10,
  },
  saveButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
  confirmModalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  confirmModalContent: {
    backgroundColor: Colors.white,
    borderRadius: 20,
    padding: 24,
    marginHorizontal: 40,
    width: '80%',
  },
  confirmModalTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
    marginBottom: 12,
  },
  confirmModalMessage: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
    textAlign: 'center',
    marginBottom: 24,
  },
  confirmModalButtons: {
    flexDirection: 'row',
    gap: 12,
  },
  confirmModalButton: {
    flex: 1,
    paddingVertical: 10,
    borderRadius: 25,
    alignItems: 'center',
  },
  confirmDeleteButton: {
    backgroundColor: Colors.red,
  },
  confirmCancelButton: {
    backgroundColor: Colors.veryLightGrey,
  },
  confirmModalButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.white,
  },
  confirmCancelButtonText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
  },
});

export default ContactsModal;