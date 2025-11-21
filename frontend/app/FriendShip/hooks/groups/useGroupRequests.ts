import groupService, { SimpleGroupRequest } from '@/api/services/groupService';
import { useMemo, useState } from 'react';
import { Alert } from 'react-native';
import { normalizeImageUrl } from './groupManageHelpers';
import { ConfirmationModalState, RequestItem } from './groupManageTypes';

export function useGroupRequests(groupId: string) {
  const [groupRequests, setGroupRequests] = useState<SimpleGroupRequest[]>([]);
  const [isLoadingRequests, setIsLoadingRequests] = useState(false);
  const [confirmationModal, setConfirmationModal] = useState<ConfirmationModalState>({
    visible: false,
    action: '',
    title: '',
    message: ''
  });

  const loadGroupRequests = async () => {
    try {
      setIsLoadingRequests(true);
      const requests = await groupService.getGroupRequests(parseInt(groupId));
      setGroupRequests(requests);
    } catch (error: any) {
      console.error('[useGroupRequests] Ошибка загрузки заявок:', error);
      setGroupRequests([]);
    } finally {
      setIsLoadingRequests(false);
    }
  };

  const pendingRequests: RequestItem[] = useMemo(() => {
    if (!Array.isArray(groupRequests)) {
      return [];
    }

    return groupRequests.map(req => ({
      id: req.id.toString(),
      name: req.name,
      username: req.us,
      imageUri: normalizeImageUrl(req.image),
    }));
  }, [groupRequests]);

  const handleAcceptRequest = async (requestId: string) => {
    try {
      await groupService.approveRequest(parseInt(requestId));
      Alert.alert('Успешно', 'Заявка принята!');
      await loadGroupRequests();
    } catch (error: any) {
      Alert.alert('Ошибка', error.message || 'Не удалось принять заявку');
    }
  };

  const handleRejectRequest = async (requestId: string) => {
    try {
      await groupService.rejectRequest(parseInt(requestId));
      Alert.alert('Успешно', 'Заявка отклонена!');
      await loadGroupRequests();
    } catch (error: any) {
      Alert.alert('Ошибка', error.message || 'Не удалось отклонить заявку');
    }
  };

  const handleAcceptAll = () => {
    if (pendingRequests.length === 0) {
      Alert.alert('Внимание', 'Нет заявок для обработки');
      return;
    }

    setConfirmationModal({
      visible: true,
      action: 'acceptAll',
      title: 'Принять все заявки?',
      message: `Вы действительно хотите принять все заявки (${pendingRequests.length})?`
    });
  };

  const handleRejectAll = () => {
    if (pendingRequests.length === 0) {
      Alert.alert('Внимание', 'Нет заявок для обработки');
      return;
    }

    setConfirmationModal({
      visible: true,
      action: 'rejectAll',
      title: 'Отклонить все заявки?',
      message: `Вы действительно хотите отклонить все заявки (${pendingRequests.length})?`
    });
  };

  const confirmAction = async () => {
    try {
      if (confirmationModal.action === 'acceptAll') {
        await groupService.approveAllRequests(parseInt(groupId));
        Alert.alert('Успешно', 'Все заявки приняты!');
      } else if (confirmationModal.action === 'rejectAll') {
        await groupService.rejectAllRequests(parseInt(groupId));
        Alert.alert('Успешно', 'Все заявки отклонены!');
      }
      await loadGroupRequests();
    } catch (error: any) {
      Alert.alert('Ошибка', error.message || 'Не удалось обработать заявки');
    } finally {
      setConfirmationModal({ visible: false, action: '', title: '', message: '' });
    }
  };

  const cancelConfirmation = () => {
    setConfirmationModal({ visible: false, action: '', title: '', message: '' });
  };

  return {
    pendingRequests,
    isLoadingRequests,
    confirmationModal,
    loadGroupRequests,
    handleAcceptRequest,
    handleRejectRequest,
    handleAcceptAll,
    handleRejectAll,
    confirmAction,
    cancelConfirmation,
  };
}