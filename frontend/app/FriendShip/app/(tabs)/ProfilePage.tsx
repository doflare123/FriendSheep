import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event } from '@/components/event/EventCard';
import PageHeader from '@/components/PageHeader';
import EditProfileModal from '@/components/profile/EditProfileModal';
import ProfileHeader from '@/components/profile/ProfileHeader';
import ProfileStats from '@/components/profile/ProfileStats';
import StatisticsBottomBars from '@/components/profile/StatisticsBottomBars';
import StatisticsChart, { StatisticsDataItem } from '@/components/profile/StatisticsChart';
import TileSelectionModal, { TileType } from '@/components/profile/TileSelectionModal';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import React, { useState } from 'react';
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

const ProfilePage = () => {
  const { sortingState, sortingActions } = useSearchState();
  const [sessionFilter, setSessionFilter] = useState<'completed' | 'upcoming'>('upcoming');
  const [tileModalVisible, setTileModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [selectedTiles, setSelectedTiles] = useState<TileType[]>(['media', 'games', 'hours', 'sessions']);

  const [profileData, setProfileData] = useState({
    avatar: require('@/assets/images/profile/profile_avatar.jpg'),
    name: 'Та самая Игфи',
    username: '@lgfi_22',
    description: 'Всем привет! Я новый участник сего проекта 🫶',
    registrationDate: '21.11.2025',
    telegramLink: 'https://t.me/your_bot',
  });

  const allStats = {
    media: 20,
    games: 20,
    table_games: 20,
    other: 20,
    hours: 20,
    sessions: 20,
  };

  const favoriteGenres = [
    { name: 'Боевики', count: 21 },
    { name: 'Приколы', count: 18 },
    { name: 'Страшилки', count: 18 },
    { name: 'Романтика', count: 18 },
    { name: 'РПГ', count: 18 },
  ];

  const subscriptions = [
    { id: 1, image: require('@/assets/images/profile/profile_avatar.jpg') },
    { id: 2, image: require('@/assets/images/profile/profile_avatar.jpg') },
    { id: 3, image: require('@/assets/images/profile/profile_avatar.jpg') },
    { id: 4, image: require('@/assets/images/profile/profile_avatar.jpg') },
  ];

  const completedSessions: Event[] = [
    {
      id: '2',
      title: 'Матрица',
      date: '15.03.2004',
      imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
      description: "мяу",
      genres: ['Фантастика'],
      group: 'Мега крутая группа',
      currentParticipants: 48,
      maxParticipants: 50,
      duration: '136 минут',
      typeEvent: 'Фильм',
      typePlace: 'offline',
      eventPlace: 'Кинотеатр «Октябрь»',
      publisher: 'Warner Bros',
      publicationDate: '1999',
      ageRating: '16+',
      category: 'movie',
    },
    {
      id: '3',
      title: 'The Elder Scrolls V: Skyrim',
      date: '10.01.2004',
      imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
      description: "Киберспортивный турнир",
      genres: ['Шутер'],
      group: 'Мега крутая группа',
      currentParticipants: 32,
      maxParticipants: 64,
      duration: '240 минут',
      typeEvent: 'Турнир',
      typePlace: 'online',
      eventPlace: 'Steam',
      publisher: 'Valve',
      publicationDate: '2012',
      ageRating: '16+',
      category: 'game',
    },
  ];

  const upcomingSessions: Event[] = [
    {
      id: '1',
      title: 'Крестный отец',
      date: '12.02.2004',
      imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
      description: "ЭЩКЕРЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕ",
      genres: ['Драма', 'Криминал'],
      group: 'Мега крутая группа',
      currentParticipants: 52,
      maxParticipants: 52,
      duration: '175 минут',
      typeEvent: 'Фильм',
      typePlace: 'online',
      eventPlace: 'https://cinema.com',
      publisher: 'Paramount Pictures',
      publicationDate: '1972',
      ageRating: '18+',
      category: 'movie',
    },
  ];

  const statisticsData: StatisticsDataItem[] = [
    { name: 'Боевики', percentage: 25, color: '#4A90E2', legendFontColor: Colors.black },
    { name: 'РПГ', percentage: 22, color: '#7B68EE', legendFontColor: Colors.black },
    { name: 'Приколы', percentage: 18, color: '#50C878', legendFontColor: Colors.black },
    { name: 'Страшилки', percentage: 15, color: '#FFB6C1', legendFontColor: Colors.black },
    { name: 'Романтика', percentage: 12, color: '#FFA500', legendFontColor: Colors.black },
    { name: 'Стратегии', percentage: 8, color: '#FF6B6B', legendFontColor: Colors.black },
  ];

  const handleEditProfile = () => {
    setEditModalVisible(true);
  };

  const handleChangeTiles = () => {
    setTileModalVisible(true);
  };

  const handleProfileSave = (updatedProfile: any) => {
    setProfileData(prev => ({
      ...prev,
      ...updatedProfile,
    }));
  };

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <PageHeader title="Ваш профиль" showWave />
        
        <ProfileHeader 
          avatar={profileData.avatar}
          name={profileData.name}
          username={profileData.username}
          description={profileData.description}
          registrationDate={profileData.registrationDate}
          telegramLink={profileData.telegramLink}
          onEditProfile={handleEditProfile}
          onChangeTiles={handleChangeTiles}
        />

        <ProfileStats 
          selectedTiles={selectedTiles}
          stats={allStats}
        />

        <CategorySection title="Любимые жанры:" marginBottom={16}>
          <View style={styles.genresContainer}>
            <View style={styles.genresColumn}>
              {favoriteGenres.slice(0, 3).map((genre, index) => (
                <Text key={index} style={styles.genreItem}>
                  {index + 1}. {genre.name} - {genre.count}
                </Text>
              ))}
            </View>
            <View style={styles.genresColumn}>
              {favoriteGenres.slice(3).map((genre, index) => (
                <Text key={index} style={styles.genreItem}>
                  {index + 4}. {genre.name} - {genre.count}
                </Text>
              ))}
            </View>
          </View>
        </CategorySection>

        <CategorySection title="Подписки:" marginBottom={16}>
          <View style={styles.subscriptionsContainer}>
            {subscriptions.map((sub) => (
              <Image key={sub.id} source={sub.image} style={styles.subscriptionImage} />
            ))}
          </View>
        </CategorySection>

        <CategorySection 
          title="Сессии:" 
          customActionButton={{
            icon: sessionFilter === 'completed' 
              ? require('@/assets/images/profile/finished.png')
              : require('@/assets/images/profile/current.png'),
            onPress: () => setSessionFilter(sessionFilter === 'completed' ? 'upcoming' : 'completed'),
          }}
          events={sessionFilter === 'completed' ? completedSessions : upcomingSessions}
        />

        <StatisticsChart 
          statisticsData={statisticsData}
        />

        <StatisticsBottomBars
          mostPopularDay="Воскресенье"
          sessionsCreated={4}
          sessionsSeries={4}
        />

        <View style={{ height: 20 }} />
      </ScrollView>
      <BottomBar />

      <TileSelectionModal
        visible={tileModalVisible}
        onClose={() => setTileModalVisible(false)}
        selectedTiles={selectedTiles}
        onTilesChange={setSelectedTiles}
      />

      <EditProfileModal
        visible={editModalVisible}
        onClose={() => setEditModalVisible(false)}
        currentProfile={profileData}
        onSave={handleProfileSave}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  scrollView: {
    flex: 1,
  },
  genresContainer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    justifyContent: 'space-between',
    marginBottom: 8
  },
  genresColumn: {
    flex: 1,
  },
  genreItem: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    marginBottom: 2,
  },
  subscriptionsContainer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    gap: 12,
    justifyContent: 'space-between',
    marginBottom: 16
  },
  subscriptionImage: {
    width: 75,
    height: 75,
    borderRadius: 40,
  },
});

export default ProfilePage;