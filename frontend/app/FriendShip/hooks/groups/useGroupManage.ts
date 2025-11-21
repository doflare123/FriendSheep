import { useEffect, useState } from 'react';
import { TabType } from './groupManageTypes';
import { useGroupData } from './useGroupData';
import { useGroupEvents } from './useGroupEvents';
import { useGroupRequests } from './useGroupRequests';

export function useGroupManage(groupId: string) {
  const [activeTab, setActiveTab] = useState<TabType>('info');
  const [searchQuery, setSearchQuery] = useState('');
  const [contactsModalVisible, setContactsModalVisible] = useState(false);

  const groupDataHook = useGroupData(groupId);
  const requestsHook = useGroupRequests(groupId);
  const eventsHook = useGroupEvents(groupId, groupDataHook.groupData);

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
    handleEditEventSave: (eventId: string, eventData: any) => 
      eventsHook.handleEditEventSave(eventId, eventData, groupDataHook.loadGroupData),
  };
}