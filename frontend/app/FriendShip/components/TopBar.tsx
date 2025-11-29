import groupService from '@/api/services/groupService';
import notificationService from '@/api/services/notificationService';
import { GroupInvite, Notification } from '@/api/types/notification';
import barsStyle from '@/app/styles/barsStyle';
import MainSearchBar from '@/components/search/MainSearchBar';
import { useToast } from '@/components/ToastContext';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { SortingActions, SortingState } from '@/hooks/useSearchState';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  ActivityIndicator,
  Image,
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';

interface TopBarProps {
  sortingState: SortingState;
  sortingActions: SortingActions;
}

const TopBar: React.FC<TopBarProps> = ({ sortingState, sortingActions }) => {
  const { showToast } = useToast();
  const [notificationsVisible, setNotificationsVisible] = useState(false);
  const [notifModalPos, setNotifModalPos] = useState({ top: 0, left: 0 });
  const notifIconRef = useRef<View>(null);

  const [loading, setLoading] = useState(false);
  const [hasUnread, setHasUnread] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [invites, setInvites] = useState<GroupInvite[]>([]);
  const [serverError, setServerError] = useState(false);
  const [isLoadingNotifications, setIsLoadingNotifications] = useState(false);

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

  const handleMarkAsViewed = async (notificationId: number) => {
    try {
      await notificationService.markAsViewed(notificationId);

      setNotifications((prev) => prev.filter((n) => n.id !== notificationId));
      
      showToast({
        type: 'success',
        title: 'Готово',
        message: 'Уведомление отмечено как просмотренное',
      });
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось отметить уведомление',
      });
    }
  };

  const handleAcceptInvite = async (inviteId: number, groupId: number) => {
    try {
      await groupService.respondToInvite(inviteId.toString(), 'accepted');

      setInvites((prev) => prev.filter((inv) => inv.id !== inviteId));
      
      showToast({
        type: 'success',
        title: 'Успешно',
        message: 'Вы присоединились к группе',
      });
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось принять приглашение',
      });
    }
  };

  const handleRejectInvite = async (inviteId: number) => {
    try {
      await groupService.respondToInvite(inviteId.toString(), 'rejected');

      setInvites((prev) => prev.filter((inv) => inv.id !== inviteId));
      
      showToast({
        type: 'success',
        title: 'Готово',
        message: 'Приглашение отклонено',
      });
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось отклонить приглашение',
      });
    }
  };

  const formatTime = (dateString: string): string => {
    try {
      const date = new Date(dateString);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffMins = Math.floor(diffMs / 60000);
      const diffHours = Math.floor(diffMins / 60);
      const diffDays = Math.floor(diffHours / 24);

      if (diffMins < 1) return 'только что';
      if (diffMins < 60) return `${diffMins} мин назад`;
      if (diffHours < 24) return `${diffHours} ч назад`;
      if (diffDays < 7) return `${diffDays} дн назад`;
      
      return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
    } catch {
      return '';
    }
  };

  const totalCount = notifications.length + invites.length;

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

      <Modal
        transparent
        animationType="fade"
        visible={notificationsVisible}
        onRequestClose={() => setNotificationsVisible(false)}
      >
        <Pressable
          style={styles.modalOverlay}
          onPress={() => setNotificationsVisible(false)}
        >
          <Pressable
            style={[styles.notificationsModal, { top: notifModalPos.top, left: notifModalPos.left }]}
            onPress={() => {}}
          >
            <Text style={styles.title}>
              {serverError 
                ? 'Сервис уведомлений недоступен'
                : totalCount > 0
                ? `У вас ${totalCount} ${totalCount === 1 ? 'уведомление' : totalCount < 5 ? 'уведомления' : 'уведомлений'}:`
                : 'Нет новых уведомлений'}
            </Text>

            {loading ? (
              <View style={styles.loadingContainer}>
                <ActivityIndicator size="small" color={Colors.lightBlue} />
              </View>
            ) : serverError ? (
              <View style={styles.errorContainer}>
                <Text style={styles.errorText}>
                  Не удалось загрузить уведомления.{'\n'}
                  Возможно, notify_service не запущен.
                </Text>
                <TouchableOpacity 
                  style={styles.retryButton}
                  onPress={loadNotifications}
                >
                  <Text style={styles.retryButtonText}>Попробовать снова</Text>
                </TouchableOpacity>
              </View>
            ) : (
              <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
                {invites.map((invite) => (
                  <View key={`invite-${invite.id}`} style={styles.notificationItem}>
                    <View style={styles.iconWrapper}>
                      <Image
                        style={styles.notificationIcon}
                        source={require('@/assets/images/top_bar/notifications.png')}
                      />
                    </View>
                    <View style={{ flex: 1 }}>
                      <Text style={styles.notificationText}>
                        Вас приглашают в группу {invite.groupName}
                      </Text>
                      <View style={styles.buttonsRow}>
                        <TouchableOpacity
                          style={styles.acceptButton}
                          onPress={() => handleAcceptInvite(invite.id, invite.groupId)}
                        >
                          <Text style={styles.buttonText}>Принять</Text>
                        </TouchableOpacity>
                        <TouchableOpacity
                          style={styles.rejectButton}
                          onPress={() => handleRejectInvite(invite.id)}
                        >
                          <Text style={styles.buttonText}>Отклонить</Text>
                        </TouchableOpacity>
                      </View>
                    </View>
                    <Text style={styles.notificationTime}>{formatTime(invite.createdAt)}</Text>
                  </View>
                ))}

                {notifications.map((notification) => (
                  <TouchableOpacity
                    key={`notification-${notification.id}`}
                    style={styles.notificationItem}
                    onPress={() => handleMarkAsViewed(notification.id)}
                    activeOpacity={0.7}
                  >
                    <View style={styles.iconWrapper}>
                      <Image
                        style={styles.notificationIcon}
                        source={require('@/assets/images/top_bar/notifications.png')}
                      />
                    </View>
                    <Text style={styles.notificationText}>{notification.text}</Text>
                    <Text style={styles.notificationTime}>{formatTime(notification.sendAt)}</Text>
                  </TouchableOpacity>
                ))}

                {totalCount === 0 && !serverError && (
                  <Text style={styles.emptyText}>Здесь пока ничего нет</Text>
                )}
              </ScrollView>
            )}
          </Pressable>
        </Pressable>
      </Modal>
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
  modalOverlay: {
    flex: 1,
    backgroundColor: Colors.lightBlack,
  },
  notificationsModal: {
    position: 'absolute',
    width: 320,
    maxHeight: 400,
    backgroundColor: Colors.white,
    borderRadius: 12,
    padding: 16,
    borderTopEndRadius: 0,
    shadowColor: Colors.black,
    shadowRadius: 10,
    elevation: 10,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 12,
    fontWeight: '600',
    marginBottom: 12,
  },
  scrollView: {
    maxHeight: 320,
  },
  loadingContainer: {
    paddingVertical: 24,
    alignItems: 'center',
  },
  errorContainer: {
    paddingVertical: 24,
    alignItems: 'center',
  },
  errorText: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.grey,
    textAlign: 'center',
    marginBottom: 16,
  },
  retryButton: {
    backgroundColor: Colors.lightBlue,
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 12,
  },
  retryButtonText: {
    fontFamily: Montserrat.regular,
    color: Colors.white,
    fontSize: 12,
  },
  notificationItem: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    marginBottom: 16,
  },
  iconWrapper: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: Colors.lightBlue,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  notificationIcon: {
    width: 32,
    height: 32,
    tintColor: Colors.white,
  },
  notificationText: {
    fontFamily: Montserrat.regular,
    flex: 1,
    fontSize: 11,
  },
  notificationTime: {
    fontFamily: Montserrat.regular,
    fontSize: 10,
    color: Colors.lightGrey,
    marginLeft: 8,
    alignSelf: 'flex-end',
  },
  buttonsRow: {
    flexDirection: 'row',
    marginTop: 8,
  },
  acceptButton: {
    backgroundColor: Colors.lightBlue,
    paddingHorizontal: 12,
    paddingVertical: 2,
    borderRadius: 12,
    marginRight: 8,
  },
  rejectButton: {
    backgroundColor: Colors.red,
    paddingVertical: 2,
    paddingHorizontal: 8,
    borderRadius: 12,
  },
  buttonText: {
    fontFamily: Montserrat.regular,
    color: Colors.white,
    fontSize: 10,
  },
  emptyText: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.lightGrey,
    textAlign: 'center',
    paddingVertical: 24,
  },
});

export default TopBar;