import React, { useMemo } from 'react';
import { FlatList, StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import UserCard, { User } from '@/components/profile/UserCard';
import SearchResultsSection from '@/components/search/SearchResultsSection';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { useSearchState } from '@/hooks/useSearchState';
import { createHighlightedText } from '@/utils/textHighlight';

const mockUsers: User[] = [
  {
    id: '1',
    name: 'Лейс с крабом',
    username: '@laysKRAB',
    description: 'Вкуснее, чем Pringles 😎',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
  {
    id: '2',
    name: 'Игорь Петров',
    username: '@igorpetrov',
    description: 'Люблю путешествовать и фотографировать',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
  {
    id: '3',
    name: 'Анна Смирнова',
    username: '@annasmirn',
    description: 'Дизайнер и любитель кофе ☕',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
  {
    id: '4',
    name: 'Максим Волков',
    username: '@maxvolkov',
    description: 'Программист и геймер 🎮',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
  {
    id: '5',
    name: 'Елена Козлова',
    username: '@elenakozlova',
    description: 'Художница и мечтательница 🎨',
    imageUri: 'https://i.pinimg.com/736x/9e/b9/76/9eb976bc8832404d75c575763a37bfe0.jpg',
  },
];

const highlightUserText = (user: User, query: string): User => {
  const highlightedName = createHighlightedText(user.name, query);
  const highlightedUsername = createHighlightedText(user.username, query);
  
  return {
    ...user,
    ...(highlightedName && { highlightedName }),
    ...(highlightedUsername && { highlightedUsername }),
  };
};

const UserSearchPage: React.FC = () => {
  const { sortingState, sortingActions } = useSearchState();

  const filteredUsers = useMemo((): User[] => {
    let users = [...mockUsers];
    
    if (sortingState.searchQuery.trim()) {
      users = users.filter(user =>
        user.name.toLowerCase().includes(sortingState.searchQuery.toLowerCase()) ||
        user.username.toLowerCase().includes(sortingState.searchQuery.toLowerCase())
      );
      
      users = users.map(user => highlightUserText(user, sortingState.searchQuery));
    } else {
      return [];
    }
    
    return users;
  }, [sortingState.searchQuery]);

  const handleUserPress = (userId: string) => {
    console.log('User pressed:', userId);
  };

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />

      <View style={styles.content}>
        <SearchResultsSection
          title="Поиск по профилям"
          searchQuery={sortingState.searchQuery}
          hasResults={filteredUsers.length > 0}
          showWave
        >
          {filteredUsers.length > 0 && (
            <View style={styles.contentContainer}>
              <FlatList
                data={filteredUsers}
                keyExtractor={(item) => item.id}
                renderItem={({ item }) => (
                  <View style={styles.cardWrapper}>
                    <UserCard
                      {...item}
                      onPress={() => handleUserPress(item.id)}
                    />
                  </View>
                )}
                showsVerticalScrollIndicator={false}
                contentContainerStyle={styles.listContent}
              />
            </View>
          )}
        </SearchResultsSection>
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
  content: {
    flex: 1,
  },
  contentContainer: {
    flex: 1,
    paddingHorizontal: 16,
  },
  listContent: {
    paddingBottom: 16,
  },
  cardWrapper: {
    marginBottom: 16,
  },
});

export default UserSearchPage;