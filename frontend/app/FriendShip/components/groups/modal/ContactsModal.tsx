import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React, { useState } from 'react';
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

interface Contact {
  id: string;
  name: string;
  icon: any;
  description: string;
  link: string;
}

interface ContactsModalProps {
  visible: boolean;
  onClose: () => void;
  onSave?: (contacts: Contact[]) => void;
}

const ContactsModal: React.FC<ContactsModalProps> = ({ visible, onClose, onSave }) => {
  const [contacts, setContacts] = useState<Contact[]>([
    {
      id: 'discord',
      name: 'Discord',
      icon: require('@/assets/images/groups/contacts/discord.png'),
      description: '',
      link: ''
    },
    {
      id: 'vk',
      name: 'VK',
      icon: require('@/assets/images/groups/contacts/vk.png'),
      description: '',
      link: ''
    },
    {
      id: 'telegram',
      name: 'Telegram',
      icon: require('@/assets/images/groups/contacts/telegram.png'),
      description: '',
      link: ''
    },
    {
      id: 'twitch',
      name: 'Twitch',
      icon: require('@/assets/images/groups/contacts/twitch.png'),
      description: '',
      link: ''
    },
    {
      id: 'youtube',
      name: 'YouTube',
      icon: require('@/assets/images/groups/contacts/youtube.png'),
      description: '',
      link: ''
    },
    {
      id: 'whatsapp',
      name: 'WhatsApp',
      icon: require('@/assets/images/groups/contacts/whatsapp.png'),
      description: '',
      link: ''
    },
    {
      id: 'max',
      name: 'Max',
      icon: require('@/assets/images/groups/contacts/max.png'),
      description: '',
      link: ''
    }
  ]);

  const updateContact = (id: string, field: 'description' | 'link', value: string) => {
    setContacts(prev => prev.map(contact => 
      contact.id === id ? { ...contact, [field]: value } : contact
    ));
  };

  const handleSave = () => {
    const filledContacts = contacts.filter(contact => 
      contact.description.trim() !== '' || contact.link.trim() !== ''
    );
    onSave?.(filledContacts);
    onClose();
  };

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <View style={styles.overlay}>
        <TouchableWithoutFeedback onPress={onClose}>
          <View style={StyleSheet.absoluteFill} />
        </TouchableWithoutFeedback>

        <View style={styles.modal}>
          <View style={styles.header}>
            <Text style={styles.title}>Выберите и заполните контакты</Text>
          </View>

          <ScrollView
            showsVerticalScrollIndicator={false}
            keyboardShouldPersistTaps="handled"
            bounces={false}
            style={styles.scrollView}
          >
            <View style={styles.content}>
              {contacts.map((contact) => (
                <View key={contact.id} style={styles.contactItem}>
                  <View style={styles.iconContainer}>
                    <Image source={contact.icon} style={styles.contactIcon} />
                  </View>
                  <View style={styles.inputsContainer}>
                    <TextInput
                      style={styles.input}
                      placeholder="Описание"
                      placeholderTextColor={Colors.grey}
                      value={contact.description}
                      onChangeText={(text) => updateContact(contact.id, 'description', text)}
                      maxLength={50}
                    />
                    <TextInput
                      style={styles.input}
                      placeholder="Ссылка"
                      placeholderTextColor={Colors.grey}
                      value={contact.link}
                      onChangeText={(text) => updateContact(contact.id, 'link', text)}
                      maxLength={100}
                      keyboardType={contact.id === 'whatsapp' ? 'phone-pad' : 'url'}
                    />
                  </View>
                </View>
              ))}
            </View>
          </ScrollView>

          <View style={styles.footer}>
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
    maxHeight: screenHeight * 0.75,
    overflow: 'hidden',
  },
  header: {
    paddingVertical: 10,
    paddingHorizontal: 20,
    backgroundColor: Colors.white,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.black,
    textAlign: 'center',
  },
  scrollView: {
    flexGrow: 1,
  },
  content: {
    padding: 22,
  },
  contactItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 20,
    paddingBottom: 15,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightLightGrey,
  },
  iconContainer: {
    width: 50,
    height: 50,
    borderRadius: 25,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 15,
  },
  contactIcon: {
    width: 70,
    height: 70,
    resizeMode: 'contain',
  },
  inputsContainer: {
    flex: 1,
  },
  input: {
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
    paddingVertical: 8,
    paddingHorizontal: 0,
    marginBottom: 8,
    marginLeft: 10,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
  },
  footer: {
    padding: 20,
    backgroundColor: Colors.white,
    borderTopWidth: 1,
    borderTopColor: Colors.lightGrey,
  },
  saveButton: {
    backgroundColor: Colors.lightBlue,
    paddingVertical: 6,
    borderRadius: 25,
    alignItems: 'center',
  },
  saveButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
});

export default ContactsModal;