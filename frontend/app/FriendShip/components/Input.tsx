import React from 'react';
import { StyleSheet, TextInput, TextInputProps, View } from 'react-native';
import { Colors } from '../constants/Colors';

const Input = (props: TextInputProps) => (
  <View style={styles.container}>
    <TextInput
      {...props}
      style={styles.input}
      placeholderTextColor={Colors.lightGrey}
    />
  </View>
);

const styles = StyleSheet.create({
  container: {
    marginBottom: 20,
  },
  input: {
    color: Colors.black,
    borderWidth: 3,
    borderColor: Colors.blue,
    borderRadius: 12,
    padding: 6,
    fontFamily: '../assets/fonts/Inter-Regular.otf',
    fontSize: 18,
  },
});

export default Input;
