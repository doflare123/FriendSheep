import { inter } from '@/constants/Inter';
import { StyleSheet } from 'react-native';
import { Colors } from '../../constants/Colors';

const authorizeStyle = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
    padding: 24,
    position: 'relative',
  },
  topContainer: {
    alignItems: 'center',
    marginTop: 60,
    marginBottom: 20,
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: 'flex-start',
    paddingBottom: 80,
  },
  footer: {
    position: 'absolute',
    bottom: 60,
    left: 0,
    right: 0,
    textAlign: 'center',
    fontSize: 12,
    color: Colors.lightGrey,
    fontFamily: inter.regular,
  },
  title: {
    fontSize: 30,
    fontFamily: inter.regular,
    color: Colors.lightBlue,
    textAlign: 'center',
    marginBottom: 40,
  },
  centerContainer:{
    justifyContent: 'center',
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
    marginTop: -8,
  },
  link: {
    fontFamily: inter.regular,
    fontSize: 12,
    color: Colors.lightBlue,
  },
  terms: {
    marginTop: 12,
    fontFamily: inter.regular,
    fontSize: 12,
    color: Colors.lightGrey,
  },
  doneImage:{
    width: 60,
    height: 60,
    marginBottom: 20,
    alignSelf: 'center'
  }
});

export default authorizeStyle;
