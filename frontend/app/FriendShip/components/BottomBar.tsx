import barsStyle from '@/app/styles/barsStyle';
import { Colors } from '@/constants/Colors';
import { useNavigation, useRoute } from '@react-navigation/native';
import React from 'react';
import { Image, StyleSheet, TouchableOpacity, View } from 'react-native';

const BottomBar = () => {
  const navigation = useNavigation();
  const route = useRoute();
  const currentRoute = route.name;
  const isActive = (routeName: string) => currentRoute === routeName;
  
  return (
    <View style={styles.container}>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('Register' as never)}>
          <Image style={barsStyle.iconsMenu} source={require("@/assets/images/bottom_bar/settings.png")} />
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('GroupsPage' as never)}>
          <Image 
            style={[barsStyle.iconsMenu, (isActive('GroupsPage') || isActive('GroupPage') || isActive('GroupManagePage') || isActive('GroupSearchPage')) && { tintColor: Colors.darkGrey }]} 
            source={require("@/assets/images/bottom_bar/groups.png")} />
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('MainPage' as never)}>
        <Image
          style={[barsStyle.iconsMenu, (isActive('MainPage') || isActive('CategoryPage')) && { tintColor: Colors.darkGrey }]}
          source={require("@/assets/images/bottom_bar/main.png")}
        />
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('Register' as never)}>
          <Image style={barsStyle.iconsMenu} source={require("@/assets/images/bottom_bar/news.png")} />
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('ProfilePage' as never)}>
          <Image style={[barsStyle.iconsMenu, {width: 25, height: 25}, isActive('ProfilePage') && { tintColor: Colors.darkGrey }]} source={require("@/assets/images/bottom_bar/profile.png")} />
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'center',
    paddingVertical: 10,
    backgroundColor: Colors.white,
  },
});

export default BottomBar;
