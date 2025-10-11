import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { Image, StyleSheet, Text, View } from 'react-native';

const Button = ( {title} : { title: string; }) => (
    <View style={styles.container}>
        <Image source={require('@/assets/images/event_card/movie.png')} style={styles.iconImage}/>
        <Text style={styles.text}>{title} -</Text>
        <Text style={styles.count}>20</Text>
    </View>
);

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.lightBlue2,
    flexDirection: 'row',    
    borderRadius: 40,
    paddingVertical: 8,
    padding: 24,
    marginRight: 8,
    marginVertical: 8,
    alignItems: 'center',
  },
  text: {
    color: Colors.black,
    fontSize: 20,
    fontFamily: Montserrat.regular,
  },
  iconImage:{
    width: 30,
    height: 30,
    resizeMode: 'contain',
    marginRight: 12,
  },
  count:{
    color: Colors.black,
    fontSize: 20,
    fontFamily: Montserrat.regular,
    marginLeft: 6
  }
});

export default Button;
