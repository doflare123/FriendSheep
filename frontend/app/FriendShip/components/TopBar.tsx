import notificationService from '@/api/services/notificationService';
import permissionsService from '@/api/services/permissionsService';
import userService from '@/api/services/userService';
import { GroupInvite, Notification } from '@/api/types/notification';
import barsStyle from '@/app/styles/barsStyle';
import ConfirmationModal from '@/components/ConfirmationModal';
import NotificationsModal from '@/components/notifications/NotificationsModal';
import MainSearchBar from '@/components/search/MainSearchBar';
import { Colors } from '@/constants/Colors';
import { SortingActions, SortingState } from '@/hooks/useSearchState';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react';
import {
  Image,
  Linking,
  StyleSheet,
  TouchableOpacity,
  View
} from 'react-native';

interface TopBarProps {
  sortingState: SortingState;
  sortingActions: SortingActions;
}

export interface TopBarHandle {
  openNotificationsModal: () => void;
}

const TopBar = forwardRef<TopBarHandle, TopBarProps>(({ sortingState, sortingActions }, ref) => {
  const colors = useThemedColors();
  const [notificationsVisible, setNotificationsVisible] = useState(false);
  const [notifModalPos, setNotifModalPos] = useState({ top: 0, left: 0 });
  const notifIconRef = useRef<View>(null);

  const [hasUnread, setHasUnread] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [invites, setInvites] = useState<GroupInvite[]>([]);
  const [serverError, setServerError] = useState(false);
  const [isLoadingNotifications, setIsLoadingNotifications] = useState(false);
  const [userProfile, setUserProfile] = useState<any>(null);
  const [notificationsPermissionGranted, setNotificationsPermissionGranted] = useState(false);

  const [showPermissionModal, setShowPermissionModal] = useState(false);
  const [showPermissionDeniedModal, setShowPermissionDeniedModal] = useState(false);

  const MIN_REFRESH_INTERVAL = 30000;
  const lastRefreshTime = useRef(0);

  useEffect(() => {
    const checkNotificationsPermission = async () => {
      const granted = await permissionsService.checkNotificationsPermission();
      setNotificationsPermissionGranted(granted);
      console.log('[TopBar] 🔔 Разрешение на уведомления:', granted ? 'Есть' : 'Нет');
    };
    
    checkNotificationsPermission();
  }, []);

  const checkUnreadNotifications = useCallback(async () => {
    if (!notificationsPermissionGranted) {
      console.log('[TopBar] ⚠️ Пропускаем проверку - нет разрешения на уведомления');
      return;
    }

    try {
      const unreadData = await notificationService.checkUnreadNotifications();
      const hasNotifications = unreadData.has_unread;
      setHasUnread(hasNotifications);
      setServerError(false);
    } catch (error: any) {
      console.warn('[TopBar] Не удалось проверить уведомления:', error.message);
      setServerError(true);
    }
  }, [notificationsPermissionGranted]);

  const loadNotifications = async () => {
    const now = Date.now();
    
    if (now - lastRefreshTime.current < MIN_REFRESH_INTERVAL) {
      console.log('[TopBar] Слишком частые обновления, пропускаем');
      return;
    }
    
    lastRefreshTime.current = now;

    try {
      setIsLoadingNotifications(true);
      
      console.log('[TopBar] 📥 Загружаем уведомления и приглашения...');

      const notificationsData = await notificationService.getNotifications();
      console.log('[TopBar] ✅ Уведомлений:', notificationsData.length);
      setNotifications(notificationsData);

      const invitesData = await notificationService.getGroupInvites();
      console.log('[TopBar] ✅ Приглашений:', invitesData.length);
      setInvites(invitesData);

      const unreadData = await notificationService.checkUnreadNotifications();
      setHasUnread(unreadData.has_unread);
      
      console.log('[TopBar] ✅ Загрузка завершена. Всего:', notificationsData.length + invitesData.length);
      
    } catch (error) {
      console.error('[TopBar] ❌ Ошибка загрузки уведомлений:', error);
      setNotifications([]);
      setInvites([]);
      setHasUnread(false);
    } finally {
      setIsLoadingNotifications(false);
    }
  };

  const handleUnreadStatusChange = useCallback((newHasUnread: boolean) => {
    console.log('[TopBar] 🔔 Обновление статуса непрочитанных:', newHasUnread);
    setHasUnread(newHasUnread);
  }, []);

  useEffect(() => {
    if (notificationsPermissionGranted) {
      checkUnreadNotifications();
      const interval = setInterval(checkUnreadNotifications, 30000);
      return () => clearInterval(interval);
    }
  }, [checkUnreadNotifications, notificationsPermissionGranted]);

  useEffect(() => {
    const loadUserProfile = async () => {
      try {
        const profile = await userService.getCurrentUserProfile();
        setUserProfile(profile);
      } catch (error) {
        console.error('[TopBar] Ошибка загрузки профиля:', error);
      }
    };
    
    loadUserProfile();
  }, []);

  const requestNotificationsPermission = async (): Promise<boolean> => {
    try {
      const granted = await permissionsService.requestNotificationsPermission();
      setNotificationsPermissionGranted(granted);
      return granted;
    } catch (error) {
      console.error('[TopBar] Ошибка запроса разрешения:', error);
      return false;
    }
  };

  const handlePermissionRequest = async () => {
    setShowPermissionModal(false);
    const granted = await requestNotificationsPermission();
    
    if (granted) {
      await openNotifications();
    } else {
      setShowPermissionDeniedModal(true);
    }
  };

  const handleOpenSettings = () => {
    setShowPermissionModal(false);
    setShowPermissionDeniedModal(false);
    Linking.openSettings();
  };

  const openNotifications = async () => {
    const hasPermission = await permissionsService.checkNotificationsPermission();
    
    if (!hasPermission) {
      console.log('[TopBar] ⚠️ Нет разрешения на уведомления');
      setShowPermissionModal(true);
      return;
    }

    notifIconRef.current?.measureInWindow((x, y, width, height) => {
      const modalWidth = 320;
      const centerX = x + width / 2;
      const left = centerX - modalWidth + 5;
      const top = y + height * 0.5;

      setNotifModalPos({ top, left });
      setNotificationsVisible(true);
    });

    await loadNotifications();
  };

  useImperativeHandle(ref, () => ({
    openNotificationsModal: async () => {
      console.log('[TopBar] 🔔 Внешний вызов открытия модалки уведомлений');
      await openNotifications();
    },
  }));

  const hasTelegramLink = userProfile?.telegram_link === true;

  return (
    <View style={[styles.container, { backgroundColor: colors.white }]}>
      <MainSearchBar sortingState={sortingState} sortingActions={sortingActions} />
      
      <TouchableOpacity
        ref={notifIconRef}
        onPress={openNotifications}
        activeOpacity={0.7}
        style={styles.notificationButton}
      >
        <Image
          style={[barsStyle.notifications, {tintColor: colors.darkGrey}]}
          source={require('@/assets/images/top_bar/notifications.png')}
        />
        {hasUnread && !serverError && notificationsPermissionGranted && (
          <View style={styles.unreadIndicator} />
        )}
      </TouchableOpacity>

      <NotificationsModal
        visible={notificationsVisible}
        onClose={() => setNotificationsVisible(false)}
        position={notifModalPos}
        notifications={notifications}
        invites={invites}
        isLoading={isLoadingNotifications}
        serverError={serverError}
        hasTelegramLink={hasTelegramLink}
        onReload={loadNotifications}
        onUpdateNotifications={setNotifications}
        onUpdateInvites={setInvites}
        onUnreadStatusChange={handleUnreadStatusChange}
      />

      <ConfirmationModal
        visible={showPermissionModal}
        title="Разрешение на уведомления"
        message="Для просмотра уведомлений необходимо разрешить доступ к уведомлениям. Хотите предоставить разрешение?"
        onConfirm={handlePermissionRequest}
        onCancel={() => setShowPermissionModal(false)}
      />

      <ConfirmationModal
        visible={showPermissionDeniedModal}
        title="Разрешение не предоставлено"
        message="Вы можете включить уведомления в настройках устройства"
        onConfirm={handleOpenSettings}
        onCancel={() => setShowPermissionDeniedModal(false)}
      />
    </View>
  );
});

TopBar.displayName = 'TopBar';

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    zIndex: 10,
  },
  notificationButton: {
    position: 'relative',
  },
  unreadIndicator: {
    position: 'absolute',
    top: 0,
    right: 0,
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: Colors.red,
  },
});

export default TopBar;