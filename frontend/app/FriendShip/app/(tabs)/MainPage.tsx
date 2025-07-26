import CategoryButton from '@/components/CategoryButton';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useNavigation } from '@react-navigation/native';
import React from 'react';
import { ImageBackground, ScrollView, StyleSheet, Text, View } from 'react-native';
import BottomBar from '../../components/BottomBar';
import EventCard from '../../components/EventCard';
import TopBar from '../../components/TopBar';

const MainPage = () => {
  const navigation = useNavigation();
  
  return (
    <ImageBackground
      source={require('../../assets/images/wallpaper.png')}
      style={styles.background}
      resizeMode="cover"
    >
        <TopBar />
        <ScrollView>
          <CategoryButton title="Популярные события" imageSource={require('../../assets/images/category/popular-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)}/>
          <EventCard />
          <CategoryButton title="Новые события" imageSource={require('../../assets/images/category/new-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)}/>
          <EventCard />
          <View style={styles.category}><Text style={styles.textCategory}>Категории</Text></View>
          <CategoryButton title="Фильмы" imageSource={require('../../assets/images/category/movies-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)}/>
          <EventCard />
          <CategoryButton title="Игры" imageSource={require('../../assets/images/category/games-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)}/>
          <EventCard />
          <CategoryButton title="Настольные игры" imageSource={require('../../assets/images/category/table_games-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)}/>
          <EventCard />
          <CategoryButton title="Другое" imageSource={require('../../assets/images/category/other-pattern.png')} onPress={() => navigation.navigate('MainPage' as never)}/>
          <EventCard />
        </ScrollView>
        <BottomBar />
    </ImageBackground>
  );
};

const styles = StyleSheet.create({
  background: {
    flex: 1,
  },
  header: {
    fontSize: 24,
    fontWeight: 'bold',
    margin: 16,
  },
  subHeader: {
    fontSize: 20,
    fontWeight: 'bold',
    marginLeft: 16,
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
  }
});

export default MainPage;
