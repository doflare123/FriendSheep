import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useEffect } from 'react';
import {
  BackHandler,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';

interface ConfirmationModalProps {
  visible: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  link?: string;
  isBlocked?: boolean;
}

const ConfirmationModal: React.FC<ConfirmationModalProps> = ({
  visible,
  title,
  message,
  onConfirm,
  onCancel,
  link,
  isBlocked = false,
}) => {
  const colors = useThemedColors();

   useEffect(() => {
    if (!visible) return;

    const backHandler = BackHandler.addEventListener('hardwareBackPress', () => {
      onCancel();
      return true;
    });

    return () => backHandler.remove();
  }, [visible]);

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
          
          {isBlocked && (
            <View style={[styles.warningBox, { backgroundColor: Colors.red + '15' }]}>
              <Text style={[styles.warningText, { color: Colors.red }]}>
                Этот сервис заблокирован на территории РФ
              </Text>
            </View>
          )}

          <Text style={[styles.modalMessage, { color: colors.grey }]}>
            {message}
            {link && (
              <>
                  <Text 
                    style={[styles.linkText, { color: Colors.lightBlue }]}
                    selectable
                  >
                    {link}
                  </Text>
              </>
            )}
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
              <Text style={styles.modalButtonText}>
                {link ? 'Перейти' : 'Подтвердить'}
              </Text>
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
  warningBox: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 12,
    borderRadius: 12,
    marginBottom: 12,
  },
  warningText: {
    textAlign: 'center',
    fontFamily: Montserrat.bold,
    fontSize: 12,
    flex: 1,
  },
  modalMessage: {
    fontFamily: Montserrat.regular,
    fontSize: 15,
    textAlign: 'center',
    marginBottom: 12,
  },
  linkContainer: {
    maxHeight: 100,
    padding: 12,
    borderRadius: 8,
    marginBottom: 12,
  },
  linkText: {
    fontFamily: Montserrat.regular,
    fontSize: 13,
    lineHeight: 18,
  },
  infoText: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    textAlign: 'center',
    marginBottom: 8,
    fontStyle: 'italic',
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