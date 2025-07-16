import React, { useRef, useState } from 'react';
import { StyleSheet, TextInput, View } from 'react-native';
import { Colors } from '../constants/Colors';

const InputNums = ({ length = 6, onChange }: { length?: number; onChange?: (code: string) => void }) => {
  const [values, setValues] = useState<string[]>(Array(length).fill(''));
  const inputs = useRef<TextInput[]>([]);

const handleChange = (text: string, index: number) => {
  const newValues = [...values];

  if (/^\d$/.test(text)) {
    newValues[index] = text;
    setValues(newValues);
    onChange?.(newValues.join(''));

    if (index < length - 1) {
      inputs.current[index + 1]?.focus();
    }
  } else if (text === '') {
    newValues[index] = '';
    setValues(newValues);
    onChange?.(newValues.join(''));

    if (index > 0) {
      inputs.current[index - 1]?.focus();
    }
  }
};

  const handleKeyPress = (e: any, index: number) => {
    if (e.nativeEvent.key === 'Backspace' && values[index] === '' && index > 0) {
      inputs.current[index - 1]?.focus();
    }
  };

  return (
    <View style={styles.container}>
      {values.map((value, index) => (
        <TextInput
          key={index}
          ref={(ref) => {
            if (ref) inputs.current[index] = ref;
          }}
          value={value}
          onChangeText={(text) => handleChange(text, index)}
          onKeyPress={(e) => handleKeyPress(e, index)}
          keyboardType="number-pad"
          maxLength={1}
          style={styles.input}
          textAlign="center"
        />
      ))}
    </View>
  );
};

export default InputNums;

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 10,
    marginTop: 20,
    marginBottom: 20,
  },
  input: {
    borderWidth: 3,
    borderColor: Colors.blue,
    borderRadius: 14,
    width: 45,
    height: 60,
    fontSize: 24,
    color: Colors.black,
  },
});
