import { inter } from '@/constants/Inter';
import { StyleSheet } from 'react-native';
import { Colors } from '../../constants/Colors';

const barsStyle = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
    padding: 24,
    position: 'relative',
  },
  switch: {
    width: 25,
    height: 25,
    resizeMode: 'contain',
    alignSelf: 'center'
  },
  options: {
    width: 25,
    height: 25,
    resizeMode: 'contain',
    alignSelf: 'center'
  },
  notifications: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
    marginLeft: 10,
    alignSelf: 'center'
  },
  menu: {
    width: 55,
    alignItems: 'center'
  },
  iconsMenu: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
    alignSelf: 'center',
    marginTop: 0,  // убираем
  },
  textMenu: {
    fontFamily: inter.regular,
    color: Colors.lightGreyBlue,
    fontSize: 10,
  }
});

export default barsStyle;
