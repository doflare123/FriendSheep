import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React, { useEffect, useRef, useState } from 'react';
import { Platform, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';

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
    for (let i = 0; i < filtered.length; i++) newValues[i] = filtered[i];
    updateValues(newValues);
    setActiveIndex(Math.min(filtered.length, length - 1));
  };

  const ensureFocus = (index: number) => {
    setActiveIndex(index);
    const input = hiddenInputRef.current;
    if (!input) return;

    if (Platform.OS === 'android') {
      if (input.isFocused && input.isFocused()) {
        input.blur();
        setTimeout(() => input.focus(), 0);
      } else {
        input.focus();
      }
    } else {
      input.focus();
    }
  };

  useEffect(() => {
    const id = setTimeout(() => hiddenInputRef.current?.focus(), Platform.OS === 'android' ? 300 : 100);
    return () => clearTimeout(id);
  }, []);

  const joined = values.join('');
  const caretPos = joined.length;

  return (
    <View style={styles.container}>
      <TextInput
        ref={hiddenInputRef}
        style={styles.hiddenInput}
        value={joined}
        onChangeText={handleChangeText}
        keyboardType="default"
        autoCapitalize="characters"
        autoCorrect={false}
        maxLength={length}
        selection={{ start: caretPos, end: caretPos }}
        caretHidden
      />

      {values.map((char, index) => (
        <TouchableOpacity
          key={index}
          style={[styles.input, activeIndex === index && styles.inputActive]}
          activeOpacity={0.8}
          onPressIn={() => ensureFocus(index)}
        >
          <Text style={styles.fakeInputText}>{char}</Text>
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
    top: -1000,
    left: 0,
    width: 200,
    height: 50,
    color: 'transparent',
    backgroundColor: 'transparent',
    includeFontPadding: false,
    padding: 0,
  },
  input: {
    fontFamily: Montserrat.regular,
    borderWidth: 3,
    borderColor: Colors.blue3,
    borderRadius: 14,
    width: 45,
    height: 60,
    justifyContent: 'center',
    alignItems: 'center',
  },
  inputActive: {
    borderColor: Colors.lightBlue3,
  },
  fakeInputText: {
    fontSize: 24,
    color: Colors.black,
    textAlign: 'center',
    fontFamily: Montserrat.regular,
  },
});
