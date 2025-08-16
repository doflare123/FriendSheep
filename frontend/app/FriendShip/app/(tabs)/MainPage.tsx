import CategoryButton from '@/components/CategoryButton';
import EventCarousel from '@/components/EventCarousel';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useNavigation } from '@react-navigation/native';
import React from 'react';
import { ImageBackground, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import BottomBar from '../../components/BottomBar';
import TopBar from '../../components/TopBar';

const dummyEvents = [
  { id: '1', title: 'Крестный отец', date: '12.02.2004', imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg', genres: ['Драма', 'Криминал'], participants: '52/52', duration: '175 минут' },
  { id: '2', title: 'Матрица', date: '12.02.2004', imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg', genres: ['Фантастика'], participants: '48/50', duration: '136 минут' },
];

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
        <CategoryButton title="Популярные события" imageSource={require('../../assets/images/category/popular-pattern.png')} onPress={() => {}} />
        <EventCarousel events={dummyEvents} />

        <CategoryButton title="Новые события" imageSource={require('../../assets/images/category/new-pattern.png')} onPress={() => {}} />
        <EventCarousel events={dummyEvents} />

        <View style={styles.category}>
          <Text style={styles.textCategory}>Категории</Text>
        </View>
        <CategoryButton title="Фильмы" imageSource={require('../../assets/images/category/movies-pattern.png')} onPress={() => {}} />
        <CategoryButton title="Игры" imageSource={require('../../assets/images/category/games-pattern.png')} onPress={() => {}} />
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
