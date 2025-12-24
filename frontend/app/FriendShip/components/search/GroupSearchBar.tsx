import { Montserrat } from '@/constants/Montserrat';
import { GroupSearchActions, GroupSearchState } from '@/hooks/useGroupSearchState';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useRef, useState } from 'react';
import { Image, Modal, Pressable, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

interface GroupSearchBarProps {
  searchState: GroupSearchState;
  searchActions: GroupSearchActions;
}

const getSortingLabel = (order: 'asc' | 'desc' | 'none') => {
  switch (order) {
    case 'asc':
      return 'По возрастанию';
    case 'desc':
      return 'По убыванию';
    case 'none':
      return 'Отсутствует';
    default:
      return 'Отсутствует';
  }
};

const GroupSearchBar: React.FC<GroupSearchBarProps> = ({ 
  searchState, 
  searchActions
}) => {
  const colors = useThemedColors();
  const insets = useSafeAreaInsets();
  const [filterModalVisible, setFilterModalVisible] = useState(false);
  const [inputText, setInputText] = useState(searchState.searchQuery);
  const [filterModalPos, setFilterModalPos] = useState({ top: 0, left: 0 });
  
  const filterIconRef = useRef<View>(null);
  const [filterIconLayout, setFilterIconLayout] = useState({ width: 0, height: 0 });

  const { 
    sortByParticipants, 
    sortByRegistration,
    checkedCategories 
  } = searchState;
  
  const { 
    setSearchQuery,
    setSortByParticipants, 
    setSortByRegistration,
    setCheckedCategories,
    toggleCategoryCheckbox
  } = searchActions;

  const handleSearchInputChange = (text: string) => {
    setInputText(text);
    if (text.trim() === '' && searchState.searchQuery.trim() !== '') {
      setSearchQuery('');
    }
  };

  const handleSearchSubmit = () => {
    setSearchQuery(inputText);
  };

  return (
    <>
      <View style={[styles.container, { backgroundColor: colors.veryLightGrey }]}>
        <TextInput
          placeholder="Поиск групп..."
          placeholderTextColor={colors.grey}
          value={inputText}
          onChangeText={handleSearchInputChange}
          onSubmitEditing={handleSearchSubmit}
          returnKeyType="search"
          style={[styles.input, { color: colors.black }]}
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
          <Image style={[styles.optionsIcon, {tintColor: colors.lightGreyBlue}]} source={require('@/assets/images/top_bar/search_bar/options.png')} />
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
              { 
                position: 'absolute', 
                top: filterModalPos.top, 
                left: filterModalPos.left,
                backgroundColor: colors.white
              },
            ]}
            onPress={() => {}}
          >
            <Text style={[styles.dropdownTitle, { color: colors.black }]}>
              Сортировка по участникам
            </Text>
            {(['none', 'asc', 'desc'] as const).map((order) => (
              <TouchableOpacity
                key={order}
                style={styles.radioItem}
                onPress={() => setSortByParticipants(order)}
              >
                <View style={[styles.radioCircleEmpty, { borderColor: colors.lightBlue2 }]}>
                  {sortByParticipants === order && (
                    <View style={[styles.radioInnerCircle, { backgroundColor: colors.blue2 }]} />
                  )}
                </View>
                <Text style={[styles.radioLabel, { color: colors.black }]}>
                  {getSortingLabel(order)}
                </Text>
              </TouchableOpacity>
            ))}

            <Text style={[styles.dropdownTitle, { marginTop: 12, color: colors.black }]}>
              Фильтрация по категориям
            </Text>
            {[
              { label: 'Все', value: null },
              { label: 'Фильмы', value: 'movie' as const },
              { label: 'Игры', value: 'game' as const },
              { label: 'Настолки', value: 'table_game' as const },
              { label: 'Другое', value: 'other' as const },
            ].map(({ label, value }) => (
              <TouchableOpacity
                key={label}
                style={styles.radioItem}
                onPress={() => {
                  if (value === null) {
                    setCheckedCategories([]);
                  } else {
                    toggleCategoryCheckbox(value);
                  }
                }}
              >
                <View style={[styles.radioSquareEmpty, { borderColor: colors.lightBlue2 }]}>
                  {value === null
                    ? checkedCategories.length === 0 && (
                        <View style={[styles.radioInnerSquare, { backgroundColor: colors.blue2 }]} />
                      )
                    : checkedCategories.includes(value) && (
                        <View style={[styles.radioInnerSquare, { backgroundColor: colors.blue2 }]} />
                      )}
                </View>
                <Text style={[styles.radioLabel, { color: colors.black }]}>
                  {label}
                </Text>
              </TouchableOpacity>
            ))}

            <Text style={[styles.dropdownTitle, { marginTop: 12, color: colors.black }]}>
              Сортировка по регистрации
            </Text>
            {(['none', 'asc', 'desc'] as const).map((order) => (
              <TouchableOpacity
                key={order}
                style={styles.radioItem}
                onPress={() => setSortByRegistration(order)}
              >
                <View style={[styles.radioCircleEmpty, { borderColor: colors.lightBlue2 }]}>
                  {sortByRegistration === order && (
                    <View style={[styles.radioInnerCircle, { backgroundColor: colors.blue2 }]} />
                  )}
                </View>
                <Text style={[styles.radioLabel, { color: colors.black }]}>
                  {getSortingLabel(order)}
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
  },
  input: {
    flex: 1,
    paddingVertical: 8,
    paddingHorizontal: 10,
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  optionsIcon: {
    width: 24,
    height: 24,
    resizeMode: 'contain',
  },
  modalOverlay: {
    flex: 1,
    justifyContent: 'flex-start',
    alignItems: 'flex-start',
    paddingRight: 80,
  },
  dropdown: {
    width: 230,
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
    backgroundColor: 'transparent',
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 8,
  },
  radioInnerCircle: {
    width: 12,
    height: 12,
    borderRadius: 10,
    alignSelf: 'center',
  },
  radioSquareEmpty: {
    height: 20,
    width: 20,
    borderWidth: 2,
    borderRadius: 6,
    backgroundColor: 'transparent',
    marginRight: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  radioInnerSquare: {
    width: 12,
    height: 12,
    borderRadius: 2,
    alignSelf: 'center',
    justifyContent: 'center',
  },
});

export default GroupSearchBar;