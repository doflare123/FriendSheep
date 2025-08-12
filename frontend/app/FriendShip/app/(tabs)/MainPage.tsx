import CategoryButton from '@/components/CategoryButton';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useNavigation } from '@react-navigation/native';
import React from 'react';
import { ImageBackground, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import BottomBar from '../../components/BottomBar';
import EventCard from '../../components/EventCard';
import TopBar from '../../components/TopBar';

const MainPage = () => {
  const navigation = useNavigation();
  
  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: Colors.white }} edges={['top', 'left', 'right']}>
      <ImageBackground
        source={require('../../assets/images/wallpaper.png')}
        style={{ flex: 1 }}
        resizeMode="cover"
      >
        <TopBar />
        <ScrollView contentContainerStyle={{ paddingBottom: 80 }}>
          <CategoryButton
            title="Популярные события"
            imageSource={require('../../assets/images/category/popular-pattern.png')}
            onPress={() => navigation.navigate('MainPage' as never)}
          />
          <EventCard />
          <CategoryButton
            title="Новые события"
            imageSource={require('../../assets/images/category/new-pattern.png')}
            onPress={() => navigation.navigate('MainPage' as never)}
          />
          <EventCard />
          <View style={styles.category}>
            <Text style={styles.textCategory}>Категории</Text>
          </View>
          <CategoryButton title="Фильмы" imageSource={require('../../assets/images/category/movies-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)} />
          <EventCard />
          <CategoryButton title="Игры" imageSource={require('../../assets/images/category/games-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)} />
          <EventCard />
          <CategoryButton title="Настольные игры" imageSource={require('../../assets/images/category/table_games-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)} />
          <EventCard />
          <CategoryButton title="Другое" imageSource={require('../../assets/images/category/other-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)} />
          <EventCard />
        </ScrollView>
      </ImageBackground>
      <BottomBar />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  background: {
    flex: 1,
  },
  category: {
    backgroundColor: Colors.white,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    overflow: 'hidden',
    margin: 8,
    marginTop: 16,
    height: 40,
    width: '90%',
    alignSelf: 'center',
  },
  textCategory: {
    color: Colors.blue,
    fontSize: 20,
    fontFamily: inter.bold,    
  },
});

export default MainPage;
