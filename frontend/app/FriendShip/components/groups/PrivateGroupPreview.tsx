import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import {
  ActivityIndicator,
  Image,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';

interface PrivateGroupPreviewProps {
  groupName?: string;
  onRequestJoin: () => void;
  isProcessing: boolean;
  requestStatus?: 'none' | 'pending' | 'approved' | 'rejected';
}

const PrivateGroupPreview: React.FC<PrivateGroupPreviewProps> = ({
  groupName,
  onRequestJoin,
  isProcessing,
  requestStatus = 'none',
}) => {
  const colors = useThemedColors();

  const getButtonText = () => {
    switch (requestStatus) {
      case 'pending':
        return 'Заявка подана';
      case 'approved':
        return 'Заявка одобрена';
      case 'rejected':
        return 'Заявка отклонена';
      default:
        return 'Подать заявку';
    }
  };

  const getButtonStyle = () => {
    switch (requestStatus) {
      case 'pending':
        return [styles.requestButton, { backgroundColor: colors.grey }];
      case 'approved':
        return [styles.requestButton, { backgroundColor: Colors.green }];
      case 'rejected':
        return [styles.requestButton, { backgroundColor: Colors.red }];
      default:
        return [styles.requestButton, { backgroundColor: colors.lightBlue }];
    }
  };

  const isButtonDisabled = isProcessing || requestStatus === 'pending' || requestStatus === 'approved';

  return (
    <View style={styles.container}>
      <View style={[styles.iconContainer, { backgroundColor: colors.veryLightGrey }]}>
        <Image
          source={require('@/assets/images/groups/lock.png')}
          style={[styles.lockIcon, { tintColor: colors.grey }]}
          resizeMode="contain"
        />
      </View>

      <Text style={[styles.title, { color: colors.black }]}>
        Приватная группа
      </Text>
      
      {groupName && (
        <Text style={[styles.groupName, { color: colors.lightBlue }]}>
          {groupName}
        </Text>
      )}

      <Text style={[styles.description, { color: colors.grey }]}>
        {requestStatus === 'pending' 
          ? 'Ваша заявка отправлена и ожидает одобрения администратора.'
          : requestStatus === 'approved'
          ? 'Ваша заявка одобрена! Теперь у вас есть доступ к группе.'
          : requestStatus === 'rejected'
          ? 'Ваша заявка была отклонена. Вы можете подать заявку повторно.'
          : 'Это приватная группа.\nЧтобы увидеть её содержимое, отправьте заявку на вступление.'
        }
      </Text>

      <TouchableOpacity
        style={[getButtonStyle(), isButtonDisabled && styles.requestButtonDisabled]}
        onPress={onRequestJoin}
        disabled={isButtonDisabled}
      >
        {isProcessing ? (
          <ActivityIndicator color={Colors.white} size="small" />
        ) : (
          <Text style={styles.requestButtonText}>{getButtonText()}</Text>
        )}
      </TouchableOpacity>

      {requestStatus === 'none' && (
        <Text style={[styles.hint, { color: colors.grey }]}>
          После одобрения заявки администратором вы получите полный доступ к группе
        </Text>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
    paddingVertical: 40,
  },
  iconContainer: {
    width: 120,
    height: 120,
    borderRadius: 60,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 24,
  },
  lockIcon: {
    width: 60,
    height: 60,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 24,
    marginBottom: 8,
    textAlign: 'center',
  },
  groupName: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    marginBottom: 16,
    textAlign: 'center',
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    textAlign: 'center',
    lineHeight: 24,
    marginBottom: 32,
  },
  requestButton: {
    paddingVertical: 12,
    paddingHorizontal: 48,
    borderRadius: 25,
    marginBottom: 16,
    minWidth: 200,
    alignItems: 'center',
  },
  requestButtonDisabled: {
    opacity: 0.6,
  },
  requestButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
  hint: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    textAlign: 'center',
    fontStyle: 'italic',
  },
});

export default PrivateGroupPreview;