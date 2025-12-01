import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import {
    Image,
    Modal,
    StyleSheet,
    Text,
    TextInput,
    TouchableOpacity,
    View
} from 'react-native';

interface DescriptionModalProps {
  visible: boolean;
  onClose: () => void;
  description: string;
  onChangeDescription: (text: string) => void;
  maxLength?: number;
  title?: string;
  placeholder?: string;
}

const DescriptionModal: React.FC<DescriptionModalProps> = ({
  visible,
  onClose,
  description,
  onChangeDescription,
  maxLength = 300,
  title = 'Описание события',
  placeholder = 'Введите описание события...',
}) => {
  return (
    <Modal
      visible={visible}
      animationType="fade"
      transparent
      onRequestClose={onClose}
    >
      <View style={styles.overlay}>
        <View style={styles.content}>
          <View style={styles.header}>
            <Text style={styles.title}>{title}</Text>
            <TouchableOpacity
              onPress={onClose}
              style={styles.closeButton}
            >
              <Image
                source={require('@/assets/images/event_card/back.png')}
                style={styles.closeIcon}
              />
            </TouchableOpacity>
          </View>
          
          <TextInput
            style={styles.textArea}
            placeholder={placeholder}
            placeholderTextColor={Colors.grey}
            value={description}
            onChangeText={onChangeDescription}
            multiline
            textAlignVertical="top"
            maxLength={maxLength}
            autoFocus
          />
          
          <Text style={styles.counter}>
            {description.length} / {maxLength}
          </Text>
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
    paddingHorizontal: 20,
    minHeight: 240,
  },
  content: {
    backgroundColor: Colors.white,
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 12,
    maxHeight: '80%',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.black,
  },
  closeButton: {
    padding: 4,
  },
  closeIcon: {
    width: 24,
    height: 24,
    tintColor: Colors.black,
    transform: [{ rotate: '180deg' }],
  },
  textArea: {
    borderWidth: 1,
    borderColor: Colors.grey,
    borderRadius: 12,
    padding: 16,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    minHeight: 230,
    textAlignVertical: 'top',
  },
  counter: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.grey,
    textAlign: 'right',
    marginTop: 8,
  },
});

export default DescriptionModal;