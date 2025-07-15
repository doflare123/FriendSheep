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
    fontSize: 35,
    fontFamily: '../assets/fonts/Inter-Regular.otf',
    color: Colors.lightBlue,
    textAlign: 'center',
    marginBottom: 40,
  },
  label: {
    fontSize: 21,
    marginBottom: 8,
  },
  account: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    marginBottom: 12,
    marginTop: -8,
  },
  link: {
    fontSize: 12,
    color: Colors.lightBlue,
  },
  terms: {
    fontSize: 12,
    color: Colors.lightGrey,
  },
  footer: {
    position: 'absolute',
    bottom: 45,
    left: 0,
    right: 0,
    textAlign: 'center',
    fontSize: 12,
    color: Colors.lightGrey,
    paddingVertical: 10,
  }
});

export default authorizeStyle;
