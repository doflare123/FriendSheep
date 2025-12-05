import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { StyleSheet, TextInput, TextInputProps, View } from 'react-native';

interface InputProps extends TextInputProps {
  isValid?: boolean;
  borderColor?: string;
}

const Input = ({ isValid = true, borderColor, ...props }: InputProps) => (
  <View style={styles.container}>
    <TextInput
      {...props}
      style={[
        styles.input,
        !isValid && styles.inputInvalid,
        borderColor && { borderColor }
      ]}
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
  inputInvalid: {
    borderColor: Colors.red,
  },
});

export default Input;