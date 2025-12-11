import userService from '@/api/services/userService';
import { Subscription, UpdateProfileRequest, UserProfile } from '@/api/types/user';
import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import PageHeader from '@/components/PageHeader';
import EditProfileModal from '@/components/profile/EditProfileModal';
import FavoriteGenres from '@/components/profile/FavoriteGenres';
import InviteToGroupModal from '@/components/profile/InviteToGroupModal';
import ProfileHeader from '@/components/profile/ProfileHeader';
import ProfileStats from '@/components/profile/ProfileStats';
import StatisticsBottomBars from '@/components/profile/StatisticsBottomBars';
import StatisticsChart from '@/components/profile/StatisticsChart';
import TileSelectionModal, { TileType } from '@/components/profile/TileSelectionModal';
import { useToast } from '@/components/ToastContext';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { sessionsToEvents } from '@/utils/dataAdapters';
import { formatDate } from '@/utils/dateUtils';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import React, { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Image,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

type ProfilePageRouteProp = RouteProp<RootStackParamList, 'ProfilePage'>;

const ProfilePage: React.FC = () => {
  const route = useRoute<ProfilePageRouteProp>();
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const { showToast } = useToast();
  const userId = route.params?.userId;
  
  const { sortingState, sortingActions } = useSearchState();
  const [sessionFilter, setSessionFilter] = useState<'completed' | 'upcoming'>('upcoming');
  const [tileModalVisible, setTileModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [inviteModalVisible, setInviteModalVisible] = useState(false);

  const [loading, setLoading] = useState(true);
  const [profileData, setProfileData] = useState<UserProfile | null>(null);
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
  const [selectedTiles, setSelectedTiles] = useState<TileType[]>([]);

  const [isOwnProfile, setIsOwnProfile] = useState(!userId);
  const [currentUserUsername, setCurrentUserUsername] = useState<string | null>(null);

  const loadProfileData = useCallback(async () => {
    try {
      setLoading(true);

      let currentProfile: UserProfile | null = null;
      
      if (!currentUserUsername) {
        currentProfile = await userService.getCurrentUserProfile();
        setCurrentUserUsername(currentProfile.us);
        console.log('[ProfilePage] 🔍 Текущий username:', currentProfile.us);
      }

      const myUsername = currentProfile?.us || currentUserUsername;
      const isViewingOwnProfile = !userId || userId === myUsername;
      setIsOwnProfile(isViewingOwnProfile);
      
      console.log('[ProfilePage] 🔍 Запрошенный userId:', userId);
      console.log('[ProfilePage] 🔍 Это свой профиль:', isViewingOwnProfile);

      const profile = isViewingOwnProfile
        ? (currentProfile || await userService.getCurrentUserProfile())
        : await userService.getUserProfileById(userId);

      console.log('[ProfilePage] Загруженный профиль:', profile.name, '(свой:', isViewingOwnProfile, ')');

      if (!profile || !profile.name || !profile.tiles) {
        throw new Error('Получены некорректные данные профиля');
      }
      
      setProfileData(profile);

      const tiles = (profile.tiles || []).map(tile => {
        const tileMap: Record<string, TileType> = {
          'count_all': 'all',
          'count_films': 'movies',
          'count_games': 'video_games',
          'count_table': 'board_games',
          'count_other': 'other',     
          'spent_time': 'time',
        };
        return tileMap[tile] || 'all';
      });
      setSelectedTiles(tiles);

      try {
        const subsUserId = isViewingOwnProfile 
          ? undefined 
          : (typeof profile.id === 'number' ? profile.id : parseInt(userId));
          
        const subs = await userService.getUserSubscriptions(subsUserId);
        setSubscriptions(subs);
      } catch (error) {
        console.warn('[ProfilePage] Не удалось загрузить подписки');
      }

    } catch (error: any) {
      console.error('[ProfilePage] ❌ Ошибка загрузки профиля:', error);
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось загрузить профиль',
      });

      navigation.goBack();
    } finally {
      setLoading(false);
    }
  }, [userId, currentUserUsername, showToast, navigation]);

  useEffect(() => {
    loadProfileData();
  }, []);

  const handleEditProfile = () => {
    setEditModalVisible(true);
  };

  const handleChangeTiles = () => {
    setTileModalVisible(true);
  };

  const handleInviteToGroup = () => {
    setInviteModalVisible(true);
  };

  const handleInviteSent = () => {
    showToast({
      type: 'success',
      title: 'Успешно',
      message: 'Приглашение отправлено пользователю',
    });
  };

  const handleProfileSave = async (updatedProfile: any) => {
    try {
      setLoading(true);

      let imageUrl = profileData?.image;

      if (updatedProfile.avatar?.uri && typeof updatedProfile.avatar === 'object') {
        console.log('[ProfilePage] 📤 Загружаем новое изображение...');
        imageUrl = await userService.uploadImage(updatedProfile.avatar.uri);
        console.log('[ProfilePage] ✅ Изображение загружено:', imageUrl);
      }

      const updateData: UpdateProfileRequest = {};

      if (updatedProfile.name?.trim()) {
        updateData.name = updatedProfile.name.trim();
      }

      if (updatedProfile.username?.trim()) {
        updateData.us = updatedProfile.username.trim();
      }

      if (updatedProfile.description?.trim()) {
        updateData.status = updatedProfile.description.trim();
      }

      if (imageUrl) {
        updateData.image = imageUrl;
      }

      console.log('[ProfilePage] 📤 Обновляем профиль с данными:', JSON.stringify(updateData, null, 2));

      await userService.updateProfile(updateData);

      showToast({
        type: 'success',
        title: 'Успешно',
        message: 'Профиль обновлён',
      });

      await loadProfileData();
      setEditModalVisible(false);

    } catch (error: any) {
      console.error('[ProfilePage] ❌ Ошибка обновления профиля:', error);
      console.error('[ProfilePage] ❌ Сообщение об ошибке:', error.message);
      
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось обновить профиль',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleTilesSave = async (newTiles: TileType[]) => {
    try {
      const settings = {
        count_all: newTiles.includes('all'),
        count_films: newTiles.includes('movies'),
        count_games: newTiles.includes('video_games'),
        count_table: newTiles.includes('board_games'),
        count_other: newTiles.includes('other'),
        spent_time: newTiles.includes('time'),
      };

      await userService.updateTileSettings(settings);
      setSelectedTiles(newTiles);

      showToast({
        type: 'success',
        title: 'Успешно',
        message: 'Плитки обновлены',
      });
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось обновить плитки',
      });
    }
  };

  const handleGroupPress = (groupId: number) => {
    navigation.navigate('GroupPage', { groupId: groupId.toString() });
  };

  if (loading) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color={Colors.lightBlue} />
          <Text style={styles.loadingText}>Загрузка профиля...</Text>
        </View>
        <BottomBar />
      </SafeAreaView>
    );
  }

  if (!profileData) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <View style={styles.loadingContainer}>
          <Text style={styles.errorText}>Не удалось загрузить профиль</Text>
        </View>
        <BottomBar />
      </SafeAreaView>
    );
  }

  const generateUserStatistics = () => {
    if (!profileData.popular_genres || profileData.popular_genres.length === 0) {
      return [];
    }
    
    const total = profileData.popular_genres.reduce((sum, genre) => sum + genre.count, 0);
    
    if (total === 0) return [];
    
    const colors = [
      '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#95E1D3',
      '#F38181', '#AA96DA', '#FCBAD3', '#A8D8EA', '#FFCB85'
    ];
    
    return profileData.popular_genres.map((genre, index) => ({
      name: genre.name,
      percentage: Math.round((genre.count / total) * 100),
      color: colors[index % colors.length],
      legendFontColor: '#000'
    }));
  };

  const generateCategoryStatistics = () => {
    if (!profileData) return [];
    
    const categoryData = [
      { 
        key: 'movies', 
        name: 'Медиа', 
        count: profileData.user_stats.count_films 
      },
      { 
        key: 'video_games', 
        name: 'Видеоигры', 
        count: profileData.user_stats.count_games 
      },
      { 
        key: 'board_games', 
        name: 'Настольные игры', 
        count: profileData.user_stats.count_table_games 
      },
      { 
        key: 'other', 
        name: 'Другое', 
        count: profileData.user_stats.count_another 
      },
    ].filter(cat => cat.count > 0);
    
    const total = categoryData.reduce((sum, cat) => sum + cat.count, 0);
    
    if (total === 0) return [];
    
    const colors: Record<string, string> = {
      'movies': '#FF6B6B',
      'video_games': '#4ECDC4',
      'board_games': '#45B7D1',
      'other': '#FFA07A',
    };
    
    return categoryData.map((cat) => ({
      name: cat.name,
      percentage: Math.round((cat.count / total) * 100),
      color: colors[cat.key] || '#95E1D3',
      legendFontColor: '#000'
    }));
  };

  const statisticsData = generateUserStatistics();
  const categoryStatisticsData = generateCategoryStatistics();

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <PageHeader 
          title={isOwnProfile ? "Ваш профиль" : profileData.name} 
          showWave 
        />
        
        <ProfileHeader 
          avatar={{ uri: profileData.image }}
          name={profileData.name}
          username={'@' + profileData.us}
          description={profileData.status}
          registrationDate={formatDate(profileData.data_register)}
          telegramLink={profileData.telegram_link ? `https://t.me/${profileData.us}` : undefined}
          isOwnProfile={isOwnProfile}
          onEditProfile={isOwnProfile ? handleEditProfile : undefined}
          onChangeTiles={isOwnProfile ? handleChangeTiles : undefined}
          onInviteToGroup={!isOwnProfile ? handleInviteToGroup : undefined}
        />

        <ProfileStats 
          selectedTiles={selectedTiles}
          stats={{
            all: profileData.user_stats.count_all,
            movies: profileData.user_stats.count_films,
            video_games: profileData.user_stats.count_games,
            board_games: profileData.user_stats.count_table_games,
            other: profileData.user_stats.count_another,
            time: Math.round(profileData.user_stats.spent_time / 60),
          }}
        />

        <CategorySection title="Любимые жанры:" marginBottom={16}>
          <FavoriteGenres genres={profileData.popular_genres || []} />
        </CategorySection>

        <CategorySection 
          title="Сессии:" 
          customActionButton={isOwnProfile ? {
            icon: sessionFilter === 'completed' 
              ? require('@/assets/images/profile/finished.png')
              : require('@/assets/images/profile/current.png'),
            onPress: () => setSessionFilter(
              sessionFilter === 'completed' ? 'upcoming' : 'completed'
            ),
          } : undefined}
          secondaryActionButton={{
            icon: require('@/assets/images/more.png'),
            onPress: () => navigation.navigate('AllEventsPage', { 
              mode: 'user',
              userId: isOwnProfile ? undefined : userId
            }),
          }}
          events={
            sessionFilter === 'completed' 
              ? sessionsToEvents(profileData.recent_sessions)
              : sessionsToEvents(profileData.upcoming_sessions)
          }
        >
          {((sessionFilter === 'completed' && (!profileData.recent_sessions || profileData.recent_sessions.length === 0)) ||
            (sessionFilter === 'upcoming' && (!profileData.upcoming_sessions || profileData.upcoming_sessions.length === 0))) && (
            <Text style={styles.emptyText}>
              {sessionFilter === 'completed' 
                ? 'Завершённых сессий пока нет' 
                : 'Предстоящих сессий пока нет'}
            </Text>
          )}
        </CategorySection>

        {subscriptions.length > 0 && (
          <CategorySection title="Подписки:" marginBottom={16}>
            <View style={styles.subscriptionsContainer}>
              {subscriptions.map((sub) => (
                <TouchableOpacity 
                  key={sub.id}
                  onPress={() => handleGroupPress(sub.id)}
                  activeOpacity={0.7}
                >
                  <Image 
                    source={{ uri: sub.image }} 
                    style={styles.subscriptionImage} 
                  />
                </TouchableOpacity>
              ))}
            </View>
          </CategorySection>
        )}

        {(profileData.user_stats.count_all > 0 || 
          profileData.user_stats.count_films > 0 || 
          profileData.user_stats.count_games > 0) && (
          <>
            <StatisticsChart 
              genresData={statisticsData}
              categoriesData={categoryStatisticsData}
            />

            <StatisticsBottomBars
              mostPopularDay={profileData.user_stats.most_pop_day}
              sessionsCreated={profileData.user_stats.count_create_session}
              sessionsSeries={profileData.user_stats.series_session_count}
            />
          </>
        )}

        <View style={{ height: 20 }} />
      </ScrollView>
      <BottomBar />

      {isOwnProfile && (
        <>
          <TileSelectionModal
            visible={tileModalVisible}
            onClose={() => setTileModalVisible(false)}
            selectedTiles={selectedTiles}
            onTilesChange={handleTilesSave}
            stats={{
              all: profileData.user_stats.count_all,
              movies: profileData.user_stats.count_films,
              video_games: profileData.user_stats.count_games,
              board_games: profileData.user_stats.count_table_games,
              other: profileData.user_stats.count_another,
              time: Math.round(profileData.user_stats.spent_time / 60),
            }}
          />

          <EditProfileModal
            visible={editModalVisible}
            onClose={() => setEditModalVisible(false)}
            currentProfile={{
              avatar: { uri: profileData.image },
              name: profileData.name,
              username: profileData.us,
              description: profileData.status,
            }}
            onSave={handleProfileSave}
          />
        </>
      )}
      
      {!isOwnProfile && userId && (
        <InviteToGroupModal
          visible={inviteModalVisible}
          onClose={() => setInviteModalVisible(false)}
          userId={parseInt(userId)}
          onInviteSent={handleInviteSent}
        />
      )}
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
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 16,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  errorText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.red,
  },
  subscriptionsContainer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    gap: 12,
    marginBottom: 16
  },
  subscriptionImage: {
    width: 75,
    height: 75,
    borderRadius: 40,
  },
  emptyText:{
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    alignSelf: 'center',
    marginVertical: 8
  }
});

export default ProfilePage;