import groupService, { GroupDetailResponse, SimpleGroupRequest } from '@/api/services/groupService';
import { TabType } from '@/components/groups/management/GroupManageTabPanel';
import { Contact } from '@/components/groups/modal/ContactsModal';
// eslint-disable-next-line import/no-unresolved
import { LOCAL_IP } from '@env';
 
import { useEffect, useMemo, useState } from 'react';
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

const CATEGORY_MAPPING: { [key: string]: string } = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};

const CATEGORY_IDS: { [key: string]: number } = {
  'movie': 1,
  'game': 2,
  'table_game': 3,
  'other': 4,
};

export const useGroupManage = (groupId: string) => {
  const [activeTab, setActiveTab] = useState<TabType>('info');
  const [groupData, setGroupData] = useState<GroupDetailResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  const [groupName, setGroupName] = useState('');
  const [shortDescription, setShortDescription] = useState('');
  const [fullDescription, setFullDescription] = useState('');
  const [country, setCountry] = useState('');
  const [city, setCity] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [selectedContacts, setSelectedContacts] = useState<Contact[]>([]);
  const [groupImage, setGroupImage] = useState<string>('');
  const [originalImage, setOriginalImage] = useState<string>('');

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

  const [isProcessingRequests, setIsProcessingRequests] = useState(false);
  const [isLoadingRequests, setIsLoadingRequests] = useState(false);

  const [groupRequests, setGroupRequests] = useState<SimpleGroupRequest[]>([]);

  const loadGroupData = async () => {
    try {
      setIsLoading(true);
      const data = await groupService.getGroupDetail(groupId);

      if (data.image && data.image.includes('localhost')) {
        data.image = data.image.replace('http://localhost:8080', 'http://' + LOCAL_IP + ':8080');
      }
      
      setGroupData(data);

      setGroupName(data.name);
      setShortDescription(data.small_description);
      setFullDescription(data.description);
      setCity(data.city);
      setIsPrivate(data.private);
      setGroupImage(data.image);
      setOriginalImage(data.image);

      const mappedCategories = data.categories
        .map(cat => CATEGORY_MAPPING[cat])
        .filter(cat => cat !== undefined);
      setSelectedCategories(mappedCategories);

      const contactIcons = {
        discord: require('@/assets/images/groups/contacts/discord.png'),
        vk: require('@/assets/images/groups/contacts/vk.png'),
        telegram: require('@/assets/images/groups/contacts/telegram.png'),
        twitch: require('@/assets/images/groups/contacts/twitch.png'),
        youtube: require('@/assets/images/groups/contacts/youtube.png'),
        whatsapp: require('@/assets/images/groups/contacts/whatsapp.png'),
        max: require('@/assets/images/groups/contacts/max.png'),
      };
      
      const getContactType = (name: string, link: string) => {
        const lowerLink = link.toLowerCase();
        const lowerName = name.toLowerCase();
        
        if (lowerLink.includes('discord') || lowerName.includes('discord')) return 'discord';
        if (lowerLink.includes('vk.com') || lowerName.includes('vk')) return 'vk';
        if (lowerLink.includes('t.me') || lowerLink.includes('telegram') || lowerName.includes('telegram')) return 'telegram';
        if (lowerLink.includes('twitch') || lowerName.includes('twitch')) return 'twitch';
        if (lowerLink.includes('youtube') || lowerLink.includes('youtu')) return 'youtube';
        if (lowerLink.includes('wa.me') || lowerLink.includes('whatsapp') || lowerName.includes('whatsapp')) return 'whatsapp';
        
        return 'max';
      };
      
      const mappedContacts = (data.contacts || []).map(contact => {
        const type = getContactType(contact.name, contact.link);
        return {
          id: type,
          name: type,
          icon: contactIcons[type],
          description: contact.name,
          link: contact.link
        } as Contact;
      });
      
      setSelectedContacts(mappedContacts);
      
    } catch (error: any) {
      console.error('Ошибка загрузки данных группы:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось загрузить данные группы');
    } finally {
      setIsLoading(false);
    }
  };

    useEffect(() => {
    if (groupId) {
      console.log('[useGroupManage] Инициализация, groupId:', groupId);
      loadGroupData();
      loadGroupRequests();
    }
  }, [groupId]);

  const loadGroupRequests = async () => {
    try {
      setIsLoadingRequests(true);
      console.log('[useGroupManage] 🔄 Начинаем загрузку заявок для группы:', groupId);
      
      const requests = await groupService.getGroupRequests(parseInt(groupId));
      
      console.log('[useGroupManage] 📦 Получены заявки (тип):', typeof requests);
      console.log('[useGroupManage] 📦 Получены заявки (массив?):', Array.isArray(requests));
      console.log('[useGroupManage] 📦 Получены заявки (длина):', requests?.length);
      console.log('[useGroupManage] 📦 Получены заявки (данные):', JSON.stringify(requests, null, 2));
      
      setGroupRequests(requests);
      
      console.log('[useGroupManage] ✅ Заявки сохранены в state');
    } catch (error: any) {
      console.error('[useGroupManage] ❌ Ошибка загрузки заявок:', error);
      console.error('[useGroupManage] ❌ Детали:', error.response?.data);
      setGroupRequests([]);
    } finally {
      setIsLoadingRequests(false);
    }
  };

  const pendingRequests: RequestItem[] = useMemo(() => {
    console.log('[useGroupManage] 🎯 === ФОРМИРОВАНИЕ PENDING REQUESTS ===');
    console.log('[useGroupManage] 🎯 groupRequests:', groupRequests);
    console.log('[useGroupManage] 🎯 Тип groupRequests:', typeof groupRequests);
    console.log('[useGroupManage] 🎯 Это массив?:', Array.isArray(groupRequests));
    console.log('[useGroupManage] 🎯 Длина:', groupRequests?.length);
    
    if (!Array.isArray(groupRequests)) {
      console.error('[useGroupManage] ❌ groupRequests не является массивом!');
      return [];
    }

    const pending = groupRequests.map(req => {
      let imageUri = req.image;
      if (imageUri && imageUri.includes('localhost')) {
        imageUri = imageUri.replace('http://localhost:8080', 'http://' + LOCAL_IP + ':8080');
      }
      
      const mapped = {
        id: req.id.toString(),
        name: req.name,
        username: req.us,
        imageUri: imageUri,
      };
      
      console.log('[useGroupManage] ✅ Преобразованная заявка:', mapped);
      
      return mapped;
    });
    
    console.log('[useGroupManage] 🎯 === РЕЗУЛЬТАТ ===');
    console.log('[useGroupManage] 🎯 Количество pending заявок:', pending.length);
    console.log('[useGroupManage] 🎯 Финальный список:', pending);
    
    return pending;
  }, [groupRequests]);

  const handleAcceptRequest = async (requestId: string) => {
    try {
      console.log('Принятие заявки:', requestId);
      await groupService.approveRequest(parseInt(requestId));
      
      Alert.alert('Успешно', 'Заявка принята!');

      await loadGroupRequests();
    } catch (error: any) {
      console.error('Ошибка принятия заявки:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось принять заявку');
    }
  };

  const handleRejectRequest = async (requestId: string) => {
    try {
      console.log('Отклонение заявки:', requestId);
      await groupService.rejectRequest(parseInt(requestId));
      
      Alert.alert('Успешно', 'Заявка отклонена!');

      await loadGroupRequests();
    } catch (error: any) {
      console.error('Ошибка отклонения заявки:', error);
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
      message: `Вы действительно хотите принять все заявки (${pendingRequests.length}) на вступление в группу?`
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
      message: `Вы действительно хотите отклонить все заявки (${pendingRequests.length}) на вступление в группу?`
    });
  };

  const confirmAction = async () => {
    try {
      if (confirmationModal.action === 'acceptAll') {
        console.log('Принятие всех заявок');
        await groupService.approveAllRequests(parseInt(groupId));
        Alert.alert('Успешно', 'Все заявки приняты!');

        await loadGroupRequests();
        
      } else if (confirmationModal.action === 'rejectAll') {
        console.log('Отклонение всех заявок');
        await groupService.rejectAllRequests(parseInt(groupId));
        Alert.alert('Успешно', 'Все заявки отклонены!');

        await loadGroupRequests();
      }
    } catch (error: any) {
      console.error('Ошибка обработки заявок:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось обработать заявки');
    } finally {
      setConfirmationModal({ visible: false, action: '', title: '', message: '' });
    }
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

  const handleContactsSave = (contacts: Contact[]) => {
    setSelectedContacts(contacts);
    setContactsModalVisible(false);
  };

  const handleSaveChanges = async () => {
    try {
      setIsSaving(true);
      
      let imageUrl = groupImage;

      if (groupImage !== originalImage && !groupImage.startsWith('http')) {
        console.log('Загрузка нового изображения...');
        imageUrl = await groupService.uploadGroupPhoto(groupImage);
        console.log('Новый URL изображения:', imageUrl);
      }

      const categoryIds = selectedCategories.map(cat => CATEGORY_IDS[cat]);

      const contactsString = selectedContacts
        .filter(contact => contact.link && contact.link.trim() !== '')
        .map(contact => {
          const name = contact.description || contact.id || contact.name;
          return `${name}:${contact.link}`;
        })
        .join(', ');
      
      const updateData = {
        name: groupName,
        small_description: shortDescription,
        description: fullDescription,
        city: city,
        categories: categoryIds,
        is_private: isPrivate,
        image: imageUrl,
        contacts: contactsString || '',
      };
      
      console.log('Отправка данных на сервер:', updateData);
      
      await groupService.updateGroup(groupId, updateData);
      
      Alert.alert('Успешно', 'Изменения сохранены!');

      await loadGroupData();
      
    } catch (error: any) {
      console.error('Ошибка сохранения изменений:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось сохранить изменения');
    } finally {
      setIsSaving(false);
    }
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

  const formattedEvents = useMemo(() => {
    if (!groupData?.sessions) return [];
    
    return groupData.sessions.map(session => ({
      id: session.id.toString(),
      title: session.title,
      date: new Date(session.start_time).toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric'
      }),
      genres: session.genres || [],
      currentParticipants: session.current_users,
      maxParticipants: session.count_users_max,
      duration: `${session.duration} мин`,
      imageUri: session.image_url?.includes('localhost')
        ? session.image_url.replace('http://localhost:8080', 'http://' + LOCAL_IP + ':8080')
        : session.image_url,
      description: '',
      typeEvent: session.session_type,
      typePlace: session.session_place === 'offline' || session.session_place === 'online' 
        ? session.session_place as 'online' | 'offline'
        : 'online',
      eventPlace: session.city || '',
      publisher: groupData.name,
      publicationDate: session.start_time,
      ageRating: '',
      category: (selectedCategories[0] as any) || 'other',
      group: groupData.name,
    }));
  }, [groupData]);

  return {
    activeTab,
    setActiveTab,

    groupData,
    isLoading,
    isSaving,
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
    formattedEvents,
    isProcessingRequests,
    isLoadingRequests,
  };
};