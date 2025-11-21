import { GroupDetailResponse } from '@/api/services/groupService';
import { CreateSessionData, UpdateSessionData } from '@/api/services/session';
import sessionService from '@/api/services/session/sessionService';
import { useMemo, useState } from 'react';
import { Alert } from 'react-native';
import {
    categoryToSessionType,
    convertToRFC3339,
    formatSessionDate,
    normalizeImageUrl,
    sessionTypeToCategory
} from './groupManageHelpers';

export function useGroupEvents(groupId: string, groupData: GroupDetailResponse | null) {
  const [availableGenres, setAvailableGenres] = useState<string[]>([]);
  const [isCreatingEvent, setIsCreatingEvent] = useState(false);
  const [isUpdatingEvent, setIsUpdatingEvent] = useState(false);
  
  const [createEventModalVisible, setCreateEventModalVisible] = useState(false);
  const [editEventModalVisible, setEditEventModalVisible] = useState(false);
  const [selectedEventId, setSelectedEventId] = useState<string>('');

  const loadGenres = async () => {
    try {
      const genres = await sessionService.getGenres();
      setAvailableGenres(genres);
    } catch (error: any) {
      console.error('[useGroupEvents] Ошибка загрузки жанров:', error);
    }
  };

  const handleCreateEvent = () => {
    setCreateEventModalVisible(true);
  };

  const handleEditEvent = (eventId: string) => {
    setSelectedEventId(eventId);
    setEditEventModalVisible(true);
  };

  const handleCreateEventSave = async (eventData: any, onSuccess: () => void) => {
    try {
      setIsCreatingEvent(true);

      const sessionData: CreateSessionData = {
        title: eventData.title,
        session_type: categoryToSessionType[eventData.category] || 'Игры',
        session_place: eventData.typePlace === 'online' ? 1 : 2,
        group_id: parseInt(groupId),
        start_time: convertToRFC3339(eventData.date),
        duration: parseInt(eventData.duration),
        count_users: eventData.maxParticipants,
        genres: eventData.genres?.join(',') || '',
        location: eventData.eventPlace || '',
        year: eventData.year,
        country: eventData.country || '',
        age_limit: eventData.ageRating || '',
        notes: eventData.description || '',
        image: eventData.image,
      };

      await sessionService.createSession(sessionData);
      
      setCreateEventModalVisible(false);
      Alert.alert('Успешно', 'Событие создано!');
      onSuccess();
    } catch (error: any) {
      Alert.alert('Ошибка', error.message || 'Не удалось создать событие');
    } finally {
      setIsCreatingEvent(false);
    }
  };

  const handleEditEventSave = async (eventId: string, eventData: any, onSuccess: () => void) => {
    try {
        setIsUpdatingEvent(true);

        const updateData: UpdateSessionData = {
        title: eventData.title,
        duration: parseInt(eventData.duration),
        count_users_max: eventData.maxParticipants,
        genres: eventData.genres,
        location: eventData.eventPlace || '',
        year: eventData.year,
        country: eventData.country || '',
        age_limit: eventData.ageRating || '',
        notes: eventData.description || '',
        };

        if (eventData.date) {
        updateData.start_time = convertToRFC3339(eventData.date);
        }

        if (eventData.image?.uri && !eventData.imageUri?.startsWith('http')) {
        const imageUrl: string = await sessionService.uploadSessionImage(eventData.image.uri);
        updateData.image_url = imageUrl;
        }

        await sessionService.updateSession(parseInt(eventId), updateData);
        
        setEditEventModalVisible(false);
        setSelectedEventId('');
        Alert.alert('Успешно', 'Событие обновлено!');
        onSuccess();
    } catch (error: any) {
        Alert.alert('Ошибка', error.message || 'Не удалось обновить событие');
    } finally {
        setIsUpdatingEvent(false);
    }
    };

  const formattedEvents = useMemo(() => {
    if (!groupData?.sessions) return [];
    
    return groupData.sessions.map(session => ({
        id: session.id.toString(),
        title: session.title,
        date: formatSessionDate(session.start_time),
        genres: session.genres || [],
        currentParticipants: session.current_users,
        maxParticipants: session.count_users_max,
        duration: `${session.duration} мин`,
        imageUri: normalizeImageUrl(session.image_url),
        description: '',
        typeEvent: session.session_type,
        typePlace: (session.session_place === 'offline' || session.session_place === 'online' 
            ? session.session_place : 'online') as 'online' | 'offline',
        eventPlace: session.city || '',
        publisher: groupData.name,
        publicationDate: session.start_time,
        ageRating: '',
        category: sessionTypeToCategory[session.session_type] || 'other',
        group: groupData.name,
        }));
    }, [groupData]);

  return {
    availableGenres,
    isCreatingEvent,
    isUpdatingEvent,
    createEventModalVisible,
    setCreateEventModalVisible,
    editEventModalVisible,
    setEditEventModalVisible,
    selectedEventId,
    formattedEvents,
    loadGenres,
    handleCreateEvent,
    handleEditEvent,
    handleCreateEventSave,
    handleEditEventSave,
  };
}