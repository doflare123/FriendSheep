import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import {
    Image,
    Modal,
    ScrollView,
    StyleSheet,
    Text,
    TouchableOpacity,
    View
} from 'react-native';

interface PermissionsModalProps {
  visible: boolean;
  onRequestPermissions: () => void;
  onSkip?: () => void;
}

const PermissionsModal: React.FC<PermissionsModalProps> = ({
  visible,
  onRequestPermissions,
  onSkip,
}) => {
  const colors = useThemedColors();

  return (
    <Modal
      visible={visible}
      animationType="fade"
      transparent
      statusBarTranslucent
    >
      <View style={styles.overlay}>
        <View style={[styles.container, { backgroundColor: colors.white }]}>
          <ScrollView 
            showsVerticalScrollIndicator={false}
            contentContainerStyle={styles.scrollContent}
          >
            <View style={styles.header}>
              <Image
                source={require('@/assets/images/logo.png')}
                style={styles.logo}
              />
              <Text style={[styles.title, { color: colors.black }]}>
                Добро пожаловать!
              </Text>
              <Text style={[styles.subtitle, { color: colors.grey }]}>
                Для полноценной работы приложению нужны некоторые разрешения
              </Text>
            </View>

            <View style={styles.permissionsContainer}>
              <View style={styles.permissionItem}>
                <View style={[styles.iconContainer, { backgroundColor: colors.lightBlue }]}>
                  <Text style={styles.iconText}>📷</Text>
                </View>
                <View style={styles.permissionText}>
                  <Text style={[styles.permissionTitle, { color: colors.black }]}>
                    Доступ к фото
                  </Text>
                  <Text style={[styles.permissionDescription, { color: colors.grey }]}>
                    Для загрузки аватара, обложек событий и изображений групп
                  </Text>
                </View>
              </View>

              <View style={styles.permissionItem}>
                <View style={[styles.iconContainer, { backgroundColor: colors.lightBlue }]}>
                  <Text style={styles.iconText}>🔔</Text>
                </View>
                <View style={styles.permissionText}>
                  <Text style={[styles.permissionTitle, { color: colors.black }]}>
                    Уведомления
                  </Text>
                  <Text style={[styles.permissionDescription, { color: colors.grey }]}>
                    Чтобы не пропустить приглашения в группы и напоминания о событиях
                  </Text>
                </View>
              </View>
            </View>

            <Text style={[styles.note, { color: colors.lightGrey }]}>
              Вы всегда можете изменить разрешения в настройках устройства
            </Text>
          </ScrollView>

          <View style={styles.buttonsContainer}>
            <TouchableOpacity
              style={[styles.primaryButton, { backgroundColor: colors.lightBlue }]}
              onPress={onRequestPermissions}
              activeOpacity={0.8}
            >
              <Text style={styles.primaryButtonText}>Разрешить доступ</Text>
            </TouchableOpacity>

            {onSkip && (
              <TouchableOpacity
                style={styles.skipButton}
                onPress={onSkip}
                activeOpacity={0.8}
              >
                <Text style={[styles.skipButtonText, { color: colors.grey }]}>
                  Пропустить
                </Text>
              </TouchableOpacity>
            )}
          </View>
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  container: {
    width: '85%',
    maxHeight: '80%',
    borderRadius: 20,
    overflow: 'hidden',
  },
  scrollContent: {
    padding: 24,
  },
  header: {
    alignItems: 'center',
    marginBottom: 32,
  },
  logo: {
    width: 80,
    height: 80,
    marginBottom: 16,
    borderRadius: 20,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 24,
    marginBottom: 8,
    textAlign: 'center',
  },
  subtitle: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    textAlign: 'center',
    lineHeight: 20,
  },
  permissionsContainer: {
    marginBottom: 24,
  },
  permissionItem: {
    flexDirection: 'row',
    marginBottom: 20,
    alignItems: 'flex-start',
  },
  iconContainer: {
    width: 50,
    height: 50,
    borderRadius: 25,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  iconText: {
    fontSize: 24,
  },
  permissionText: {
    flex: 1,
    justifyContent: 'center',
  },
  permissionTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    marginBottom: 4,
  },
  permissionDescription: {
    fontFamily: Montserrat.regular,
    fontSize: 13,
    lineHeight: 18,
  },
  note: {
    fontFamily: Montserrat.regular,
    fontSize: 11,
    textAlign: 'center',
    fontStyle: 'italic',
  },
  buttonsContainer: {
    padding: 24,
    paddingTop: 0,
  },
  primaryButton: {
    paddingVertical: 14,
    borderRadius: 25,
    alignItems: 'center',
    marginBottom: 12,
  },
  primaryButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
  skipButton: {
    paddingVertical: 12,
    alignItems: 'center',
  },
  skipButtonText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
});

export default PermissionsModal;