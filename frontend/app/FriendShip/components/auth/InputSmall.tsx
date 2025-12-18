import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useEffect, useRef, useState } from 'react';
import { Platform, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';

interface InputSmallProps {
  value?: string;
  onChangeText?: (text: string) => void;
  editable?: boolean;
  length?: number;
}

const InputSmall: React.FC<InputSmallProps> = ({ 
  value, 
  onChangeText, 
  editable = true,
  length = 6 
}) => {
  const colors = useThemedColors();
  const [values, setValues] = useState<string[]>(Array(length).fill(''));
  const hiddenInputRef = useRef<TextInput>(null);
  const [activeIndex, setActiveIndex] = useState(0);

  useEffect(() => {
    if (value !== undefined) {
      const newValues = Array(length).fill('');
      for (let i = 0; i < Math.min(value.length, length); i++) {
        newValues[i] = value[i];
      }
      setValues(newValues);
      setActiveIndex(Math.min(value.length, length - 1));
    }
  }, [value, length]);

  const updateValues = (newValues: string[]) => {
    setValues(newValues);
    onChangeText?.(newValues.join(''));
  };

  const handleChangeText = (text: string) => {
    if (!editable) return;
    
    const filtered = text.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, length);
    const newValues = Array(length).fill('');
    for (let i = 0; i < filtered.length; i++) newValues[i] = filtered[i];
    updateValues(newValues);
    setActiveIndex(Math.min(filtered.length, length - 1));
  };

  const ensureFocus = (index: number) => {
    if (!editable) return;
    
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
    if (editable) {
      const id = setTimeout(() => hiddenInputRef.current?.focus(), Platform.OS === 'android' ? 300 : 100);
      return () => clearTimeout(id);
    }
  }, [editable]);

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
        editable={editable}
      />

      {values.map((char, index) => (
        <TouchableOpacity
          key={index}
          style={[
            styles.input,
            { borderColor: colors.blue },
            activeIndex === index && { borderColor: colors.lightBlue },
            !editable && [styles.inputDisabled, { borderColor: colors.grey }]
          ]}
          activeOpacity={0.8}
          onPressIn={() => ensureFocus(index)}
          disabled={!editable}
        >
          <Text style={[
            styles.fakeInputText,
            { color: colors.black },
            !editable && { color: colors.grey }
          ]}>
            {char}
          </Text>
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
    borderRadius: 14,
    width: 45,
    height: 60,
    justifyContent: 'center',
    alignItems: 'center',
  },
  inputDisabled: {
    opacity: 0.5,
  },
  fakeInputText: {
    fontSize: 24,
    textAlign: 'center',
    fontFamily: Montserrat.regular,
  },
});