import { Montserrat } from '@/constants/Montserrat';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
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
    fontFamily: Montserrat.regular,
  },
  title: {
    fontSize: 28,
    fontFamily: Montserrat_Alternates.medium,
    color: Colors.lightBlue3,
    textAlign: 'center',
    marginBottom: 40,
  },
  label: {
    fontFamily: Montserrat.regular,
    fontSize: 18,
    marginBottom: 4,
  },
  hintLabel: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    marginBottom: 8,
    color: Colors.lightGrey,
    textAlign: 'center'
  },
  account: {
    fontFamily: Montserrat.regular,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: -8,
  },
  link: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.lightBlue,
  },
  terms: {
    marginTop: 12,
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.lightGrey,
  },
  doneImage:{
    width: 60,
    height: 60,
    marginBottom: 20,
    alignSelf: 'center'
  },
  password:{
    color: Colors.lightBlue2,
    fontSize: 12
  }
});

export default authorizeStyle;
