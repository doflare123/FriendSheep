import authService from '@/api/services/authService';
import pushNotificationService from '@/api/services/pushNotificationService';
import { useAuthContext } from '@/components/auth/AuthContext';
import BottomBar from '@/components/BottomBar';
import ConfirmationModal from '@/components/ConfirmationModal';
import PageHeader from '@/components/PageHeader';
import { useTheme } from '@/components/ThemeContext';
import Toast from '@/components/Toast';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useCallback, useState } from 'react';
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

const SettingsPage = () => {
  const { sortingState, sortingActions } = useSearchState();
  const { setIsAuthenticated } = useAuthContext();
  const { toggleTheme, isDark } = useTheme();
  const colors = useThemedColors();
  const [showLogoutModal, setShowLogoutModal] = useState(false);

  const [toastVisible, setToastVisible] = useState(false);
  const [toastType, setToastType] = useState<'success' | 'error' | 'warning'>('success');
  const [toastTitle, setToastTitle] = useState('');
  const [toastMessage, setToastMessage] = useState('');

  const showToast = useCallback((type: 'success' | 'error' | 'warning', title: string, message: string) => {
    setToastType(type);
    setToastTitle(title);
    setToastMessage(message);
    setToastVisible(true);
  }, []);

  const handleLogoutPress = () => {
    setShowLogoutModal(true);
  };

  const handleConfirmLogout = async () => {
    try {
      console.log('[Settings] 🚪 Начало выхода из аккаунта');
      
      await pushNotificationService.removeTokenFromServer();
      await pushNotificationService.setBadgeCount(0);
      await authService.logout();
      
      console.log('[Settings] ✅ Выход выполнен успешно');
      
      setShowLogoutModal(false);

      setIsAuthenticated(false);
      
      console.log('[Settings] ✅ Состояние аутентификации обновлено');
      
    } catch (error) {
      console.error('[Settings] ❌ Ошибка выхода:', error);
      setShowLogoutModal(false);
      showToast('error', 'Ошибка', 'Не удалось выйти из аккаунта');
    }
  };

  const handleCancelLogout = () => {
    setShowLogoutModal(false);
  };

  const handleThemeToggle = () => {
    toggleTheme();
    showToast('success', 'Тема изменена', `Применена ${isDark ? 'светлая' : 'тёмная'} тема`);
  };

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: colors.white }]}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <PageHeader title="Настройки" showWave />

        <View style={styles.shadowWrapper}>
          <TouchableOpacity 
            style={[styles.badge, { backgroundColor: colors.card }]} 
            onPress={handleThemeToggle}
          >
            <Text style={[styles.settingText, { color: colors.black }]}>
              {isDark ? 'Светлая тема' : 'Тёмная тема'}
            </Text>
            <Image
              source={
                isDark
                  ? require('@/assets/images/settings/light.png')
                  : require('@/assets/images/settings/dark.png')
              }
              style={styles.icon}
            />
          </TouchableOpacity>
        </View>

        <View style={styles.shadowWrapper}>
          <TouchableOpacity 
            style={[styles.badge, { backgroundColor: colors.card }]} 
            onPress={handleLogoutPress}
          >
            <Text style={styles.logout}>Выйти из аккаунта</Text>
            <Image
              source={require('@/assets/images/settings/logout.png')}
              style={styles.icon}
            />
          </TouchableOpacity>
        </View>
      </ScrollView>
      <BottomBar />

      <ConfirmationModal
        visible={showLogoutModal}
        title="Выход из аккаунта"
        message="Вы уверены, что хотите выйти из аккаунта?"
        onConfirm={handleConfirmLogout}
        onCancel={handleCancelLogout}
      />

      <Toast
        visible={toastVisible}
        type={toastType}
        title={toastTitle}
        message={toastMessage}
        onHide={() => setToastVisible(false)}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  scrollView: {
    flex: 1,
  },
  settingText: {
    fontFamily: Montserrat.regular,
    fontSize: 20,
    flex: 1,
  },
  logout: {
    fontFamily: Montserrat.regular,
    fontSize: 20,
    color: Colors.red,
    flex: 1,
  },
  shadowWrapper: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 6,
  },
  badge: {
    padding: 16,
    elevation: 2,
    borderRadius: 20,
    margin: 16,
    marginBottom: 0,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center'
  },
  icon: {
    resizeMode: 'contain',
    width: 30,
    height: 30,
  }
});

export default SettingsPage;