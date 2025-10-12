import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event } from '@/components/event/EventCard';
import PageHeader from '@/components/PageHeader';
import StatisticsBar from '@/components/profile/StatisticsBar';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import React, { useState } from 'react';
import {
  Dimensions,
  Image,
  Linking,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';
import { PieChart } from 'react-native-chart-kit';
import { SafeAreaView } from 'react-native-safe-area-context';

const ProfilePage = () => {
  const { sortingState, sortingActions } = useSearchState();
  const [sessionFilter, setSessionFilter] = useState<'completed' | 'upcoming'>('upcoming');
  const [statisticsType, setStatisticsType] = useState<'media' | 'games' | 'boardgames' | 'other'>('media');

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

  const statisticsData = {
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

  const handleTelegramPress = () => {
    Linking.openURL('https://t.me/your_bot');
  };

  const getStatisticsForType = (type: typeof statisticsType) => {
    return statisticsData[type] || statisticsData.media;
  };

  const getStatisticsTitle = (type: typeof statisticsType) => {
    const titles = {
      media: 'Медиа',
      games: 'Игры',
      boardgames: 'Настольные игры',
      other: 'Другое',
    };
    return titles[type];
  };

  const chartConfig = {
    color: (opacity = 1) => `rgba(0, 0, 0, ${opacity})`,
  };

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <PageHeader title="Ваш профиль" showWave />
        
        <View style={styles.header}>
          <Image source={require('@/assets/images/profile/profile_avatar.jpg')} style={styles.profileImage}/>
          <View style={styles.headerInfo}>
            <View style={styles.nameRow}>
              <Text style={styles.profileName}>Та самая Игфи</Text>
              <TouchableOpacity onPress={handleTelegramPress}>
                <Image 
                  source={require('@/assets/images/profile/telegram.png')} 
                  style={styles.telegramIcon}
                />
              </TouchableOpacity>
            </View>
            <Text style={styles.profileUs}>@lgfi_22</Text>
            <Text style={styles.description}>Всем привет! Я новый участник сего проекта 🫶</Text>
            <Text style={styles.registrationDate}>21.11.2025</Text>
          </View>
        </View>

        <View style={styles.statisticsRow}>
          <StatisticsBar title='Медиа' count={20} icon="movies" />
          <StatisticsBar title='Игры' count={20} icon="games" />
        </View>
        
        <View style={[styles.statisticsRow, {marginBottom: 16}]}>
          <StatisticsBar title='Часов' count={20} icon="hours" />
          <StatisticsBar title='Сессий' count={20} icon="sessions" />
        </View>

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

        <CategorySection title="Статистика:" marginBottom={16}>
          <CategorySection 
            title={getStatisticsTitle(statisticsType)}
            centerTitle
            showLineVariant="line2"
            marginBottom={8}
            customNavigationButtons={{
              leftButton: {
                icon: require('@/assets/images/arrowLeft.png'),
                onPress: () => {
                  const types: typeof statisticsType[] = ['media', 'games', 'boardgames', 'other'];
                  const currentIndex = types.indexOf(statisticsType);
                  const prevIndex = currentIndex > 0 ? currentIndex - 1 : types.length - 1;
                  setStatisticsType(types[prevIndex]);
                },
              },
              rightButton: {
                icon: require('@/assets/images/arrow.png'),
                onPress: () => {
                  const types: typeof statisticsType[] = ['media', 'games', 'boardgames', 'other'];
                  const currentIndex = types.indexOf(statisticsType);
                  const nextIndex = currentIndex < types.length - 1 ? currentIndex + 1 : 0;
                  setStatisticsType(types[nextIndex]);
                },
              },
            }}
          />

          <View style={styles.statisticsContainer}>
              <View style={styles.chartContainer}>
                <PieChart
                  data={getStatisticsForType(statisticsType).map(item => ({
                    name: item.name,
                    population: item.percentage,
                    color: item.color,
                    legendFontColor: item.legendFontColor,
                    legendFontSize: 12,
                    legendFontFamily: Montserrat.regular,
                  }))}
                  width={Dimensions.get('window').width - 60}
                  height={200}
                  chartConfig={chartConfig}
                  accessor="population"
                  backgroundColor="transparent"
                  paddingLeft="16"
                  absolute={false}
                  hasLegend={true}
                />
              </View>
            
            <View style={styles.statisticsBottomBar}>
              <StatisticsBar title='Популярный день' count='Воскресенье' icon="most-popular-day" />
              <StatisticsBar title='Создано сессий' count={4} icon="sessions-created" />
              <StatisticsBar title='Серия сессий' count={4} icon="series-sessions" />
            </View>
          </View>
        </CategorySection>

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
  header: {
    marginTop: -10,
    flexDirection: 'row',
    padding: 20,
    alignItems: 'flex-start',
  },
  headerInfo: {
    flex: 1,
  },
  profileImage: {
    width: 120,
    height: 120,
    borderRadius: 100,
    marginRight: 16,
  },
  nameRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  profileName: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.black,
    marginRight: 4,
  },
  telegramIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
  },
  profileUs: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    marginTop: -6,
    marginBottom: 8,
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    marginBottom: 8,
  },
  registrationDate: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.lightGrey,
    alignSelf: 'flex-end',
  },
  statisticsRow: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    marginBottom: 8,
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
    fontSize: 14,
    color: Colors.black,
    marginBottom: 4,
  },
  subscriptionsContainer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    gap: 12,
    justifyContent: 'space-between',
    marginBottom: 16
  },
  subscriptionImage: {
    width: 70,
    height: 70,
    borderRadius: 35,
  },
  statisticsContainer: {
    paddingHorizontal: 16,
  },
  chartContainer: {
    justifyContent: 'space-between'
  },
  statisticsLegend: {
    marginLeft: 16,
  },
  legendItem: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    marginBottom: 4,
  },
  statisticsBottomBar: {
    gap: 8,
  },
});

export default ProfilePage;