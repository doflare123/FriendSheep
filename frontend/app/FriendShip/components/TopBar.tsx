import notificationService from '@/api/services/notificationService';
import userService from '@/api/services/userService';
import { GroupInvite, Notification } from '@/api/types/notification';
import barsStyle from '@/app/styles/barsStyle';
import NotificationsModal from '@/components/notifications/NotificationsModal';
import MainSearchBar from '@/components/search/MainSearchBar';
import { Colors } from '@/constants/Colors';
import { SortingActions, SortingState } from '@/hooks/useSearchState';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  Image,
  StyleSheet,
  TouchableOpacity,
  View
} from 'react-native';

interface TopBarProps {
  sortingState: SortingState;
  sortingActions: SortingActions;
}

const TopBar: React.FC<TopBarProps> = ({ sortingState, sortingActions }) => {
  const [notificationsVisible, setNotificationsVisible] = useState(false);
  const [notifModalPos, setNotifModalPos] = useState({ top: 0, left: 0 });
  const notifIconRef = useRef<View>(null);

  const [hasUnread, setHasUnread] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [invites, setInvites] = useState<GroupInvite[]>([]);
  const [serverError, setServerError] = useState(false);
  const [isLoadingNotifications, setIsLoadingNotifications] = useState(false);
  const [userProfile, setUserProfile] = useState<any>(null);

  const MIN_REFRESH_INTERVAL = 30000;
  let lastRefreshTime = 0;

  const checkUnreadNotifications = useCallback(async () => {
    try {
      const unreadData = await notificationService.checkUnreadNotifications();
      const hasNotifications = unreadData.has_unread;
      setHasUnread(hasNotifications);
      setServerError(false);
    } catch (error: any) {
      console.warn('[TopBar] Не удалось проверить уведомления:', error.message);
      setServerError(true);
    }
  }, []);

  const loadNotifications = async () => {
    const now = Date.now();
    
    if (now - lastRefreshTime < MIN_REFRESH_INTERVAL) {
      console.log('[TopBar] Слишком частые обновления, пропускаем');
      return;
    }
    
    lastRefreshTime = now;

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

  useEffect(() => {
    checkUnreadNotifications();
    const interval = setInterval(checkUnreadNotifications, 30000);
    return () => clearInterval(interval);
  }, [checkUnreadNotifications]);

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

  const openNotifications = async () => {
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

  const hasTelegramLink = userProfile?.telegram_link && 
                          userProfile.telegram_link.trim() !== '' &&
                          userProfile.telegram_link !== 'https://t.me/FriendShipNotify_bot';

  return (
    <View style={styles.container}>
      <MainSearchBar sortingState={sortingState} sortingActions={sortingActions} />
      
      <TouchableOpacity
        ref={notifIconRef}
        onPress={openNotifications}
        activeOpacity={0.7}
        style={styles.notificationButton}
      >
        <Image
          style={barsStyle.notifications}
          source={require('@/assets/images/top_bar/notifications.png')}
        />
        {hasUnread && !serverError && <View style={styles.unreadIndicator} />}
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
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.white,
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