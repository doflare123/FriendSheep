import BottomBar from '@/components/BottomBar';
import CreateGroupModal from '@/components/CreateGroupModal';
import { Group } from '@/components/GroupCard';
import GroupCarousel from '@/components/GroupCarousel';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useState } from 'react';
import { Image, ImageBackground, ScrollView, StyleSheet, Text, TouchableOpacity } from 'react-native';
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
    onPress: () => navigation.navigate('GroupPage', { groupId: '2', mode: 'manage' }),
  },
];

const mockSubscriptions: Group[] = [
  {
    id: '3',
    name: 'Мега крутая группа',
    participantsCount: 47,
    description: 'Мы крутые пацантре, ёёёёу 😎\nПрисоединяйтесь к нам',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
    onPress: () => navigation.navigate('GroupPage', { groupId: '3', mode: 'view' }),
  },
  {
    id: '4',
    name: 'Мега',
    participantsCount: 25,
    description: 'Мы крутые пацантре, ёёёёу 😎\nПрисоединяйтесь к н',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
    onPress: () => navigation.navigate('GroupPage', { groupId: '4', mode: 'view' }),
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
      <ImageBackground
        source={require('../../assets/images/wallpaper.png')}
        style={{ flex: 1 }}
        resizeMode="cover"
      >
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />

        <ScrollView contentContainerStyle={{ paddingVertical: 30 }}>
          <ImageBackground style={{backgroundColor: Colors.white, paddingVertical: 10}}>
            <ImageBackground
              source={require('../../assets/images/title_rectangle.png')}
              style={styles.sectionHeader}
              resizeMode="cover"
            >
              <Text style={styles.sectionTitle}>Группы под управлением:</Text>
              <TouchableOpacity style={styles.managementIcon} onPress={handleAddGroup}>
                <Image
                  source={require('../../assets/images/groups/group_add.png')}
                  style={styles.managementIconImage}
                  resizeMode="contain"
                />
              </TouchableOpacity>
            </ImageBackground>

            <GroupCarousel 
              groups={mockManagedGroups} 
              actionText="Управлять"
              actionColor={[Colors.lightBlue, Colors.blue]}
            />

            <ImageBackground
              source={require('../../assets/images/title_rectangle.png')}
              style={styles.sectionHeader}
              resizeMode="cover"
            >
              <Text style={styles.sectionTitle}>Подписки:</Text>
            </ImageBackground>

            <GroupCarousel 
              groups={mockSubscriptions} 
              actionText="Перейти"
              actionColor={[Colors.lightBlue, Colors.blue]}
            />
          </ImageBackground>
        </ScrollView>
      </ImageBackground>

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
  contentBackground: {
    backgroundColor: Colors.white,
    marginHorizontal: 8,
    marginTop: 16,
    marginBottom: 16,
    borderRadius: 12,
    paddingBottom: 16,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    resizeMode: 'contain',
    marginHorizontal: 0,
    marginTop: 0,
    paddingHorizontal: 16,
    paddingVertical: 12,
    overflow: 'hidden',
  },
  sectionTitle: {
    color: Colors.black,
    fontSize: 18,
    fontFamily: inter.bold,
  },
  managementIcon: {
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 6
  },
  managementIconImage: {
    width: 30,
    height: 30,
  },
});

export default GroupsPage;