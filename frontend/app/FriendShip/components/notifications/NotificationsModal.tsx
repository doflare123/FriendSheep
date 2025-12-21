import groupMemberService from '@/api/services/group/groupMemberService';
import notificationService from '@/api/services/notificationService';
import pushNotificationService from '@/api/services/pushNotificationService';
import { GroupInvite, Notification } from '@/api/types/notification';
import { useToast } from '@/components/ToastContext';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import {
  ActivityIndicator,
  Image,
  Linking,
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';

interface NotificationsModalProps {
  visible: boolean;
  onClose: () => void;
  position: { top: number; left: number };
  notifications: Notification[];
  invites: GroupInvite[];
  isLoading: boolean;
  serverError: boolean;
  hasTelegramLink: boolean;
  onReload: () => void;
  onUpdateNotifications: (notifications: Notification[]) => void;
  onUpdateInvites: (invites: GroupInvite[]) => void;
  onUnreadStatusChange?: (hasUnread: boolean) => void;
}

const NotificationsModal: React.FC<NotificationsModalProps> = ({
  visible,
  onClose,
  position,
  notifications,
  invites,
  isLoading,
  serverError,
  hasTelegramLink,
  onReload,
  onUpdateNotifications,
  onUpdateInvites,
  onUnreadStatusChange,
}) => {
  const colors = useThemedColors();
  const { showToast } = useToast();

  const checkAndUpdateUnreadStatus = async (
    remainingNotifications: Notification[], 
    remainingInvites: GroupInvite[]
  ) => {
    const totalUnread = remainingNotifications.length + remainingInvites.length;
    const hasUnread = totalUnread > 0;

    await pushNotificationService.setBadgeCount(totalUnread);
    
    if (onUnreadStatusChange) {
      onUnreadStatusChange(hasUnread);
    }
  };

  const handleMarkAsViewed = async (notificationId: number) => {
    try {
      await notificationService.markAsViewed(notificationId);

      const updatedNotifications = notifications.filter((n) => n.id !== notificationId);
      onUpdateNotifications(updatedNotifications);

      checkAndUpdateUnreadStatus(updatedNotifications, invites);
      
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
      await groupMemberService.respondToInvite(inviteId.toString(), 'accepted');

      const updatedInvites = invites.filter((inv) => inv.id !== inviteId);
      onUpdateInvites(updatedInvites);

      checkAndUpdateUnreadStatus(notifications, updatedInvites);
      
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
      await groupMemberService.respondToInvite(inviteId.toString(), 'rejected');

      const updatedInvites = invites.filter((inv) => inv.id !== inviteId);
      onUpdateInvites(updatedInvites);

      checkAndUpdateUnreadStatus(notifications, updatedInvites);
      
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

  const handleTelegramBotPress = () => {
    Linking.openURL('https://t.me/FriendShipNotify_bot');
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
    <Modal
      transparent
      animationType="fade"
      visible={visible}
      onRequestClose={onClose}
    >
      <Pressable style={styles.modalOverlay} onPress={onClose}>
        <Pressable
          style={[styles.notificationsModal, { top: position.top, left: position.left, backgroundColor: colors.white }]}
          onPress={() => {}}
        >
          <Text style={[styles.title, {color: colors.black}]}>
            {serverError 
              ? 'Сервис уведомлений недоступен'
              : totalCount > 0
              ? `У вас ${totalCount} ${totalCount === 1 ? 'уведомление' : totalCount < 5 ? 'уведомления' : 'уведомлений'}:`
              : 'Нет новых уведомлений'}
          </Text>

          {isLoading ? (
            <View style={styles.loadingContainer}>
              <ActivityIndicator size="small" color={colors.lightBlue} />
            </View>
          ) : serverError ? (
            <View style={styles.errorContainer}>
              <Text style={[styles.errorText, {color: colors.grey}]}>
                Не удалось загрузить уведомления.{'\n'}
                Возможно, notify_service не запущен.
              </Text>
              <TouchableOpacity 
                style={[styles.retryButton, {backgroundColor: colors.lightBlue}]}
                onPress={onReload}
              >
                <Text style={styles.retryButtonText}>Попробовать снова</Text>
              </TouchableOpacity>
            </View>
          ) : (
            <>
              {!hasTelegramLink && (
                <View style={styles.telegramBanner}>
                  <Image
                    source={require('@/assets/images/profile/telegram.png')}
                    style={styles.telegramBannerIcon}
                  />
                  <View style={styles.telegramBannerContent}>
                    <Text style={[styles.telegramBannerText, {color: colors.black}]}>
                      Получайте уведомления в Telegram!
                    </Text>
                    <TouchableOpacity onPress={handleTelegramBotPress}>
                      <Text style={[styles.telegramBannerLink, {color: colors.blue2}]}>
                        Привязать бота
                      </Text>
                    </TouchableOpacity>
                  </View>
                </View>
              )}

              <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
                {invites.map((invite) => (
                  <View key={`invite-${invite.id}`} style={styles.notificationItem}>
                    <View style={[styles.iconWrapper, {backgroundColor: colors.lightBlue}]}>
                      <Image
                        style={[styles.notificationIcon, {tintColor: colors.white}]}
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
                    <Text style={[styles.notificationTime, {color: colors.lightGrey}]}>{formatTime(invite.createdAt)}</Text>
                  </View>
                ))}

                {notifications.map((notification) => (
                  <TouchableOpacity
                    key={`notification-${notification.id}`}
                    style={styles.notificationItem}
                    onPress={() => handleMarkAsViewed(notification.id)}
                    activeOpacity={0.7}
                  >
                    <View style={[styles.iconWrapper, {backgroundColor: colors.lightBlue}]}>
                      <Image
                        style={[styles.notificationIcon, {tintColor: colors.white}]}
                        source={require('@/assets/images/top_bar/notifications.png')}
                      />
                    </View>
                    <Text style={[styles.notificationText, {color: colors.black}]}>{notification.text}</Text>
                    <Text style={[styles.notificationTime, {color: colors.lightGrey}]}>{formatTime(notification.sendAt)}</Text>
                  </TouchableOpacity>
                ))}

                {totalCount === 0 && !serverError && (
                  <Text style={[styles.emptyText, {color: colors.lightGrey}]}>Здесь пока ничего нет</Text>
                )}
                <Text style={[styles.hintText, {color: colors.lightGrey}]}>В данной версии приложения уведомления доступны только внутри приложения или через Telegram-бота</Text>
              </ScrollView>
            </>
          )}
        </Pressable>
      </Pressable>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: Colors.lightBlack,
  },
  notificationsModal: {
    position: 'absolute',
    width: 320,
    maxHeight: 400,
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
  telegramBanner: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  telegramBannerIcon: {
    width: 30,
    height: 30,
    marginRight: 16
  },
  telegramBannerContent: {
    flex: 1,
  },
  telegramBannerText: {
    fontFamily: Montserrat.regular,
    fontSize: 11,
    marginBottom: 2,
  },
  telegramBannerLink: {
    fontFamily: Montserrat.bold,
    fontSize: 11,
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
    textAlign: 'center',
    marginBottom: 16,
  },
  retryButton: {
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
    marginVertical: 16,
  },
  iconWrapper: {
    width: 48,
    height: 48,
    borderRadius: 24,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  notificationIcon: {
    width: 32,
    height: 32,
  },
  notificationText: {
    fontFamily: Montserrat.regular,
    flex: 1,
    fontSize: 11,
  },
  notificationTime: {
    fontFamily: Montserrat.regular,
    fontSize: 10,
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
    textAlign: 'center',
    paddingVertical: 24,
  },
  hintText: {
    fontFamily: Montserrat.regular,
    fontSize: 8,
    paddingVertical: 4,
  },
});

export default NotificationsModal;