import * as Calendar from 'expo-calendar';
import { Platform } from 'react-native';

export async function requestCalendarPermissions(): Promise<boolean> {
  try {
    const { status } = await Calendar.requestCalendarPermissionsAsync();
    return status === 'granted';
  } catch (error) {
    console.error('[CalendarHelper] Ошибка запроса разрешений:', error);
    return false;
  }
}

async function getDefaultCalendarSource() {
  if (Platform.OS === 'ios') {
    const defaultCalendar = await Calendar.getDefaultCalendarAsync();
    return defaultCalendar.source;
  } else {
    const calendars = await Calendar.getCalendarsAsync(Calendar.EntityTypes.EVENT);
    const defaultCalendar = calendars.find(cal => cal.allowsModifications);
    if (!defaultCalendar?.source) {
      throw new Error('Не найден календарь с source');
    }
    return defaultCalendar.source;
  }
}

async function getOrCreateFriendShipCalendar(): Promise<string> {
  try {
    const calendars = await Calendar.getCalendarsAsync(Calendar.EntityTypes.EVENT);

    let friendShipCalendar = calendars.find(
      cal => cal.title === 'FriendShip' && cal.allowsModifications
    );

    if (friendShipCalendar) {
      console.log('[CalendarHelper] ✅ Календарь FriendShip найден:', friendShipCalendar.id);
      return friendShipCalendar.id;
    }

    console.log('[CalendarHelper] 📅 Создаём новый календарь FriendShip');
    
    const defaultCalendar = await getDefaultCalendarSource();
    
    const newCalendarId = await Calendar.createCalendarAsync({
      title: 'FriendShip',
      color: '#4A90E2',
      entityType: Calendar.EntityTypes.EVENT,
      sourceId: defaultCalendar.id,
      source: defaultCalendar,
      name: 'FriendShip Events',
      ownerAccount: 'personal',
      accessLevel: Calendar.CalendarAccessLevel.OWNER,
    });

    console.log('[CalendarHelper] ✅ Календарь создан:', newCalendarId);
    return newCalendarId;
  } catch (error) {
    console.error('[CalendarHelper] ❌ Ошибка получения/создания календаря:', error);

    const calendars = await Calendar.getCalendarsAsync(Calendar.EntityTypes.EVENT);
    const writableCalendar = calendars.find(cal => cal.allowsModifications);
    
    if (!writableCalendar) {
      throw new Error('Не найден календарь для записи');
    }
    
    console.log('[CalendarHelper] ⚠️ Используем fallback календарь:', writableCalendar.title);
    return writableCalendar.id;
  }
}

export interface AddEventToCalendarParams {
  title: string;
  location?: string;
  startDate: Date;
  endDate: Date;
  notes?: string;
  groupName?: string;
}

export async function addEventToCalendar(
  params: AddEventToCalendarParams
): Promise<string> {
  try {
    const hasPermission = await requestCalendarPermissions();
    if (!hasPermission) {
      throw new Error('Нет доступа к календарю');
    }

    console.log('[CalendarHelper] 📅 Добавление события в календарь:', params.title);
    console.log('[CalendarHelper] 📅 Дата начала:', params.startDate.toISOString());
    console.log('[CalendarHelper] 📅 Дата окончания:', params.endDate.toISOString());

    const calendarId = await getOrCreateFriendShipCalendar();

    const notes = params.groupName 
      ? `Событие из FriendShip\nОрганизатор: ${params.groupName}\n\n${params.notes || ''}`
      : `Событие из FriendShip\n\n${params.notes || ''}`;

    const eventId = await Calendar.createEventAsync(calendarId, {
      title: params.title,
      location: params.location,
      startDate: params.startDate,
      endDate: params.endDate,
      notes: notes.trim(),
    });

    console.log('[CalendarHelper] ✅ Событие добавлено в календарь:', eventId);

    try {
      const createdEvent = await Calendar.getEventAsync(eventId);
      console.log('[CalendarHelper] ✅ Событие подтверждено:', createdEvent.title);
    } catch (verifyError) {
      console.warn('[CalendarHelper] ⚠️ Не удалось проверить событие, но скорее всего оно создано');
    }

    return eventId;
  } catch (error: any) {
    console.error('[CalendarHelper] ❌ Ошибка добавления события в календарь:', error);
    throw error;
  }
}

export async function removeEventFromCalendar(eventId: string): Promise<void> {
  try {
    const hasPermission = await requestCalendarPermissions();
    if (!hasPermission) {
      throw new Error('Нет доступа к календарю');
    }

    console.log('[CalendarHelper] 🗑️ Удаление события из календаря:', eventId);
    
    await Calendar.deleteEventAsync(eventId);
    
    console.log('[CalendarHelper] ✅ Событие удалено из календаря');
  } catch (error: any) {
    console.error('[CalendarHelper] ❌ Ошибка удаления события из календаря:', error);
    console.log('[CalendarHelper] ℹ️ Возможно событие уже было удалено вручную');
  }
}