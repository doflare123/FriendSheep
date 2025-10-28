import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { StyleSheet, Text, TouchableOpacity } from 'react-native';

const Button = ({ title, onPress }: { title: string; onPress: () => void }) => (
  <TouchableOpacity onPress={onPress} style={styles.button}>
    <Text style={styles.text}>{title}</Text>
  </TouchableOpacity>
);

const styles = StyleSheet.create({
  button: {
    backgroundColor: Colors.lightBlue3,
    borderRadius: 40,
    paddingVertical: 8,
    alignItems: 'center',
    marginTop: 16,
  },
  text: {
    color: Colors.white,
    fontSize: 20,
    fontFamily: Montserrat.regular,
  },
});

export default Button;
