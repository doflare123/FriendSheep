import React from 'react';
import { Image, StyleSheet } from 'react-native';

const Logo = () => (
  <Image
    source={require('../assets/images/logo.png')}
    style={styles.logo}
  />
);

const styles = StyleSheet.create({
  logo: {
    width: 80,
    height: 80,
    borderRadius: 30,
    marginTop: 80,
    marginBottom: 12,
    alignSelf: 'center',
  },
});

export default Logo;
