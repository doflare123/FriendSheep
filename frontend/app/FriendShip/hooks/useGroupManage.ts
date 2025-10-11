import { TabType } from '@/components/groups/management/GroupManageTabPanel';
import { getGroupData } from '@/data/groupsData';
import { useEffect, useState } from 'react';
import { Alert } from 'react-native';

export interface RequestItem {
  id: string;
  name: string;
  username: string;
  imageUri: string;
}

export interface ConfirmationModalState {
  visible: boolean;
  action: string;
  title: string;
  message: string;
}

export const useGroupManage = (groupId: string) => {
  const [activeTab, setActiveTab] = useState<TabType>('info');

  const [groupName, setGroupName] = useState('');
  const [shortDescription, setShortDescription] = useState('');
  const [fullDescription, setFullDescription] = useState('');
  const [country, setCountry] = useState('');
  const [city, setCity] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [selectedContacts, setSelectedContacts] = useState<any[]>([]);
  const [groupImage, setGroupImage] = useState<string>('');

  const [searchQuery, setSearchQuery] = useState('');
  const [confirmationModal, setConfirmationModal] = useState<ConfirmationModalState>({
    visible: false,
    action: '',
    title: '',
    message: ''
  });
  
  const [contactsModalVisible, setContactsModalVisible] = useState(false);

  const [createEventModalVisible, setCreateEventModalVisible] = useState(false);
  const [editEventModalVisible, setEditEventModalVisible] = useState(false);
  const [selectedEventId, setSelectedEventId] = useState<string>('');

  const pendingRequests: RequestItem[] = [
    {
      id: '6',
      name: 'Ассемблер',
      username: '@doflare',
      imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
    },
    {
      id: '7',
      name: 'Ассемблер',
      username: '@doflare',
      imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
    }
  ];

  const groupData = getGroupData(groupId);

  useEffect(() => {
    if (groupData) {
      setGroupName(groupData.name);
      setShortDescription(groupData.description.substring(0, 100));
      setFullDescription(groupData.description);
      setCountry(groupData.country);
      setCity(groupData.city);
      setIsPrivate(false);
      setSelectedCategories(groupData.categories);
      setSelectedContacts(groupData.contacts);
      setGroupImage(groupData.imageUri);
    }
  }, [groupData]);

  const handleAcceptRequest = (userId: string) => {
    console.log('Accepting request for user:', userId);
    Alert.alert('Успешно', 'Заявка принята!');
  };

  const handleRejectRequest = (userId: string) => {
    console.log('Rejecting request for user:', userId);
    Alert.alert('Успешно', 'Заявка отклонена!');
  };

  const handleAcceptAll = () => {
    setConfirmationModal({
      visible: true,
      action: 'acceptAll',
      title: 'Принять все заявки?',
      message: 'Вы действительно хотите принять все заявки на вступление в группу?'
    });
  };

  const handleRejectAll = () => {
    setConfirmationModal({
      visible: true,
      action: 'rejectAll',
      title: 'Отклонить все заявки?',
      message: 'Вы действительно хотите отклонить все заявки на вступление в группу?'
    });
  };

  const confirmAction = () => {
    if (confirmationModal.action === 'acceptAll') {
      console.log('Accepting all requests');
      Alert.alert('Успешно', 'Все заявки приняты!');
    } else if (confirmationModal.action === 'rejectAll') {
      console.log('Rejecting all requests');
      Alert.alert('Успешно', 'Все заявки отклонены!');
    }
    setConfirmationModal({ visible: false, action: '', title: '', message: '' });
  };

  const cancelConfirmation = () => {
    setConfirmationModal({ visible: false, action: '', title: '', message: '' });
  };

  const toggleCategory = (categoryId: string) => {
    setSelectedCategories(prev => 
      prev.includes(categoryId) 
        ? prev.filter(id => id !== categoryId)
        : [...prev, categoryId]
    );
  };

  const handleContactsPress = () => {
    setContactsModalVisible(true);
  };

  const handleContactsSave = (contacts: any[]) => {
    setSelectedContacts(contacts);
    setContactsModalVisible(false);
  };

  const handleSaveChanges = () => {
    const updatedData = {
      name: groupName,
      shortDescription,
      fullDescription,
      country,
      city,
      isPrivate,
      categories: selectedCategories,
      contacts: selectedContacts,
      imageUri: groupImage,
    };
    
    console.log('Saving group changes:', updatedData);
    Alert.alert('Успешно', 'Изменения сохранены!');
  };

  const handleCreateEvent = () => {
    setCreateEventModalVisible(true);
  };

  const handleEditEvent = (eventId: string) => {
    setSelectedEventId(eventId);
    setEditEventModalVisible(true);
  };

  const handleCreateEventSave = (eventData: any) => {
    console.log('Creating new event:', eventData);
    setCreateEventModalVisible(false);
    Alert.alert('Успешно', 'Событие создано!');
  };

  const handleEditEventSave = (eventData: any) => {
    console.log('Editing event:', selectedEventId, eventData);
    setEditEventModalVisible(false);
    setSelectedEventId('');
    Alert.alert('Успешно', 'Событие обновлено!');
  };

  const getSectionTitle = () => {
    switch (activeTab) {
      case 'info':
        return 'Основная информация';
      case 'requests':
        return 'Управление заявками';
      case 'events':
        return 'Управление событиями';
      default:
        return 'Основная информация';
    }
  };

  return {
    activeTab,
    setActiveTab,

    groupData,
    groupName,
    setGroupName,
    shortDescription,
    setShortDescription,
    fullDescription,
    setFullDescription,
    country,
    setCountry,
    city,
    setCity,
    isPrivate,
    setIsPrivate,
    selectedCategories,
    setSelectedContacts,
    selectedContacts,
    groupImage,
    setGroupImage,
    
    toggleCategory,
    handleContactsPress,
    handleContactsSave,
    contactsModalVisible,
    setContactsModalVisible,
    handleSaveChanges,

    pendingRequests,
    searchQuery,
    setSearchQuery,
    handleAcceptRequest,
    handleRejectRequest,
    handleAcceptAll,
    handleRejectAll,

    confirmationModal,
    confirmAction,
    cancelConfirmation,

    createEventModalVisible,
    setCreateEventModalVisible,
    editEventModalVisible,
    setEditEventModalVisible,
    selectedEventId,
    handleCreateEvent,
    handleEditEvent,
    handleCreateEventSave,
    handleEditEventSave,

    getSectionTitle,
  };
};