import barsStyle from '@/app/styles/barsStyle';
import { useThemedColors } from '@/hooks/useThemedColors';
import { CommonActions, useNavigation, useRoute } from '@react-navigation/native';
import React from 'react';
import { Image, StyleSheet, TouchableOpacity, View } from 'react-native';

const BottomBar = () => {
  const navigation = useNavigation();
  const route = useRoute();
  const currentRoute = route.name;
  const isActive = (routeName: string) => currentRoute === routeName;

  const colors = useThemedColors();

  const handleProfilePress = () => {
    if (currentRoute === 'ProfilePage') {
      navigation.dispatch(
        CommonActions.reset({
          index: 0,
          routes: [{ name: 'ProfilePage' }],
        })
      );
    } else {
      navigation.navigate('ProfilePage' as never);
    }
  };
  
  return (
    <View style={[styles.container, { backgroundColor: colors.white }]}>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('SettingsPage' as never)}>
          <Image style={[barsStyle.iconsMenu, {tintColor: colors.lightGreyBlue}, (isActive('SettingsPage') || isActive('AboutPage')) && { tintColor: colors.darkGrey }]} source={require("@/assets/images/bottom_bar/settings.png")} />
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('GroupsPage' as never)}>
          <Image 
            style={[barsStyle.iconsMenu, {tintColor: colors.lightGreyBlue}, (isActive('GroupsPage') || isActive('GroupPage') || isActive('GroupManagePage') || isActive('GroupSearchPage') || isActive('AllGroupsPage')) && { tintColor: colors.darkGrey }]} 
            source={require("@/assets/images/bottom_bar/groups.png")} />
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={() => navigation.navigate('MainPage' as never)}>
        <Image
          style={[barsStyle.iconsMenu, {tintColor: colors.lightGreyBlue}, (isActive('MainPage') || isActive('CategoryPage')) && { tintColor: colors.darkGrey }]}
          source={require("@/assets/images/bottom_bar/main.png")}
        />
      </TouchableOpacity>
      <TouchableOpacity style={barsStyle.menu} onPress={handleProfilePress}>
        <Image 
          style={[barsStyle.iconsMenu, {width: 25, height: 25, tintColor: colors.lightGreyBlue}, (isActive('ProfilePage') || isActive('UserSearchPage') || isActive('AllEventsPage')) && { tintColor: colors.darkGrey }]} 
          source={require("@/assets/images/bottom_bar/profile.png")} />
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
  },
});

export default BottomBar;