import barsStyle from '@/app/styles/barsStyle';
import CityFilterInput from '@/components/filters/CityFilterInput';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { getCategoryDisplayName } from '@/utils/categoryMapping';
import React, { useRef, useState } from 'react';
import { Image, Modal, Pressable, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { SortingActions, SortingState } from '../../hooks/useSearchState';

interface CategorySearchBarProps {
  sortingState: SortingState;
  sortingActions: SortingActions;
  showCategoryFilter?: boolean;
}

const CategorySearchBar: React.FC<CategorySearchBarProps> = ({ 
  sortingState, 
  sortingActions,
  showCategoryFilter = false 
}) => {
  const insets = useSafeAreaInsets();
  const [filterModalVisible, setFilterModalVisible] = useState(false);
  const [inputText, setInputText] = useState('');
  const [filterModalPos, setFilterModalPos] = useState({ top: 0, left: 0 });
  
  const filterIconRef = useRef<View>(null);
  const [filterIconLayout, setFilterIconLayout] = useState({ width: 0, height: 0 });

  const { checkedCategories, sortByDate, sortByParticipants, searchQuery, cityFilter } = sortingState;
  const { 
    setSortByDate, 
    setSortByParticipants, 
    setSearchQuery,
    toggleCategoryCheckbox,
    setCityFilter
  } = sortingActions;

  const handleSearchInputChange = (text: string) => {
    setInputText(text);
    if (text.trim() === '' && searchQuery.trim() !== '') {
      setSearchQuery('');
    }
  };

  const handleSearchSubmit = () => {
    setSearchQuery(inputText);
  };

  const getSortLabel = (order: string) => {
    switch (order) {
      case 'asc':
        return 'По возрастанию';
      case 'desc':
        return 'По убыванию';
      case 'none':
        return 'Отсутствует';
      default:
        return '';
    }
  };

  return (
    <>
      <View style={styles.container}>
        <TextInput
          placeholder="Поиск в категории..."
          value={inputText}
          onChangeText={handleSearchInputChange}
          onSubmitEditing={handleSearchSubmit}
          returnKeyType="search"
          style={styles.input}
        />

        <TouchableOpacity
          ref={filterIconRef}
          onLayout={(e) => {
            const { width, height } = e.nativeEvent.layout;
            setFilterIconLayout({ width, height });
          }}
          onPress={() => {
            filterIconRef.current?.measureInWindow((x, y) => {
              const modalWidth = 230;
              const left = x - (modalWidth - filterIconLayout.width / 2);
              const top = y + filterIconLayout.height / 2; 
              setFilterModalPos({ top, left });
              setFilterModalVisible(true);
            });
          }}
        >
          <Image style={barsStyle.options} source={require('@/assets/images/top_bar/search_bar/options.png')} />
        </TouchableOpacity>
      </View>

      <Modal
        transparent
        animationType="fade"
        visible={filterModalVisible}
        onRequestClose={() => setFilterModalVisible(false)}
      >
        <Pressable
          style={[styles.modalOverlay, { paddingTop: insets.top }]}
          onPress={() => setFilterModalVisible(false)}
        >
          <Pressable
            style={[
              styles.dropdown,
              { position: 'absolute', top: filterModalPos.top, left: filterModalPos.left },
            ]}
            onPress={() => {}}
          >
            <CityFilterInput
              value={cityFilter}
              onChangeText={setCityFilter}
            />

            {showCategoryFilter && (
              <>
                <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>
                  Фильтрация по категориям
                </Text>
                {['Все', 'Игры', 'Фильмы', 'Настолки', 'Другое'].map((cat) => (
                  <TouchableOpacity
                    key={cat}
                    style={styles.radioItem}
                    onPress={() => toggleCategoryCheckbox(cat)}
                  >
                    <View style={styles.radioCircleEmpty}>
                      {checkedCategories.includes(cat) && <View style={styles.radioInnerCircle} />}
                    </View>
                    <Text style={styles.radioLabel}>{getCategoryDisplayName(cat)}</Text>
                  </TouchableOpacity>
                ))}
              </>
            )}

            <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>
              Сортировка по дате
            </Text>
            {['none', 'asc', 'desc'].map((order) => (
              <TouchableOpacity
                key={order}
                style={styles.radioItem}
                onPress={() => setSortByDate(order as 'asc' | 'desc' | 'none')}
              >
                <View style={styles.radioCircleEmpty}>
                  {sortByDate === order && <View style={styles.radioInnerCircle} />}
                </View>
                <Text style={styles.radioLabel}>
                  {getSortLabel(order)}
                </Text>
              </TouchableOpacity>
            ))}

            <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>Сортировка по участникам</Text>
            {['none', 'asc', 'desc'].map((order) => (
              <TouchableOpacity
                key={order}
                style={styles.radioItem}
                onPress={() => setSortByParticipants(order as 'asc' | 'desc' | 'none')}
              >
                <View style={styles.radioCircleEmpty}>
                  {sortByParticipants === order && <View style={styles.radioInnerCircle} />}
                </View>
                <Text style={styles.radioLabel}>
                  {getSortLabel(order)}
                </Text>
              </TouchableOpacity>
            ))}
          </Pressable>
        </Pressable>
      </Modal>
    </>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    borderRadius: 28,
    paddingHorizontal: 10,
    margin: 4,
    backgroundColor: Colors.veryLightGrey
  },
  input: {
    flex: 1,
    paddingVertical: 8,
    paddingHorizontal: 10,
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
  },
  modalOverlay: {
    flex: 1,
    justifyContent: 'flex-start',
    alignItems: 'flex-start',
    paddingRight: 80,
  },
  dropdown: {
    width: 230,
    backgroundColor: Colors.white,
    borderRadius: 10,
    borderTopEndRadius: 0,
    padding: 12,
    elevation: 5,
  },
  dropdownTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    marginBottom: 8,
  },
  radioItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 4,
  },
  radioLabel: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  radioCircleEmpty: {
    height: 20,
    width: 20,
    borderRadius: 10,
    borderWidth: 2,
    borderColor: Colors.lightBlue2,
    backgroundColor: 'transparent',
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 8,
  },
  radioInnerCircle: {
    width: 12,
    height: 12,
    borderRadius: 10,
    backgroundColor: Colors.blue2,
    alignSelf: 'center',
  },
});

export default CategorySearchBar;