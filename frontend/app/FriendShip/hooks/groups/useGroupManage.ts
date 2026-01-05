import { useCallback, useEffect, useState } from 'react';
import { TabType } from './groupManageTypes';
import { useGroupData } from './useGroupData';
import { useGroupEvents } from './useGroupEvents';
import { useGroupRequests } from './useGroupRequests';
import { useGroupSubscribers } from './useGroupSubscribers';

export function useGroupManage(groupId: string) {
  const [activeTab, setActiveTab] = useState<TabType>('info');
  const [searchQuery, setSearchQuery] = useState('');
  const [contactsModalVisible, setContactsModalVisible] = useState(false);

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

  const groupDataHook = useGroupData(groupId, { showToast });
  const requestsHook = useGroupRequests(groupId, { showToast });
  const eventsHook = useGroupEvents(groupId, groupDataHook.groupData, { showToast });
  const subscribersHook = useGroupSubscribers(groupId, groupDataHook.groupData?.subscribers, { showToast });

  useEffect(() => {
    if (groupId) {
      groupDataHook.loadGroupData();
      requestsHook.loadGroupRequests();
      eventsHook.loadGenres();
    }
  }, [groupId]);

  const getSectionTitle = () => {
    switch (activeTab) {
      case 'info': return 'Основная информация';
      case 'subscribers': return 'Участники группы';
      case 'requests': return 'Управление заявками';
      case 'events': return 'Управление событиями';
      default: return 'Основная информация';
    }
  };

  const toggleCategory = (categoryId: string) => {
    groupDataHook.setSelectedCategories(prev => 
      prev.includes(categoryId) 
        ? prev.filter(id => id !== categoryId)
        : [...prev, categoryId]
    );
  };

  const handleContactsPress = () => {
    setContactsModalVisible(true);
  };

  const handleContactsSave = (contacts: any) => {
    groupDataHook.setSelectedContacts(contacts);
    setContactsModalVisible(false);
  };

  const handleDeleteGroup = async () => {
    const success = await groupDataHook.deleteGroup();
    if (success) {
      return true;
    }
    return false;
  };

  return {
    activeTab,
    setActiveTab,
    getSectionTitle,

    ...groupDataHook,
    toggleCategory,
    handleContactsPress,
    handleContactsSave,
    contactsModalVisible,
    setContactsModalVisible,
    handleSaveChanges: groupDataHook.saveGroupChanges,

    ...requestsHook,
    searchQuery,
    setSearchQuery,

    ...eventsHook,
    handleCreateEventSave: (eventData: any) => 
      eventsHook.handleCreateEventSave(eventData, groupDataHook.loadGroupData),
    ...subscribersHook,
      handleRemoveMemberPress: subscribersHook.handleRemoveMemberPress,
      handleConfirmRemoveMember: () =>
        subscribersHook.handleConfirmRemoveMember(groupDataHook.loadGroupData),
      handleCancelRemoveMember: subscribersHook.handleCancelRemoveMember,
    handleEditEventSave: (eventId: string, eventData: any) => 
      eventsHook.handleEditEventSave(eventId, eventData, groupDataHook.loadGroupData),
    handleDeleteEvent: (eventId: string) =>
      eventsHook.handleDeleteEvent(eventId, groupDataHook.loadGroupData),

    handleDeleteGroup,

    toastVisible,
    toastType,
    toastTitle,
    toastMessage,
    setToastVisible,
  };
}