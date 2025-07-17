import { inter } from '@/constants/Inter';
import { StyleSheet } from 'react-native';
import { Colors } from '../../constants/Colors';

const authorizeStyle = StyleSheet.create({
  container: {
    padding: 24,
    flex: 1,
    justifyContent: 'center',
    backgroundColor: Colors.white,
  },
  title: {
    fontSize: 30,
    fontFamily: inter.regular,
    color: Colors.lightBlue,
    textAlign: 'center',
    marginBottom: 40,
  },
  label: {
    fontFamily: inter.regular,
    fontSize: 20,
    marginBottom: 8,
  },
  hintLabel: {
    fontFamily: inter.regular,
    fontSize: 16,
    marginBottom: 8,
    color: Colors.lightGrey,
    textAlign: 'center'
  },
  account: {
    fontFamily: inter.regular,
    flexDirection: 'row',
    justifyContent: 'flex-end',
    marginBottom: 12,
    marginTop: -8,
  },
  link: {
    fontFamily: inter.regular,
    fontSize: 12,
    color: Colors.lightBlue,
  },
  terms: {
    fontFamily: inter.regular,
    fontSize: 12,
    color: Colors.lightGrey,
  },
  footer: {
    fontFamily: inter.regular,
    position: 'absolute',
    bottom: 45,
    left: 0,
    right: 0,
    textAlign: 'center',
    fontSize: 12,
    color: Colors.lightGrey,
    paddingVertical: 10,
  },
  doneImage:{
    width: 60,
    height: 60,
    marginBottom: 20,
    alignSelf: 'center'
  }
});

export default authorizeStyle;
