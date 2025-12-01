import groupService, { AdminGroup } from '@/api/services/groupService';
import GroupCard, { GroupCategory } from '@/components/groups/GroupCard';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React, { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';

interface GroupSelectorModalProps {
  visible: boolean;
  onClose: () => void;
  onSelectGroup: (group: AdminGroup) => void;
}

const CATEGORY_MAPPING: Record<string, GroupCategory> = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};

const GroupSelectorModal: React.FC<GroupSelectorModalProps> = ({
  visible,
  onClose,
  onSelectGroup,
}) => {
  const [groups, setGroups] = useState<AdminGroup[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (visible) {
      loadGroups();
    }
  }, [visible]);

  const loadGroups = async () => {
    try {
      setLoading(true);
      console.log('[GroupSelectorModal] 📥 Загрузка групп администратора...');
      
      const adminGroups = await groupService.getAdminGroups();
      
      console.log('[GroupSelectorModal] ✅ Загружено групп:', adminGroups.length);
      setGroups(adminGroups);
    } catch (error: any) {
      console.error('[GroupSelectorModal] ❌ Ошибка загрузки групп:', error);
      setGroups([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectGroup = (group: AdminGroup) => {
    console.log('[GroupSelectorModal] ✅ Выбрана группа:', group.name);
    onSelectGroup(group);
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
          onPress={() => handleSelectGroup(item)}
        />
      </View>
    );
  };

  return (
    <Modal
      visible={visible}
      animationType="fade"
      transparent
      onRequestClose={onClose}
    >
      <View style={styles.overlay}>
        <View style={styles.modalContainer}>
          <View style={styles.header}>
            <Text style={styles.title}>Выберите группу</Text>
            <TouchableOpacity onPress={onClose} style={styles.closeButton}>
              <Text style={styles.closeButtonText}>✕</Text>
            </TouchableOpacity>
          </View>

          {loading ? (
            <View style={styles.loadingContainer}>
              <ActivityIndicator size="large" color={Colors.lightBlue} />
              <Text style={styles.loadingText}>Загрузка групп...</Text>
            </View>
          ) : groups.length === 0 ? (
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>
                У вас пока нет групп, где вы администратор
              </Text>
              <Text style={styles.emptySubtext}>
                Создайте группу, чтобы добавлять события
              </Text>
            </View>
          ) : (
            <FlatList
              data={groups}
              renderItem={renderGroupCard}
              keyExtractor={(item) => item.id.toString()}
              contentContainerStyle={styles.listContent}
              showsVerticalScrollIndicator={false}
            />
          )}
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  modalContainer: {
    width: '95%',
    maxHeight: '80%',
    backgroundColor: Colors.white,
    borderRadius: 25,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.black,
  },
  closeButton: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  closeButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 24,
    color: Colors.black,
  },
  loadingContainer: {
    padding: 40,
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 12,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
  },
  emptyContainer: {
    padding: 40,
    alignItems: 'center',
  },
  emptyText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.black,
    textAlign: 'center',
    marginBottom: 8,
  },
  emptySubtext: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
    textAlign: 'center',
  },
  listContent: {
    padding: 16,
  },
  cardWrapper: {
    marginBottom: 16,
  },
});

export default GroupSelectorModal;