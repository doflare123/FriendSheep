import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
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
  console.log('[CategorySelector] 🎨 Текущий selected:', selected);
  
  return (
    <View>
      <Text style={styles.sectionLabel}>Выберите категорию:</Text>
      <View style={styles.categoriesContainer}>
        {categories.map((category) => {
          const isSelected = selected === category.id;
          console.log(`[CategorySelector] Категория ${category.id}: selected=${isSelected}`);
          
          return (
            <TouchableOpacity
              key={category.id}
              style={[
                styles.categoryButton,
                isSelected && styles.categorySelected
              ]}
              onPress={() => {
                console.log('[CategorySelector] 👆 Клик на категорию:', category.id);
                onSelect(category.id);
              }}
            >
              <Image source={category.icon} style={styles.categoryIcon} />
            </TouchableOpacity>
          );
        })}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  sectionLabel: {
    fontFamily: inter.bold,
    fontSize: 16,
    color: Colors.black,
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
    backgroundColor: Colors.lightGrey,
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  categorySelected: {
    backgroundColor: Colors.lightBlue3,
  },
  categoryIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
  },
});

export default CategorySelector;