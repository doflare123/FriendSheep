import { inter } from '@/constants/Inter';
import React, { useEffect, useRef, useState } from 'react';
import { Platform, StyleSheet, TextInput, TouchableOpacity, View } from 'react-native';
import { Colors } from '../constants/Colors';

interface InputSmallProps {
  length?: number;
  onChange?: (code: string) => void;
}

const InputSmall: React.FC<InputSmallProps> = ({ length = 6, onChange }) => {
  const [values, setValues] = useState<string[]>(Array(length).fill(''));
  const hiddenInputRef = useRef<TextInput>(null);
  const [activeIndex, setActiveIndex] = useState(0);

  const updateValues = (newValues: string[]) => {
    setValues(newValues);
    onChange?.(newValues.join(''));
  };

  const handleChangeText = (text: string) => {
    const filtered = text.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, length);
    const newValues = Array(length).fill('');
    for (let i = 0; i < filtered.length; i++) {
      newValues[i] = filtered[i];
    }
    updateValues(newValues);

    if (filtered.length < length) {
      setActiveIndex(filtered.length);
    } else {
      setActiveIndex(length - 1);
    }
  };

  const focusHiddenInput = (index: number) => {
    setActiveIndex(index);
    hiddenInputRef.current?.focus();
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      hiddenInputRef.current?.focus();
    }, Platform.OS === 'android' ? 300 : 100);
    return () => clearTimeout(timer);
  }, []);

  return (
    <View style={styles.container}>
      <TextInput
        ref={hiddenInputRef}
        style={styles.hiddenInput}
        value={values.join('')}
        onChangeText={handleChangeText}
        keyboardType="default"
        autoCapitalize="characters"
        autoCorrect={false}
        maxLength={length}
        editable
        showSoftInputOnFocus
        blurOnSubmit={false}
      />

      {values.map((char, index) => (
        <TouchableOpacity
          key={index}
          style={[
            styles.input,
            activeIndex === index && styles.inputActive
          ]}
          onPress={() => focusHiddenInput(index)}
          activeOpacity={0.8}
        >
          <TextInput
            style={styles.fakeInputText}
            value={char}
            editable={false}
            pointerEvents="none"
          />
        </TouchableOpacity>
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
  hiddenInput: {
    position: 'absolute',
    height: 1,
    width: 1,
    opacity: 0.02,
  },
  input: {
    fontFamily: inter.regular,
    borderWidth: 3,
    borderColor: Colors.blue,
    borderRadius: 14,
    width: 45,
    height: 60,
    justifyContent: 'center',
    alignItems: 'center',
  },
  inputActive: {
    borderColor: Colors.lightBlue,
  },
  fakeInputText: {
    fontSize: 24,
    color: Colors.black,
    textAlign: 'center',
    fontFamily: inter.regular,
  },
});
