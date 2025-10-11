import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import PageHeader from '@/components/PageHeader';
import StatisticsBar from '@/components/profile/StatisticsBar';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import React from 'react';
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

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <PageHeader title="Ваш профиль" showWave />
        
        <View style={styles.header}>
          <Image source={require('@/assets/images/profile/profile_avatar.jpg')} style={styles.profileImage}/>
          <View style={{flexDirection: 'column'}}>
            <View style={styles.headerInfo}>
              <Text style={styles.profileName}>Pibble</Text>
              <Text style={styles.profileUs}>@washMyBelly</Text>
              <Text style={styles.description}>Hi! I am Pibble. Wash my Belly 🧽</Text>
              <Text style={styles.registrationDate}>20.20.2020</Text>
            </View>
          </View>
        </View>
        <View style={{flexDirection: 'row', marginHorizontal: 16}}>
            <StatisticsBar title='Медиа'></StatisticsBar>
            <StatisticsBar title='Игры'></StatisticsBar>
        </View>

        <CategorySection title="Любимые жанры:">
        </CategorySection>

        <CategorySection title="Подписки:">
        </CategorySection>

        <CategorySection title="Сессии:">
        </CategorySection>

        <CategorySection title="Статистика:">
        </CategorySection>

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
    alignItems: 'center',
  },
  headerInfo: {
    flex: 1,
  },
  profileImage: {
    width: 130,
    height: 130,
    borderRadius: 100,
    marginRight: 20,
  },
  profileName: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.black,
  },
  profileUs: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    marginBottom: 8,
  },
  description: {
    width: 200,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
  },
  registrationDate:{
    fontFamily: Montserrat.regular,
    color: Colors.lightGrey,
    alignSelf: 'flex-end'
  }
});

export default ProfilePage;