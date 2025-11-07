import userService from '@/api/services/userService';
import { Subscription, UserProfile } from '@/api/types/user';
import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import PageHeader from '@/components/PageHeader';
import EditProfileModal from '@/components/profile/EditProfileModal';
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

  const isOwnProfile = !userId;

  const loadProfileData = useCallback(async () => {
    try {
      setLoading(true);

      const profile = isOwnProfile
        ? await userService.getCurrentUserProfile()
        : await userService.getUserProfileById(userId);

      setProfileData(profile);

    const tiles = profile.tiles.map(tile => {
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
        const subs = await userService.getUserSubscriptions(
          userId ? parseInt(userId) : undefined
        );
        setSubscriptions(subs);
      } catch (error) {
        console.warn('Не удалось загрузить подписки:', error);
      }

    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось загрузить профиль',
      });
    } finally {
      setLoading(false);
    }
  }, [userId, isOwnProfile, showToast]);

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

  const handleProfileSave = async (updatedProfile: any) => {
    try {
      setLoading(true);

      let imageUrl = profileData?.image;
      
      if (updatedProfile.avatar && typeof updatedProfile.avatar === 'object' && updatedProfile.avatar.uri) {
        console.log('[ProfilePage] Загрузка нового изображения:', updatedProfile.avatar.uri);
        
        try {
          const formData = new FormData();
          const uriParts = updatedProfile.avatar.uri.split('.');
          const fileType = uriParts[uriParts.length - 1];
          
          formData.append('image', {
            uri: updatedProfile.avatar.uri,
            name: `avatar.${fileType}`,
            type: `image/${fileType}`,
          } as any);

          imageUrl = updatedProfile.avatar.uri;
        } catch (uploadError) {
          console.error('[ProfilePage] Ошибка загрузки изображения:', uploadError);
          throw new Error('Не удалось загрузить изображение');
        }
      }

      await userService.updateProfile({
        name: updatedProfile.name,
        us: updatedProfile.username,
        status: updatedProfile.description,
        image: imageUrl,
      });

      showToast({
        type: 'success',
        title: 'Успешно',
        message: 'Профиль обновлён',
      });

      await loadProfileData();
    } catch (error: any) {
      console.error('[ProfilePage] Ошибка обновления профиля:', error);
      
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


  const handleGroupSelect = (groupId: string) => {
    console.log('Inviting user to group:', groupId);
    setInviteModalVisible(false);
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
    const stats = profileData.user_stats;
    const total = stats.count_films + stats.count_games + stats.count_table_games + stats.count_another;
    
    if (total === 0) return [];
    
    const data = [];
    
    if (stats.count_films > 0) {
      data.push({
        name: 'Фильмы',
        percentage: Math.round((stats.count_films / total) * 100),
        color: '#FF6B6B',
        legendFontColor: '#000'
      });
    }
    
    if (stats.count_games > 0) {
      data.push({
        name: 'Игры',
        percentage: Math.round((stats.count_games / total) * 100),
        color: '#4ECDC4',
        legendFontColor: '#000'
      });
    }
    
    if (stats.count_table_games > 0) {
      data.push({
        name: 'Настолки',
        percentage: Math.round((stats.count_table_games / total) * 100),
        color: '#45B7D1',
        legendFontColor: '#000'
      });
    }
    
    if (stats.count_another > 0) {
      data.push({
        name: 'Другое',
        percentage: Math.round((stats.count_another / total) * 100),
        color: '#FFA07A',
        legendFontColor: '#000'
      });
    }
    
    return data;
  };

  const statisticsData = generateUserStatistics();

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
            time: 0,
          }}
        />

        <CategorySection title="Любимые жанры:" marginBottom={16}>
          {profileData.popular_genres && profileData.popular_genres.length > 0 ? (
            <View style={styles.genresContainer}>
              <View style={styles.genresColumn}>
                {profileData.popular_genres.slice(0, 3).map((genre, index) => (
                  <Text key={index} style={styles.genreItem}>
                    {index + 1}. {genre.name} - {genre.count}
                  </Text>
                ))}
              </View>
              <View style={styles.genresColumn}>
                {profileData.popular_genres.slice(3).map((genre, index) => (
                  <Text key={index} style={styles.genreItem}>
                    {index + 4}. {genre.name} - {genre.count}
                  </Text>
                ))}
              </View>
            </View>
          ) : (
            <Text style={styles.emptyText}>
              Пока нет любимых жанров
            </Text>
          )}
        </CategorySection>

        {subscriptions.length > 0 && (
          <CategorySection title="Подписки:" marginBottom={16}>
            <View style={styles.subscriptionsContainer}>
              {subscriptions.map((sub) => (
                <Image 
                  key={sub.id} 
                  source={{ uri: sub.image }} 
                  style={styles.subscriptionImage} 
                />
              ))}
            </View>
          </CategorySection>
        )}

        <CategorySection 
          title="Сессии:" 
          customActionButton={{
            icon: sessionFilter === 'completed' 
              ? require('@/assets/images/profile/finished.png')
              : require('@/assets/images/profile/current.png'),
            onPress: () => setSessionFilter(
              sessionFilter === 'completed' ? 'upcoming' : 'completed'
            ),
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

        {(profileData.user_stats.count_all > 0 || 
          profileData.user_stats.count_films > 0 || 
          profileData.user_stats.count_games > 0) && (
          <>
            <StatisticsChart 
              statisticsData={statisticsData}
            />

            <StatisticsBottomBars
              mostPopularDay={profileData.user_stats.most_pop_day || '—'}
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
              time: 0,
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

      {!isOwnProfile && (
        <InviteToGroupModal
          visible={inviteModalVisible}
          onClose={() => setInviteModalVisible(false)}
          onGroupSelect={handleGroupSelect}
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
  emptyText:{
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    alignSelf: 'center',
    marginVertical: 8
  }
});

export default ProfilePage;