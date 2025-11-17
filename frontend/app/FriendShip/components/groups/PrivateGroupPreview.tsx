import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
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
}

const PrivateGroupPreview: React.FC<PrivateGroupPreviewProps> = ({
  groupName,
  onRequestJoin,
  isProcessing,
}) => {
  return (
    <View style={styles.container}>
      <View style={styles.iconContainer}>
        <Image
          source={require('@/assets/images/groups/lock.png')}
          style={styles.lockIcon}
          resizeMode="contain"
        />
      </View>

      <Text style={styles.title}>Приватная группа</Text>
      
      {groupName && (
        <Text style={styles.groupName}>{groupName}</Text>
      )}

      <Text style={styles.description}>
        Это приватная группа.{'\n'}
        Чтобы увидеть её содержимое, отправьте заявку на вступление.
      </Text>

      <TouchableOpacity
        style={[styles.requestButton, isProcessing && styles.requestButtonDisabled]}
        onPress={onRequestJoin}
        disabled={isProcessing}
      >
        {isProcessing ? (
          <ActivityIndicator color={Colors.white} size="small" />
        ) : (
          <Text style={styles.requestButtonText}>Подать заявку</Text>
        )}
      </TouchableOpacity>

      <Text style={styles.hint}>
        После одобрения заявки администратором вы получите полный доступ к группе
      </Text>
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
    backgroundColor: Colors.veryLightGrey,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 24,
  },
  lockIcon: {
    width: 60,
    height: 60,
    tintColor: Colors.grey,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 24,
    color: Colors.black,
    marginBottom: 8,
    textAlign: 'center',
  },
  groupName: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.lightBlue,
    marginBottom: 16,
    textAlign: 'center',
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
    textAlign: 'center',
    lineHeight: 24,
    marginBottom: 32,
  },
  requestButton: {
    backgroundColor: Colors.lightBlue3,
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
    color: Colors.grey,
    textAlign: 'center',
    fontStyle: 'italic',
  },
});

export default PrivateGroupPreview;