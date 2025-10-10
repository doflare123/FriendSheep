import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Group } from '@/components/groups/GroupCard';
import GroupCarousel from '@/components/groups/GroupCarousel';
import CreateGroupModal from '@/components/groups/modal/CreateGroupModal';
import PageHeader from '@/components/PageHeader';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useState } from 'react';
import { ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

type GroupsPageNavigationProp = StackNavigationProp<RootStackParamList, 'GroupsPage'>;

const GroupsPage = () => {
  const { sortingState, sortingActions } = useSearchState();
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const navigation = useNavigation<GroupsPageNavigationProp>();

  const mockManagedGroups: Group[] = [
    {
      id: '1',
      name: 'Мега крутая группа',
      participantsCount: 47,
      description: 'Мы крутые пацантре, ёёёёу 😎\nПрисоединяйтесь к нам',
      imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      onPress: () => navigation.navigate('GroupPage', { groupId: '1', mode: 'manage' }),
    },
    {
      id: '2',
      name: 'Мега',
      participantsCount: 32,
      description: 'Мы крутые пацантре, ёёёёу 😎\nПрисоединяйтесь к н',
      imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      onPress: () => navigation.navigate('GroupPage', { groupId: '1', mode: 'manage' }),
    },
  ];

  const mockSubscriptions: Group[] = [
    {
      id: '3',
      name: 'Мега крутая группа',
      participantsCount: 47,
      description: 'Мы крутые пацантре, ёёёёу 😎\nПрисоединяйтесь к нам',
      imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      onPress: () => navigation.navigate('GroupPage', { groupId: '1', mode: 'view' }),
    },
    {
      id: '4',
      name: 'Мега',
      participantsCount: 25,
      description: 'Мы крутые пацантре, ёёёёу 😎\nПрисоединяйтесь к н',
      imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
      onPress: () => navigation.navigate('GroupPage', { groupId: '1', mode: 'view' }),
    },
  ];

  const handleAddGroup = () => {
    setCreateModalVisible(true);
  };

  const handleCreateGroup = (groupData: any) => {
    console.log('Creating group:', groupData);
  };

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: Colors.white }}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <PageHeader title="Ваши группы" showWave />
      <ScrollView>
        <CategorySection
          title="Группы под управлением:"
          customActionButton={{
            icon: require('../../assets/images/groups/group_add.png'),
            onPress: handleAddGroup,
            tintColor: Colors.blue2,
          }}
        >
          <GroupCarousel 
            groups={mockManagedGroups} 
            actionText="Управлять"
            actionColor={[Colors.lightBlue, Colors.blue]}
          />
        </CategorySection>

        <CategorySection title="Подписки:">
          <GroupCarousel 
            groups={mockSubscriptions} 
            actionText="Перейти"
            actionColor={[Colors.lightBlue, Colors.blue]}
          />
        </CategorySection>
      </ScrollView>
      <BottomBar />

      <CreateGroupModal
        visible={createModalVisible}
        onClose={() => setCreateModalVisible(false)}
        onCreate={handleCreateGroup}
      />
    </SafeAreaView>
  );
};

export default GroupsPage;