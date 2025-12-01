import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { StyleSheet, TextInput, TextInputProps, View } from 'react-native';

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
    borderWidth: 2.5,
    borderColor: Colors.blue3,
    borderRadius: 12,
    padding: 6,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
});

export default Input;
