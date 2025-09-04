import barsStyle from '@/app/styles/barsStyle';
import { Colors } from '@/constants/Colors';
import { useNavigation, useRoute } from '@react-navigation/native';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

const BottomBar = () => {
  const navigation = useNavigation();
  const route = useRoute();
  const currentRoute = route.name;
  const isActive = (routeName: string) => currentRoute === routeName;
  
  return (
    <View style={styles.container}>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('Register' as never)}>
          <Image style={barsStyle.iconsMenu} source={require("../assets/images/bottom_bar/settings.png")} />
          <Text style={barsStyle.textMenu}>Настройки</Text>
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('GroupsPage' as never)}>
          <Image 
            style={[barsStyle.iconsMenu, isActive('GroupsPage') && { tintColor: Colors.darkGrey }]} 
            source={require("../assets/images/bottom_bar/groups.png")} />
          <Text style={barsStyle.textMenu}>Группы</Text>
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('MainPage' as never)}>
        <Image
          style={[barsStyle.iconsMenu, isActive('MainPage') && { tintColor: Colors.darkGrey }]}
          source={require("../assets/images/bottom_bar/main.png")}
        />
        <Text style={[barsStyle.textMenu, isActive('MainPage') && { color: Colors.darkGrey }]}>Главная</Text>
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('Register' as never)}>
          <Image style={barsStyle.iconsMenu} source={require("../assets/images/bottom_bar/news.png")} />
          <Text style={barsStyle.textMenu}>Новости</Text>
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('Register' as never)}>
          <Image style={barsStyle.iconsMenu} source={require("../assets/images/bottom_bar/profile.png")} />
          <Text style={barsStyle.textMenu}>Профиль</Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    paddingVertical: 40,
    backgroundColor: Colors.white,
  },
});

export default BottomBar;
