import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { StyleSheet, Text, TextInput } from 'react-native';

interface CityFilterInputProps {
  value: string;
  onChangeText: (text: string) => void;
}

const CityFilterInput: React.FC<CityFilterInputProps> = ({ value, onChangeText }) => {
  return (
    <>
      <Text style={styles.dropdownTitle}>Фильтрация по городу</Text>
      <TextInput
        style={styles.cityInput}
        placeholder="Введите город..."
        value={value}
        onChangeText={onChangeText}
        placeholderTextColor={Colors.grey}
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
    color: Colors.black,
    backgroundColor: Colors.veryLightGrey,
  },
});

export default CityFilterInput;