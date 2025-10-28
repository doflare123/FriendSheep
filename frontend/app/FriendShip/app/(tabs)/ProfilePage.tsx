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
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { getCurrentUser, getUserById } from '@/data/users';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import React, { useState } from 'react';
import {
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
  const userId = route.params?.userId;
  
  const { sortingState, sortingActions } = useSearchState();
  const [sessionFilter, setSessionFilter] = useState<'completed' | 'upcoming'>('upcoming');
  const [tileModalVisible, setTileModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [inviteModalVisible, setInviteModalVisible] = useState(false);

  const isOwnProfile = !userId;

  const userData = isOwnProfile ? getCurrentUser() : getUserById(userId);

  const [profileData, setProfileData] = useState(userData);
  const [selectedTiles, setSelectedTiles] = useState<TileType[]>(userData.selectedTiles);

  const handleEditProfile = () => {
    setEditModalVisible(true);
  };

  const handleChangeTiles = () => {
    setTileModalVisible(true);
  };

  const handleInviteToGroup = () => {
    setInviteModalVisible(true);
  };

  const handleProfileSave = (updatedProfile: any) => {
    setProfileData(prev => ({
      ...prev,
      ...updatedProfile,
    }));
  };

  const handleGroupSelect = (groupId: string) => {
    console.log('Inviting user to group:', groupId);
    setInviteModalVisible(false);
  };

  // Временная кнопка для тестирования
  const handleTestOtherProfile = () => {
    navigation.push('ProfilePage', { userId: '1' });
  };

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <PageHeader 
          title={isOwnProfile ? "Ваш профиль" : profileData.name} 
          showWave 
        />
        
        {/* ВРЕМЕННАЯ КНОПКА ДЛЯ ТЕСТА*/}
        {isOwnProfile && (
          <TouchableOpacity 
            style={styles.testButton}
            onPress={handleTestOtherProfile}
          >
            <Text style={styles.testButtonText}>
              🧪 Посмотреть профиль другого пользователя (тест)
            </Text>
          </TouchableOpacity>
        )}
        
        <ProfileHeader 
          avatar={profileData.avatar}
          name={profileData.name}
          username={profileData.username}
          description={profileData.description}
          registrationDate={profileData.registrationDate}
          telegramLink={profileData.telegramLink}
          isOwnProfile={isOwnProfile}
          onEditProfile={isOwnProfile ? handleEditProfile : undefined}
          onChangeTiles={isOwnProfile ? handleChangeTiles : undefined}
          onInviteToGroup={!isOwnProfile ? handleInviteToGroup : undefined}
        />

        <ProfileStats 
          selectedTiles={selectedTiles}
          stats={profileData.stats}
        />

        <CategorySection title="Любимые жанры:" marginBottom={16}>
          <View style={styles.genresContainer}>
            <View style={styles.genresColumn}>
              {profileData.favoriteGenres.slice(0, 3).map((genre, index) => (
                <Text key={index} style={styles.genreItem}>
                  {index + 1}. {genre.name} - {genre.count}
                </Text>
              ))}
            </View>
            <View style={styles.genresColumn}>
              {profileData.favoriteGenres.slice(3).map((genre, index) => (
                <Text key={index} style={styles.genreItem}>
                  {index + 4}. {genre.name} - {genre.count}
                </Text>
              ))}
            </View>
          </View>
        </CategorySection>

        <CategorySection title="Подписки:" marginBottom={16}>
          <View style={styles.subscriptionsContainer}>
            {profileData.subscriptions.map((sub) => (
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
          events={sessionFilter === 'completed' ? profileData.completedSessions : profileData.upcomingSessions}
        />

        <StatisticsChart 
          statisticsData={profileData.statisticsData}
        />

        <StatisticsBottomBars
          mostPopularDay="Воскресенье"
          sessionsCreated={4}
          sessionsSeries={4}
        />

        <View style={{ height: 20 }} />
      </ScrollView>
      <BottomBar />

      {isOwnProfile && (
        <>
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
  testButton: {
    backgroundColor: '#FF6B6B',
    marginHorizontal: 16,
    marginVertical: 10,
    padding: 12,
    borderRadius: 12,
    alignItems: 'center',
  },
  testButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.white,
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