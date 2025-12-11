import { StackNavigationProp } from '@react-navigation/stack';
import React, { useState } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import VerticalGroupList from '@/components/groups/VerticalGroupList';
import GroupSearchBar from '@/components/search/GroupSearchBar';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useAllGroups } from '@/hooks/groups/useAllGroups';
import { useGroupSearchState } from '@/hooks/useGroupSearchState';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';

type GroupsMode = 'managed' | 'subscriptions';

interface AllGroupsPageProps {
  navigation: StackNavigationProp<RootStackParamList, 'AllGroupsPage'>;
}

const AllGroupsPage: React.FC<AllGroupsPageProps> = ({ navigation }) => {
  const [groupsMode, setGroupsMode] = useState<GroupsMode>('managed');
  
  const { sortingState, sortingActions } = useSearchState();
  const { searchState: groupSearchState, searchActions: groupSearchActions } = useGroupSearchState();

  const handleGroupPress = (groupId: string) => {
    navigation.navigate('GroupPage', { groupId });
  };

  const { 
    groups,
    isLoading,
    error,
    refreshGroups 
  } = useAllGroups(groupsMode, groupSearchState, handleGroupPress);

  const getGroupWordForm = (count: number): string => {
    if (count % 10 === 1 && count % 100 !== 11) {
      return 'группа';
    } else if ([2, 3, 4].includes(count % 10) && ![12, 13, 14].includes(count % 100)) {
      return 'группы';
    } else {
      return 'групп';
    }
  };

  const handleToggleMode = () => {
    setGroupsMode(prev => prev === 'managed' ? 'subscriptions' : 'managed');
  };

  const getTitle = () => {
    return groupsMode === 'managed' ? 'Управляемые группы' : 'Мои подписки';
  };

  const getEmptyText = () => {
    if (groupSearchState.searchQuery.trim()) {
      return `Ничего не найдено по запросу "${groupSearchState.searchQuery}"`;
    }
    
    return groupsMode === 'managed' 
      ? 'У вас пока нет групп под управлением'
      : 'У вас пока нет подписок на группы';
  };

  if (isLoading) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        
        <CategorySection
          title={getTitle()}
          showBackButton
          onBackPress={() => navigation.goBack()}
        />
        
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color={Colors.lightBlue} />
          <Text style={styles.loadingText}>Загрузка групп...</Text>
        </View>
        
        <BottomBar />
      </SafeAreaView>
    );
  }

  if (error) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        
        <CategorySection
          title={getTitle()}
          showBackButton
          onBackPress={() => navigation.goBack()}
        />
        
        <View style={styles.centerContainer}>
          <Text style={styles.errorTitle}>Ошибка загрузки</Text>
          <Text style={styles.errorText}>{error}</Text>
        </View>
        
        <BottomBar />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      
      <CategorySection
        title={getTitle()}
        showBackButton
        onBackPress={() => navigation.goBack()}
        customActionButton={{
          icon: groupsMode === 'subscriptions' 
            ? require('@/assets/images/groups/subscriptions.png')
            : require('@/assets/images/groups/managed.png'),
          onPress: handleToggleMode,
        }}
      />

      <View style={styles.searchContainer}>
        <GroupSearchBar 
          searchState={groupSearchState} 
          searchActions={groupSearchActions}
        />
      </View>

      {groups.length > 0 && (
        <View style={styles.resultsHeader}>
          <Text style={styles.resultsCount}>
            Найдено: {groups.length} {getGroupWordForm(groups.length)}
          </Text>
        </View>
      )}

      <View style={styles.contentContainer}>
        {groups.length > 0 ? (
          <VerticalGroupList groups={groups} />
        ) : (
          <View style={styles.noGroupsContainer}>
            <Text style={styles.noGroupsText}>
              {getEmptyText()}
            </Text>
          </View>
        )}
      </View>

      <BottomBar />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  searchContainer: {
    paddingHorizontal: 16,
  },
  resultsHeader: {
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  resultsCount: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
  },
  contentContainer: {
    flex: 1,
    paddingHorizontal: 16,
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  loadingText: {
    marginTop: 12,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  errorTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.red,
    marginBottom: 8,
  },
  errorText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    textAlign: 'center',
    color: Colors.grey,
  },
  noGroupsContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 16,
  },
  noGroupsText: {
    fontFamily: Montserrat.regular,
    fontSize: 20,
    textAlign: 'center',
    color: Colors.blue2,
    marginBottom: 4,
  },
});

export default AllGroupsPage;