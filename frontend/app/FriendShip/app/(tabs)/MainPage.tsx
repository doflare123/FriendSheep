import CategoryButton from '@/components/CategoryButton';
import { Event } from '@/components/EventCard';
import EventCarousel from '@/components/EventCarousel';
import EventModal from '@/components/EventModal';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import { ImageBackground, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import BottomBar from '../../components/BottomBar';
import TopBar from '../../components/TopBar';

const dummyEvents: Event[] = [
  {
    id: '1',
    title: 'Крестный отец',
    date: '12.02.2004 19:20',
    imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
    description: "ЭЩКЕРЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕЕ",
    genres: ['Драма', 'Криминал'],
    participants: '52/52',
    duration: '175 минут',
    typeEvent: 'Фильм',
    typePlace: 'Онлайн',
    eventPlace: 'https://cinema.com',
    publisher: 'Paramount Pictures',
    publicationDate: 1972,
    ageRating: '18+',
  },
  {
    id: '2',
    title: 'Матрица',
    date: '12.02.2004',
    imageUri: 'https://i.pinimg.com/1200x/cf/ea/47/cfea4764cd43ffe11a177a54b1e5f4b8.jpg',
    description: "мяу",
    genres: ['Фантастика'],
    participants: '48/50',
    duration: '136 минут',
    typeEvent: 'Фильм',
    typePlace: 'Кинотеатр',
    eventPlace: 'Кинотеатр «Октябрь»',
    publisher: 'Warner Bros',
    publicationDate: 1999,
    ageRating: '16+',
  },
];

const MainPage = () => {
  const navigation = useNavigation();

  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [modalVisible, setModalVisible] = useState(false);

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
            onPress={() => {}}
          />
          <EventCarousel
            events={dummyEvents.map(event => ({
              ...event,
              onPress: () => {
                setSelectedEvent(event);
                setModalVisible(true);
              },
            }))}
          />

          <CategoryButton
            title="Новые события"
            imageSource={require('../../assets/images/category/new-pattern.png')}
            onPress={() => {}}
          />
          <EventCarousel
            events={dummyEvents.map(event => ({
              ...event,
              onPress: () => {
                setSelectedEvent(event);
                setModalVisible(true);
              },
            }))}
          />

          <View style={styles.category}>
            <Text style={styles.textCategory}>Категории</Text>
          </View>
          <CategoryButton
            title="Фильмы"
            imageSource={require('../../assets/images/category/movies-pattern.png')}
            onPress={() => {}}
          />
          <CategoryButton
            title="Игры"
            imageSource={require('../../assets/images/category/games-pattern.png')}
            onPress={() => {}}
          />
        </ScrollView>
      </ImageBackground>
      <BottomBar />

      {selectedEvent && (
        <EventModal
          visible={modalVisible}
          onClose={() => setModalVisible(false)}
          event={selectedEvent}
        />
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
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