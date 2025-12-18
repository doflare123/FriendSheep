import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
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
  const colors = useThemedColors();

  return (
    <Modal
      visible={visible}
      animationType="fade"
      transparent
      onRequestClose={onClose}
    >
      <View style={styles.overlay}>
        <View style={[styles.content, { backgroundColor: colors.white }]}>
          <View style={styles.header}>
            <Text style={[styles.title, { color: colors.black }]}>
              {title}
            </Text>
            <TouchableOpacity
              onPress={onClose}
              style={styles.closeButton}
            >
              <Image
                source={require('@/assets/images/event_card/back.png')}
                style={[styles.closeIcon, { tintColor: colors.black }]}
              />
            </TouchableOpacity>
          </View>
          
          <TextInput
            style={[
              styles.textArea,
              {
                borderColor: colors.grey,
                color: colors.black
              }
            ]}
            placeholder={placeholder}
            placeholderTextColor={colors.grey}
            value={description}
            onChangeText={onChangeDescription}
            multiline
            textAlignVertical="top"
            maxLength={maxLength}
            autoFocus
          />
          
          <Text style={[styles.counter, { color: colors.grey }]}>
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
  },
  closeButton: {
    padding: 4,
  },
  closeIcon: {
    width: 24,
    height: 24,
    transform: [{ rotate: '180deg' }],
  },
  textArea: {
    borderWidth: 1,
    borderRadius: 12,
    padding: 16,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    minHeight: 230,
    textAlignVertical: 'top',
  },
  counter: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    textAlign: 'right',
    marginTop: 8,
  },
});

export default DescriptionModal;