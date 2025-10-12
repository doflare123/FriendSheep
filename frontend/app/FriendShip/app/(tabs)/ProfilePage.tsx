import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event } from '@/components/event/EventCard';
import PageHeader from '@/components/PageHeader';
import ProfileHeader from '@/components/profile/ProfileHeader';
import ProfileStats from '@/components/profile/ProfileStats';
import StatisticsChart, { StatisticsData, StatisticsType } from '@/components/profile/StatisticsChart';
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
  const [statisticsType, setStatisticsType] = useState<StatisticsType>('media');

  const profileData = {
    avatar: require('@/assets/images/profile/profile_avatar.jpg'),
    name: 'Та самая Игфи',
    username: '@lgfi_22',
    description: 'Всем привет! Я новый участник сего проекта 🫶',
    registrationDate: '21.11.2025',
    telegramLink: 'https://t.me/your_bot',
  };

  const profileStats = {
    media: 20,
    games: 20,
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
      id: '1',
      title: "Baldur's Gate 3",
      date: '12.10.2025',
      genres: ['РПГ', 'Фэнтези', 'Приключения'],
      currentParticipants: 52752,
      maxParticipants: 100000,
      duration: '3 ч 30 мин',
      imageUri: 'https://i.pinimg.com/736x/ee/61/45/ee61457f7cb7c9e1cd8662ad77b8e464.jpg',
      description: 'Эпическая РПГ приключение',
      typeEvent: 'Игровая сессия',
      typePlace: 'online',
      eventPlace: 'Discord',
      publisher: 'Larian Studios',
      publicationDate: '10.10.2025',
      ageRating: '18+',
      category: 'game',
    },
    {
      id: '2',
      title: "Baldur's Gate 3",
      date: '12.10.2025',
      genres: ['РПГ', 'Фэнтези'],
      currentParticipants: 52752,
      maxParticipants: 100000,
      duration: '2 ч 45 мин',
      imageUri: 'https://i.pinimg.com/736x/ee/61/45/ee61457f7cb7c9e1cd8662ad77b8e464.jpg',
      description: 'Продолжение приключений',
      typeEvent: 'Игровая сессия',
      typePlace: 'online',
      eventPlace: 'Discord',
      publisher: 'Larian Studios',
      publicationDate: '10.10.2025',
      ageRating: '18+',
      category: 'game',
    },
  ];

  const upcomingSessions: Event[] = [
    {
      id: '3',
      title: "Baldur's Gate 3",
      date: '20.10.2025',
      genres: ['РПГ', 'Фэнтези', 'Приключения'],
      currentParticipants: 45000,
      maxParticipants: 100000,
      duration: '4 ч 00 мин',
      imageUri: 'https://i.pinimg.com/736x/ee/61/45/ee61457f7cb7c9e1cd8662ad77b8e464.jpg',
      description: 'Новая глава приключений',
      typeEvent: 'Игровая сессия',
      typePlace: 'online',
      eventPlace: 'Discord',
      publisher: 'Larian Studios',
      publicationDate: '15.10.2025',
      ageRating: '18+',
      category: 'game',
    },
  ];

  const statisticsData: StatisticsData = {
    media: [
      { name: 'Боевики', percentage: 44, color: '#4A90E2', legendFontColor: Colors.black },
      { name: 'Приколы', percentage: 21, color: '#7B68EE', legendFontColor: Colors.black },
      { name: 'Страшилки', percentage: 19, color: '#50C878', legendFontColor: Colors.black },
      { name: 'Романтика', percentage: 11, color: '#FFB6C1', legendFontColor: Colors.black },
      { name: 'РПГ', percentage: 6, color: '#FFA500', legendFontColor: Colors.black },
    ],
    games: [
      { name: 'РПГ', percentage: 35, color: '#4A90E2', legendFontColor: Colors.black },
      { name: 'Экшен', percentage: 30, color: '#7B68EE', legendFontColor: Colors.black },
      { name: 'Стратегии', percentage: 20, color: '#50C878', legendFontColor: Colors.black },
      { name: 'Приключения', percentage: 10, color: '#FFB6C1', legendFontColor: Colors.black },
      { name: 'Инди', percentage: 5, color: '#FFA500', legendFontColor: Colors.black },
    ],
    boardgames: [
      { name: 'Стратегии', percentage: 40, color: '#4A90E2', legendFontColor: Colors.black },
      { name: 'Карточные', percentage: 25, color: '#7B68EE', legendFontColor: Colors.black },
      { name: 'Кооператив', percentage: 20, color: '#50C878', legendFontColor: Colors.black },
      { name: 'Семейные', percentage: 10, color: '#FFB6C1', legendFontColor: Colors.black },
      { name: 'Партийные', percentage: 5, color: '#FFA500', legendFontColor: Colors.black },
    ],
    other: [
      { name: 'Разное', percentage: 50, color: '#4A90E2', legendFontColor: Colors.black },
      { name: 'Прочее', percentage: 30, color: '#7B68EE', legendFontColor: Colors.black },
      { name: 'Другое', percentage: 20, color: '#50C878', legendFontColor: Colors.black },
    ],
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
        />

        <ProfileStats 
          media={profileStats.media}
          games={profileStats.games}
          hours={profileStats.hours}
          sessions={profileStats.sessions}
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
          currentType={statisticsType}
          onTypeChange={setStatisticsType}
          mostPopularDay="Воскресенье"
          sessionsCreated={4}
          sessionsSeries={4}
        />

        <View style={{ height: 20 }} />
      </ScrollView>
      <BottomBar />
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