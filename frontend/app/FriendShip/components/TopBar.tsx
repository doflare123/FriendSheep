import barsStyle from '@/app/styles/barsStyle';
import { Colors } from '@/constants/Colors';
import React from 'react';
import { Image, StyleSheet, View } from 'react-native';
import MainSearchBar from './MainSearchBar';



const TopBar = () => {
  return (
    <View style={styles.container}>
        <MainSearchBar/>
        <Image style={barsStyle.notifications} source={require("../assets/images/top_bar/notifications.png")}/>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.white,
    paddingHorizontal: 16,
    paddingVertical: 8,
    elevation: 4,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    zIndex: 10,
  },
});

export default TopBar;
