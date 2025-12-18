import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

export interface Category {
  id: string;
  label: string;
  icon: any;
}

interface CategorySelectorProps {
  categories: Category[];
  selected: string;
  onSelect: (categoryId: string) => void;
}

const CategorySelector: React.FC<CategorySelectorProps> = ({
  categories,
  selected,
  onSelect,
}) => {
  const colors = useThemedColors();
  
  return (
    <View>
      <Text style={[styles.sectionLabel, { color: colors.black }]}>
        Выберите категорию *
      </Text>
      <View style={styles.categoriesContainer}>
        {categories.map((category) => {
          const isSelected = selected === category.id;
          
          return (
            <TouchableOpacity
              key={category.id}
              style={[
                styles.categoryButton,
                { backgroundColor: colors.veryLightGrey },
                isSelected && { backgroundColor: colors.lightBlue }
              ]}
              onPress={() => {
                onSelect(category.id);
              }}
            >
              <Image source={category.icon} style={[styles.categoryIcon, {tintColor: colors.black}]} />
            </TouchableOpacity>
          );
        })}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  sectionLabel: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    marginBottom: 10,
  },
  categoriesContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 8,
    paddingHorizontal: 10,
    marginHorizontal: 20,
    marginBottom: 20,
  },
  categoryButton: {
    width: 40,
    height: 40,
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  categoryIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
  },
});

export default CategorySelector;