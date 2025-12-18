import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import { StyleSheet, Text, TextInput } from 'react-native';

interface CityFilterInputProps {
  value: string;
  onChangeText: (text: string) => void;
}

const CityFilterInput: React.FC<CityFilterInputProps> = ({ value, onChangeText }) => {
  const colors = useThemedColors();
  
  return (
    <>
      <Text style={[styles.dropdownTitle, {color: colors.black}]}>Фильтрация по городу</Text>
      <TextInput
        style={[styles.cityInput, {color: colors.black, backgroundColor: colors.veryLightGrey}]}
        placeholder="Введите город..."
        value={value}
        onChangeText={onChangeText}
        placeholderTextColor={colors.grey}
      />
    </>
  );
};

const styles = StyleSheet.create({
  dropdownTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    marginBottom: 8,
  },
  cityInput: {
    borderRadius: 20,
    paddingHorizontal: 10,
    paddingVertical: 8,
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
});

export default CityFilterInput;