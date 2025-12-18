import groupMemberService from '@/api/services/group/groupMemberService';
import groupService from '@/api/services/group/groupService';
import { AdminGroup } from '@/api/services/group/groupTypes';
import GroupCard, { GroupCategory } from '@/components/groups/GroupCard';
import { useToast } from '@/components/ToastContext';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Dimensions,
  FlatList,
  Image,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View
} from 'react-native';

const screenHeight = Dimensions.get("window").height;

interface InviteToGroupModalProps {
  visible: boolean;
  onClose: () => void;
  userId: number;
  onInviteSent?: () => void;
}

const CATEGORY_MAPPING: Record<string, GroupCategory> = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};

const InviteToGroupModal: React.FC<InviteToGroupModalProps> = ({
  visible,
  onClose,
  userId,
  onInviteSent,
}) => {
  const colors = useThemedColors();
  const { showToast } = useToast();
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [adminGroups, setAdminGroups] = useState<AdminGroup[]>([]);

  useEffect(() => {
    if (visible) {
      loadAdminGroups();
    }
  }, [visible]);

  const loadAdminGroups = async () => {
    try {
      setLoading(true);
      console.log('[InviteToGroupModal] 📥 Загрузка групп администратора...');
      
      const groups = await groupService.getAdminGroups();
      
      console.log('[InviteToGroupModal] ✅ Загружено групп:', groups.length);
      setAdminGroups(groups);
      
    } catch (error: any) {
      console.error('[InviteToGroupModal] ❌ Ошибка загрузки групп:', error);
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось загрузить группы',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleGroupPress = async (groupId: number) => {
    try {
      setSending(true);
      console.log('[InviteToGroupModal] 📨 Отправка приглашения:', { groupId, userId });
      
      const result = await groupMemberService.sendInviteToUser(groupId, userId);
      
      console.log('[InviteToGroupModal] ✅ Результат:', result);
      
      showToast({
        type: 'success',
        title: 'Успешно',
        message: result.message || 'Приглашение отправлено',
      });

      onInviteSent?.();
      onClose();
      
    } catch (error: any) {
      console.error('[InviteToGroupModal] ❌ Ошибка отправки:', error);
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось отправить приглашение',
      });
    } finally {
      setSending(false);
    }
  };

  const renderGroupCard = ({ item }: { item: AdminGroup }) => {
    const mappedCategories = item.category
      .map(cat => CATEGORY_MAPPING[cat])
      .filter((cat): cat is GroupCategory => cat !== undefined);

    return (
      <View style={styles.cardWrapper}>
        <GroupCard
          id={item.id.toString()}
          name={item.name}
          participantsCount={item.member_count}
          description={item.small_description}
          imageUri={item.image}
          categories={mappedCategories}
          onPress={() => handleGroupPress(item.id)}
        />
      </View>
    );
  };

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <TouchableWithoutFeedback onPress={sending ? undefined : onClose}>
        <View style={styles.overlay}>
          <TouchableWithoutFeedback>
            <View style={[styles.modal, { backgroundColor: colors.white }]}>
              <View style={[
                styles.header,
                {
                  backgroundColor: colors.white,
                  borderBottomColor: colors.veryLightGrey
                }
              ]}>
                <Text style={[styles.title, { color: colors.black }]}>
                  Выберите группу
                </Text>
                <TouchableOpacity 
                  style={styles.closeButton} 
                  onPress={onClose}
                  disabled={sending}
                >
                  <Image
                    tintColor={colors.black}
                    style={{ width: 35, height: 35, resizeMode: 'cover' }}
                    source={require('@/assets/images/event_card/back.png')}
                  />
                </TouchableOpacity>
              </View>

              <View style={[styles.content, { backgroundColor: colors.white }]}>
                {loading ? (
                  <View style={styles.loadingContainer}>
                    <ActivityIndicator size="large" color={colors.lightBlue} />
                    <Text style={[styles.loadingText, { color: colors.grey }]}>
                      Загрузка групп...
                    </Text>
                  </View>
                ) : sending ? (
                  <View style={styles.loadingContainer}>
                    <ActivityIndicator size="large" color={colors.lightBlue} />
                    <Text style={[styles.loadingText, { color: colors.grey }]}>
                      Отправка приглашения...
                    </Text>
                  </View>
                ) : adminGroups.length > 0 ? (
                  <FlatList
                    data={adminGroups}
                    keyExtractor={(item) => item.id.toString()}
                    renderItem={renderGroupCard}
                    contentContainerStyle={styles.listContent}
                    showsVerticalScrollIndicator={false}
                  />
                ) : (
                  <View style={styles.emptyContainer}>
                    <Text style={[styles.emptyText, { color: colors.grey }]}>
                      У вас пока нет групп, где вы администратор
                    </Text>
                  </View>
                )}
              </View>
            </View>
          </TouchableWithoutFeedback>
        </View>
      </TouchableWithoutFeedback>
    </Modal>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  modal: {
    marginHorizontal: 6,
    borderRadius: 25,
    overflow: "hidden",
    maxHeight: screenHeight * 0.8,
    width: '95%',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 14,
    paddingHorizontal: 16,
    position: 'relative',
    borderBottomWidth: 1,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    textAlign: 'center',
  },
  closeButton: {
    position: 'absolute',
    right: 10,
    zIndex: 1,
  },
  content: {
    paddingVertical: 8,
    maxHeight: screenHeight * 0.6,
  },
  loadingContainer: {
    paddingVertical: 40,
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 12,
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  listContent: {
    padding: 16,
  },
  cardWrapper: {
    marginBottom: 16,
  },
  emptyContainer: {
    paddingVertical: 40,
    alignItems: 'center',
  },
  emptyText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    textAlign: 'center',
    paddingHorizontal: 20,
  },
});

export default InviteToGroupModal;