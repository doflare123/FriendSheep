import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import {
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';

interface ConfirmationModalProps {
  visible: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
}

const ConfirmationModal: React.FC<ConfirmationModalProps> = ({
  visible,
  title,
  message,
  onConfirm,
  onCancel,
}) => {
  const colors = useThemedColors();

  return (
    <Modal
      visible={visible}
      transparent={true}
      animationType="fade"
    >
      <View style={styles.modalOverlay}>
        <View style={[styles.modalContent, { backgroundColor: colors.white }]}>
          <Text style={[styles.modalTitle, { color: colors.black }]}>
            {title}
          </Text>
          <Text style={[styles.modalMessage, { color: colors.grey }]}>
            {message}
          </Text>
          
          <View style={styles.modalButtons}>
            <TouchableOpacity 
              style={[styles.modalButton, styles.cancelButton]}
              onPress={onCancel}
            >
              <Text style={styles.modalButtonText}>Отмена</Text>
            </TouchableOpacity>
            
            <TouchableOpacity 
              style={[
                styles.modalButton,
                { backgroundColor: colors.lightBlue }
              ]}
              onPress={onConfirm}
            >
              <Text style={styles.modalButtonText}>Подтвердить</Text>
            </TouchableOpacity>
          </View>
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  modalContent: {
    borderRadius: 40,
    padding: 20,
    marginHorizontal: 40,
  },
  modalTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    textAlign: 'center',
    marginBottom: 12,
  },
  modalMessage: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 24,
  },
  modalButtons: {
    flexDirection: 'row',
    gap: 12,
  },
  modalButton: {
    flex: 1,
    paddingVertical: 12,
    borderRadius: 40,
    alignItems: 'center',
  },
  cancelButton: {
    backgroundColor: Colors.red,
  },
  modalButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.white,
  },
});

export default ConfirmationModal;