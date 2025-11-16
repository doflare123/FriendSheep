import groupService from '@/api/services/groupService';
import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { type Group } from '@/components/groups/GroupCard';
import GroupCarousel from '@/components/groups/GroupCarousel';
import CreateGroupModal from '@/components/groups/modal/CreateGroupModal';
import PageHeader from '@/components/PageHeader';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useCallback, useState } from 'react';
import { ActivityIndicator, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

type GroupsPageNavigationProp = StackNavigationProp<RootStackParamList, 'GroupsPage'>;

// Маппинг русских названий категорий на английские ключи
const CATEGORY_MAPPING: { [key: string]: string } = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};

const GroupsPage = () => {
  const { sortingState, sortingActions } = useSearchState();
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [managedGroups, setManagedGroups] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const navigation = useNavigation<GroupsPageNavigationProp>();

  // Моковые данные для подписок (пока не подключен API)
  const mockSubscriptions: Group[] = [];

  const loadManagedGroups = async () => {
    try {
      setIsLoading(true);
      const groups = await groupService.getAdminGroups();
      
      // Проверяем, что groups - это массив
      if (!Array.isArray(groups)) {
        console.warn('Получен некорректный формат данных групп:', groups);
        setManagedGroups([]);
        return;
      }

      console.log('=== Загруженные группы с бэка ===');
      console.log('Количество:', groups.length);
      if (groups.length > 0) {
        console.log('Пример первой группы:', JSON.stringify(groups[0], null, 2));
      }
      
      // Преобразуем данные с бэка в формат компонента
      const transformedGroups: Group[] = groups.map(group => {
        // ✅ Исправляем URL изображения (заменяем localhost на реальный IP)
        let imageUri = group.image;
        if (imageUri && imageUri.includes('localhost')) {
          imageUri = imageUri.replace('http://localhost:8080', 'http://192.168.0.209:8080');
        }
        
        // ✅ Преобразуем русские названия категорий в английские ключи
        const mappedCategories = (group.category || [])
          .map(cat => CATEGORY_MAPPING[cat])
          .filter(cat => cat !== undefined); // Убираем неизвестные категории
        
        const transformed = {
          id: group.id.toString(),
          name: group.name,
          participantsCount: group.member_count,
          description: group.small_description,
          imageUri: imageUri,
          categories: mappedCategories as any[],
          onPress: () => navigation.navigate('GroupPage', { groupId: group.id.toString(), mode: 'manage' }),
        };
        
        console.log('✅ Трансформированная группа:', {
          id: transformed.id,
          name: transformed.name,
          imageUri: transformed.imageUri,
          originalCategories: group.category,
          mappedCategories: transformed.categories,
        });
        
        return transformed;
      });
      
      setManagedGroups(transformedGroups);
    } catch (error: any) {
      console.error('Ошибка загрузки групп:', error);
      setManagedGroups([]);
    } finally {
      setIsLoading(false);
    }
  };

  const onRefresh = async () => {
    setRefreshing(true);
    await loadManagedGroups();
    setRefreshing(false);
  };

  // Загружаем группы при фокусе на экране
  useFocusEffect(
    useCallback(() => {
      loadManagedGroups();
    }, [])
  );

  const handleAddGroup = () => {
    setCreateModalVisible(true);
  };

  const handleCreateGroup = async (groupData: any) => {
    console.log('Группа создана:', groupData);
    // Перезагружаем список групп после создания
    await loadManagedGroups();
  };

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: Colors.white }}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <PageHeader title="Ваши группы" showWave />
      
      {isLoading ? (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color={Colors.blue} />
          <Text style={styles.loadingText}>Загрузка групп...</Text>
        </View>
      ) : (
        <ScrollView
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={onRefresh}
              colors={[Colors.blue]}
              tintColor={Colors.blue}
            />
          }
        >
          <CategorySection
            title="Группы под управлением:"
            customActionButton={{
              icon: require('../../assets/images/groups/group_add.png'),
              onPress: handleAddGroup,
              tintColor: Colors.blue2,
            }}
          >
            {managedGroups.length > 0 ? (
              <GroupCarousel 
                groups={managedGroups} 
                actionText="Управлять"
                actionColor={[Colors.lightBlue, Colors.blue]}
              />
            ) : (
              <View style={styles.emptyContainer}>
                <Text style={styles.emptyText}>
                  У вас пока нет групп под управлением
                </Text>
                <Text style={styles.emptySubtext}>
                  Создайте свою первую группу!
                </Text>
              </View>
            )}
          </CategorySection>

          {mockSubscriptions.length > 0 && (
            <CategorySection title="Подписки:">
              <GroupCarousel 
                groups={mockSubscriptions} 
                actionText="Перейти"
                actionColor={[Colors.lightBlue, Colors.blue]}
              />
            </CategorySection>
          )}
        </ScrollView>
      )}
      
      <BottomBar />

      <CreateGroupModal
        visible={createModalVisible}
        onClose={() => setCreateModalVisible(false)}
        onCreate={handleCreateGroup}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  loadingText: {
    marginTop: 10,
    fontFamily: Montserrat.regular,
    fontSize: 16,
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
});

export default GroupsPage;