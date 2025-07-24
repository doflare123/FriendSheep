import { inter } from '@/constants/Inter';
import React, { useEffect, useRef, useState } from 'react';
import { NativeSyntheticEvent, StyleSheet, TextInput, TextInputKeyPressEventData, View } from 'react-native';
import { Colors } from '../constants/Colors';

const InputSmall = ({ length = 6, onChange } : { length?: number; onChange?: (code: string) => void}) => {
  
  const [values, setValues] = useState<string[]>(Array(length).fill(''));
  const inputs = useRef<TextInput[]>([]);
  const lastKey = useRef('');
  const backspaceInterval = useRef<number | null>(null);

  useEffect(() => {
    inputs.current[0]?.focus();
  }, []);

  const updateValues = (newValues: string[]) => {
    setValues(newValues);
    onChange?.(newValues.join(''));
  };

  const handlePaste = (text: string) => {
    const filtered = text.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, length);
    const newValues = Array(length).fill('');
    for (let i = 0; i < filtered.length; i++) {
      newValues[i] = filtered[i];
    }
    updateValues(newValues);
    if (filtered.length < length) {
      inputs.current[filtered.length]?.focus();
    } else {
      inputs.current[length - 1]?.blur();
    }
  };

  const handleChange = (text: string, index: number) => {
    const filtered = text.toUpperCase().replace(/[^A-Z0-9]/g, '');

    if (filtered.length > 1) {
      handlePaste(filtered);
      return;
    }

    const newValues = [...values];
    newValues[index] = filtered;
    updateValues(newValues);

    if (filtered && index < length - 1) {
      inputs.current[index + 1]?.focus();
    }
  };

  const startBackspaceDeletion = (index: number) => {
    let currentIndex = index;

    const deleteNext = () => {
      const newValues = [...values];
      while (currentIndex >= 0) {
        if (newValues[currentIndex] !== '') {
          newValues[currentIndex] = '';
          updateValues(newValues);
          inputs.current[currentIndex]?.focus();
          currentIndex--;
          break;
        }
        currentIndex--;
      }

      if (currentIndex >= 0) {
        backspaceInterval.current = setTimeout(deleteNext, 100);
      }
    };

    deleteNext();
  };

  const stopBackspace = () => {
    if (backspaceInterval.current) {
      clearTimeout(backspaceInterval.current);
      backspaceInterval.current = null;
    }
  };

  const handleKeyPress = (
    e: NativeSyntheticEvent<TextInputKeyPressEventData>,
    index: number
  ) => {
    const key = e.nativeEvent.key;
    lastKey.current = key;

    if (key === 'Backspace') {
      if (values[index] === '') {
        if (!backspaceInterval.current) {
          startBackspaceDeletion(index - 1);
        }
      } else {
        const newValues = [...values];
        newValues[index] = '';
        updateValues(newValues);
      }
    } else {
      stopBackspace();
    }
  };

  return (
    <View style={styles.container}>
      {values.map((value, id) => (
        <TextInput
          key={id}
          ref={(ref) => {
            if (ref) inputs.current[id] = ref;
          }}
          value={value}
          onChangeText={(text) => handleChange(text, id)}
          onKeyPress={(e) => handleKeyPress(e, id)}
          onBlur={stopBackspace}
          autoCapitalize="characters"
          autoCorrect={false}
          keyboardType="default"
          maxLength={1}
          style={styles.input}
          textAlign="center"
        />
      ))}
    </View>
  );
};

export default InputSmall;

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 10,
    marginVertical: 20,
  },
  input: {
    fontFamily: inter.regular,
    borderWidth: 3,
    borderColor: Colors.blue,
    borderRadius: 14,
    width: 45,
    height: 60,
    fontSize: 24,
    color: Colors.black,
  },
});
