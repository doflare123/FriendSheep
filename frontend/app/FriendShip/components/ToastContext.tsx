import Toast from '@/components/Toast';
import React, { createContext, useContext, useState } from 'react';
import { Modal, StyleSheet, View } from 'react-native';

interface ToastData {
  type?: 'success' | 'error' | 'warning';
  title: string;
  message: string;
}

interface ToastContextType {
  showToast: (data: ToastData) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [toastVisible, setToastVisible] = useState(false);
  const [toastData, setToastData] = useState<ToastData>({
    type: 'success',
    title: '',
    message: '',
  });

  const showToast = (data: ToastData) => {
    setToastData(data);
    setToastVisible(true);
  };

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <Modal
        visible={toastVisible}
        transparent={true}
        animationType="none"
        statusBarTranslucent
        onRequestClose={() => setToastVisible(false)}
      >
        <View style={styles.modalContainer} pointerEvents="box-none">
          <Toast
            visible={toastVisible}
            type={toastData.type}
            title={toastData.title}
            message={toastData.message}
            onHide={() => setToastVisible(false)}
          />
        </View>
      </Modal>
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
};

const styles = StyleSheet.create({
  modalContainer: {
    flex: 1,
    backgroundColor: 'transparent',
  },
});