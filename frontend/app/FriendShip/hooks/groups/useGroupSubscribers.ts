import groupMemberService from '@/api/services/group/groupMemberService';
import type { GroupSubscriber } from '@/api/services/group/groupTypes';
import { useMemo, useState } from 'react';
import { Alert } from 'react-native';
import { normalizeImageUrl } from './groupManageHelpers';

export function useGroupSubscribers(groupId: string, subscribers: GroupSubscriber[] | undefined) {
  const [subscriberSearchQuery, setSubscriberSearchQuery] = useState('');
  const [isProcessingSubscriber, setIsProcessingSubscriber] = useState(false);
  const [removeMemberModal, setRemoveMemberModal] = useState({
    visible: false,
    userId: '',
    userName: '',
  });

  const formattedSubscribers = useMemo(() => {
    if (!subscribers) return [];
    
    return subscribers.map(sub => ({
      id: sub.userId.toString(),
      name: sub.name,
      imageUri: normalizeImageUrl(sub.image),
      role: sub.role,
    }));
  }, [subscribers]);

  const handleRemoveMemberPress = (userId: string) => {
    const subscriber = formattedSubscribers.find(s => s.id === userId);
    setRemoveMemberModal({
      visible: true,
      userId: userId,
      userName: subscriber?.name || 'этого участника',
    });
  };

  const handleConfirmRemoveMember = async (onSuccess: () => void) => {
    try {
      setIsProcessingSubscriber(true);
      
      await groupMemberService.removeMember(parseInt(groupId), parseInt(removeMemberModal.userId));
      
      Alert.alert('Успешно', 'Участник удалён из группы');
      setRemoveMemberModal({ visible: false, userId: '', userName: '' });
      onSuccess();
    } catch (error: any) {
      console.error('[useGroupSubscribers] Ошибка удаления участника:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось удалить участника');
    } finally {
      setIsProcessingSubscriber(false);
    }
  };

  const handleCancelRemoveMember = () => {
    setRemoveMemberModal({ visible: false, userId: '', userName: '' });
  };

  const handleUserPress = (userId: string) => {
    return userId;
  };

  return {
    formattedSubscribers,
    subscriberSearchQuery,
    setSubscriberSearchQuery,
    isProcessingSubscriber,
    removeMemberModal,
    handleRemoveMemberPress,
    handleConfirmRemoveMember,
    handleCancelRemoveMember,
    handleUserPress,
  };
}