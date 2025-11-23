import { GroupDetailResponse } from '@/api/services/groupService';
import { CreateSessionData, UpdateSessionData } from '@/api/services/session';
import sessionService from '@/api/services/session/sessionService';
import { useMemo, useState } from 'react';
import { Alert } from 'react-native';
import {
  categoryToSessionType,
  convertToRFC3339,
  formatSessionDate,
  formatSessionDateTime,
  normalizeImageUrl,
  sessionTypeToCategory
} from './groupManageHelpers';

export function useGroupEvents(groupId: string, groupData: GroupDetailResponse | null) {
  const [availableGenres, setAvailableGenres] = useState<string[]>([]);
  const [isCreatingEvent, setIsCreatingEvent] = useState(false);
  const [isUpdatingEvent, setIsUpdatingEvent] = useState(false);
  const [isLoadingEventDetails, setIsLoadingEventDetails] = useState(false);
  
  const [createEventModalVisible, setCreateEventModalVisible] = useState(false);
  const [editEventModalVisible, setEditEventModalVisible] = useState(false);
  const [selectedEventId, setSelectedEventId] = useState<string>('');
  const [selectedEventData, setSelectedEventData] = useState<any>(null);

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

  const handleEditEvent = async (eventId: string) => {
    try {
      setIsLoadingEventDetails(true);
      setSelectedEventId(eventId);
      
      console.log('[useGroupEvents] 📋 Загрузка полных данных сессии для редактирования:', eventId);

      const fullSessionData = await sessionService.getSessionDetail(parseInt(eventId));
      
      console.log('[useGroupEvents] ✅ Полные данные получены:', fullSessionData);
      console.log('[useGroupEvents] 🔍 Metadata:', fullSessionData.metadata);

      const sessionFromList = groupData?.sessions?.find(s => s.id.toString() === eventId);
      
      if (!sessionFromList) {
        throw new Error('Сессия не найдена в списке');
      }

      const eventForEdit = {
        id: eventId,
        title: fullSessionData.session.title,
        date: formatSessionDateTime(fullSessionData.session.start_time),
        genres: fullSessionData.metadata?.Genres || [],
        description: fullSessionData.metadata?.Notes || '',
        eventPlace: fullSessionData.metadata?.Location || '',
        publisher: fullSessionData.metadata?.Country || '',
        publicationDate: fullSessionData.metadata?.Year?.toString() || '',
        ageRating: fullSessionData.metadata?.AgeLimit || '',   
        currentParticipants: fullSessionData.session.current_users,
        maxParticipants: fullSessionData.session.count_users_max,
        duration: `${fullSessionData.session.duration} мин`,
        imageUri: normalizeImageUrl(fullSessionData.session.image_url),
        typeEvent: fullSessionData.session.session_type,
        typePlace: (fullSessionData.session.session_place === 'offline' || 
                   fullSessionData.session.session_place === 'online' 
          ? fullSessionData.session.session_place : 'online') as 'online' | 'offline',
        category: sessionTypeToCategory[fullSessionData.session.session_type] || 'other',
        group: groupData?.name || '',
      };
      
      console.log('[useGroupEvents] 📦 Данные для редактирования сформированы:', eventForEdit);
      console.log('[useGroupEvents] 📝 Проверка полей:');
      console.log('  - description:', eventForEdit.description);
      console.log('  - publisher:', eventForEdit.publisher);
      console.log('  - ageRating:', eventForEdit.ageRating);
      console.log('  - eventPlace:', eventForEdit.eventPlace);
      console.log('  - genres:', eventForEdit.genres);
      console.log('  - publicationDate:', eventForEdit.publicationDate);
      
      setSelectedEventData(eventForEdit);
      setEditEventModalVisible(true);
      
    } catch (error: any) {
      console.error('[useGroupEvents] ❌ Ошибка загрузки данных для редактирования:', error);
      Alert.alert('Ошибка', 'Не удалось загрузить данные события');
    } finally {
      setIsLoadingEventDetails(false);
    }
  };

  const handleCreateEventSave = async (eventData: any, onSuccess: () => void) => {
    try {
      setIsCreatingEvent(true);

      console.log('[useGroupEvents] 📦 Данные из модалки:', eventData);

      const sessionData: CreateSessionData = {
        title: eventData.title,
        session_type: categoryToSessionType[eventData.category] || 'Игры',
        session_place: eventData.typePlace === 'online' ? 1 : 2,
        group_id: parseInt(groupId),
        start_time: convertToRFC3339(eventData.date),
        duration: parseInt(eventData.duration) || undefined,
        count_users: eventData.maxParticipants,
        genres: eventData.genres?.join(',') || '',
        location: eventData.eventPlace || '',
        year: eventData.year || undefined,
        country: eventData.publisher || '',
        age_limit: eventData.ageRating || '',
        notes: eventData.description || '',
        image: eventData.image,
      };

      console.log('[useGroupEvents] 📝 Сформированные данные для API:');
      console.log('  - notes (описание):', sessionData.notes);
      console.log('  - country (издатель):', sessionData.country);
      console.log('  - year:', sessionData.year);
      console.log('  - age_limit:', sessionData.age_limit);
      console.log('  - location:', sessionData.location);

      await sessionService.createSession(sessionData);
      
      setCreateEventModalVisible(false);
      Alert.alert('Успешно', 'Событие создано!');
      onSuccess();
    } catch (error: any) {
      console.error('[useGroupEvents] ❌ Ошибка создания:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось создать событие');
    } finally {
      setIsCreatingEvent(false);
    }
  };

  const handleEditEventSave = async (eventId: string, eventData: any, onSuccess: () => void) => {
    try {
      setIsUpdatingEvent(true);

      console.log('[useGroupEvents] 📝 Обновление события:', eventId);
      console.log('[useGroupEvents] 📦 Данные для обновления:', eventData);

      const updateData: UpdateSessionData = {
        title: eventData.title,
        duration: parseInt(eventData.duration) || undefined,
        count_users_max: eventData.maxParticipants,
        genres: eventData.genres,
        location: eventData.eventPlace || '',
        year: eventData.year || undefined,
        country: eventData.publisher || '',
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

      console.log('[useGroupEvents] 📝 Финальные данные для PATCH:');
      console.log('  - notes:', updateData.notes);
      console.log('  - country:', updateData.country);
      console.log('  - year:', updateData.year);
      console.log('  - age_limit:', updateData.age_limit);

      await sessionService.updateSession(parseInt(eventId), updateData);
      
      setEditEventModalVisible(false);
      setSelectedEventId('');
      setSelectedEventData(null);
      Alert.alert('Успешно', 'Событие обновлено!');
      onSuccess();
    } catch (error: any) {
      console.error('[useGroupEvents] ❌ Ошибка обновления:', error);
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
    isLoadingEventDetails,
    createEventModalVisible,
    setCreateEventModalVisible,
    editEventModalVisible,
    setEditEventModalVisible,
    selectedEventId,
    selectedEventData,
    formattedEvents,
    loadGenres,
    handleCreateEvent,
    handleEditEvent,
    handleCreateEventSave,
    handleEditEventSave,
  };
}