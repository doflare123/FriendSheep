import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import { StyleSheet, TextInput, TextInputProps, View } from 'react-native';

interface InputProps extends TextInputProps {
  isValid?: boolean;
  borderColor?: string;
}

const Input = ({ isValid = true, borderColor, ...props }: InputProps) => {
  const colors = useThemedColors();
  
  return (
    <View style={styles.container}>
      <TextInput
        {...props}
        style={[
          styles.input,
          {
            color: colors.black,
            borderColor: borderColor || colors.auth
          },
          !isValid && { borderColor: colors.red }
        ]}
        placeholderTextColor={colors.lightGrey}
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    marginBottom: 20,
  },
  input: {
    borderWidth: 2.5,
    borderRadius: 12,
    padding: 6,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
});

export default Input;