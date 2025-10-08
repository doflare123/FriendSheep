import BottomBar from '@/components/BottomBar';
import CreateGroupModal from '@/components/CreateGroupModal';
import { Group } from '@/components/GroupCard';
import GroupCarousel from '@/components/GroupCarousel';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useState } from 'react';
import { Image, ScrollView, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
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
        <View style={styles.headerContainer}>
          <Text style={styles.headerTitle}>Ваши группы</Text>
        </View>
         <Image
            source={require('../../assets/images/wave.png')}
            style={{resizeMode: 'none', marginBottom: -20}}
          />
        <ScrollView>
          <View style={styles.resultsHeader}>
            <View style={{flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'}}>
              <Text style={styles.resultsText}>Группы под управлением:</Text>
              <TouchableOpacity onPress={handleAddGroup}>
                <Image
                  source={require('../../assets/images/groups/group_add.png')}
                  style={{resizeMode: 'contain', width: 30, height: 30}}
                  tintColor={Colors.blue2}
                />
              </TouchableOpacity>
            </View>
            <Image
              source={require('../../assets/images/line.png')}
              style={{resizeMode: 'none'}}
            />
          </View>
          <GroupCarousel 
            groups={mockManagedGroups} 
            actionText="Управлять"
            actionColor={[Colors.lightBlue, Colors.blue]}
          />

          <View style={styles.resultsHeader}>
            <Text style={styles.resultsText}>Подписки:</Text>
            <Image
              source={require('../../assets/images/line.png')}
              style={{resizeMode: 'none'}}
            />
          </View>

          <GroupCarousel 
            groups={mockSubscriptions} 
            actionText="Перейти"
            actionColor={[Colors.lightBlue, Colors.blue]}
          />
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
  headerContainer: {
    backgroundColor: Colors.lightBlue2,
    paddingVertical: 20,
    alignItems: 'center',
    marginBottom: 12,
  },
  headerTitle: {
    fontFamily: Montserrat_Alternates.medium,
    fontSize: 24,
    color: Colors.white,
    textTransform: 'none',
  },
  resultsHeader: {
    paddingHorizontal: 16,
  },
  resultsText: {
    fontFamily: Montserrat_Alternates.medium,
    fontSize: 20,
    color: Colors.blue2,
    marginBottom: 4
  },
});

export default GroupsPage;