import groupService from '@/api/services/group/groupService';
import type { GroupDetailResponse } from '@/api/services/group/groupTypes';
import { Contact } from '@/components/groups/modal/ContactsModal';
import { useState } from 'react';
import { Alert } from 'react-native';
import { getContactType, normalizeImageUrl } from './groupManageHelpers';
import { CATEGORY_IDS, CATEGORY_MAPPING } from './groupManageTypes';

type ContactIconType = 'discord' | 'vk' | 'telegram' | 'twitch' | 'youtube' | 'whatsapp' | 'max';

const contactIcons = {
  discord: require('@/assets/images/groups/contacts/discord.png'),
  vk: require('@/assets/images/groups/contacts/vk.png'),
  telegram: require('@/assets/images/groups/contacts/telegram.png'),
  twitch: require('@/assets/images/groups/contacts/twitch.png'),
  youtube: require('@/assets/images/groups/contacts/youtube.png'),
  whatsapp: require('@/assets/images/groups/contacts/whatsapp.png'),
  max: require('@/assets/images/groups/contacts/max.png'),
};

export function useGroupData(groupId: string) {
  const [groupData, setGroupData] = useState<GroupDetailResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  const [groupName, setGroupName] = useState('');
  const [shortDescription, setShortDescription] = useState('');
  const [fullDescription, setFullDescription] = useState('');
  const [city, setCity] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [selectedContacts, setSelectedContacts] = useState<Contact[]>([]);
  const [groupImage, setGroupImage] = useState<string>('');
  const [originalImage, setOriginalImage] = useState<string>('');

  const loadGroupData = async () => {
    try {
      setIsLoading(true);
      const data = await groupService.getGroupDetail(groupId);

      data.image = normalizeImageUrl(data.image);
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

      const mappedContacts = (data.contacts || []).map(contact => {
        const type = getContactType(contact.name, contact.link) as ContactIconType;
        return {
            id: type,
            name: type,
            icon: contactIcons[type],
            description: contact.name,
            link: contact.link
        } as Contact;
      });

      console.log('[useGroupData] 📋 Контакты после загрузки:', mappedContacts);
      setSelectedContacts(mappedContacts);
      
    } catch (error: any) {
      console.error('[useGroupData] Ошибка загрузки данных группы:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось загрузить данные группы');
    } finally {
      setIsLoading(false);
    }
  };

  const saveGroupChanges = async () => {
    try {
      setIsSaving(true);
      
      let imageUrl = groupImage;

      if (groupImage !== originalImage && !groupImage.startsWith('http')) {
        imageUrl = await groupService.uploadGroupPhoto(groupImage);
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
      
      await groupService.updateGroup(groupId, updateData);
      
      Alert.alert('Успешно', 'Изменения сохранены!');

      await loadGroupData();
      
    } catch (error: any) {
      console.error('[useGroupData] Ошибка сохранения изменений:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось сохранить изменения');
    } finally {
      setIsSaving(false);
    }
  };

  const deleteGroup = async () => {
    try {
      setIsSaving(true);
      await groupService.deleteGroup(groupId);
      Alert.alert('Успешно', 'Группа успешно удалена');
      return true;
    } catch (error: any) {
      console.error('[useGroupData] Ошибка удаления группы:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось удалить группу');
      return false;
    } finally {
      setIsSaving(false);
    }
  };
  
  return {
    groupData,
    isLoading,
    isSaving,
    groupName,
    setGroupName,
    shortDescription,
    setShortDescription,
    fullDescription,
    setFullDescription,
    city,
    setCity,
    isPrivate,
    setIsPrivate,
    selectedCategories,
    setSelectedCategories,
    selectedContacts,
    setSelectedContacts,
    groupImage,
    setGroupImage,
    loadGroupData,
    saveGroupChanges,
    deleteGroup
  };
}