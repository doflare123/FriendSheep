import barsStyle from '@/app/styles/barsStyle';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React, { useRef, useState } from 'react';
import { Image, Modal, Pressable, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

interface GroupEventsSearchBarProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  checkedCategories: string[];
  onCategoryToggle: (category: string) => void;
  sortByDate: 'asc' | 'desc' | 'none';
  onDateSortChange: (order: 'asc' | 'desc' | 'none') => void;
  sortByParticipants: 'asc' | 'desc' | 'none';
  onParticipantsSortChange: (order: 'asc' | 'desc' | 'none') => void;
}

const GroupEventsSearchBar: React.FC<GroupEventsSearchBarProps> = ({
  searchQuery,
  onSearchChange,
  checkedCategories,
  onCategoryToggle,
  sortByDate,
  onDateSortChange,
  sortByParticipants,
  onParticipantsSortChange,
}) => {
  const insets = useSafeAreaInsets();
  const [filterModalVisible, setFilterModalVisible] = useState(false);
  const [filterModalPos, setFilterModalPos] = useState({ top: 0, left: 0 });
  
  const filterIconRef = useRef<View>(null);
  const [filterIconLayout, setFilterIconLayout] = useState({ width: 0, height: 0 });

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

  return (
    <>
      <View style={styles.container}>

        <TextInput
          placeholder="Найти событие..."
          value={searchQuery}
          onChangeText={onSearchChange}
          returnKeyType="search"
          style={styles.input}
          placeholderTextColor={Colors.grey}
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
          <Image 
            style={barsStyle.options} 
            source={require('@/assets/images/top_bar/search_bar/options.png')} 
          />
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
            <Text style={styles.dropdownTitle}>Фильтрация по категориям</Text>
            {['Все', 'Игры', 'Фильмы', 'Настолки', 'Другое'].map((cat) => (
              <TouchableOpacity
                key={cat}
                style={styles.radioItem}
                onPress={() => onCategoryToggle(cat)}
              >
                <View style={styles.radioCircleEmpty}>
                  {checkedCategories.includes(cat) && <View style={styles.radioInnerCircle} />}
                </View>
                <Text style={styles.radioLabel}>{cat}</Text>
              </TouchableOpacity>
            ))}

            <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>Сортировка по дате</Text>
            {(['none', 'asc', 'desc'] as const).map((order) => (
              <TouchableOpacity
                key={order}
                style={styles.radioItem}
                onPress={() => onDateSortChange(order)}
              >
                <View style={styles.radioCircleEmpty}>
                  {sortByDate === order && <View style={styles.radioInnerCircle} />}
                </View>
                <Text style={styles.radioLabel}>
                  {getSortingLabel(order)}
                </Text>
              </TouchableOpacity>
            ))}

            <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>Сортировка по участникам</Text>
            {(['none', 'asc', 'desc'] as const).map((order) => (
              <TouchableOpacity
                key={order}
                style={styles.radioItem}
                onPress={() => onParticipantsSortChange(order)}
              >
                <View style={styles.radioCircleEmpty}>
                  {sortByParticipants === order && <View style={styles.radioInnerCircle} />}
                </View>
                <Text style={styles.radioLabel}>
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
    backgroundColor: Colors.veryLightGrey,
    height: 40,
  },
  iconContainer: {
    paddingRight: 8,
  },
  searchIcon: {
    width: 25,
    height: 25,
    resizeMode: 'contain',
  },
  input: {
    flex: 1,
    paddingVertical: 8,
    paddingHorizontal: 10,
    fontFamily: inter.regular,
    fontSize: 16,
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
    borderColor: Colors.blue,
    borderWidth: 1,
  },
  dropdownTitle: {
    fontFamily: inter.bold,
    marginBottom: 8,
  },
  radioItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 6,
  },
  radioLabel: {
    fontFamily: inter.regular,
    fontSize: 14,
  },
  radioCircleEmpty: {
    height: 20,
    width: 20,
    borderRadius: 10,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
    backgroundColor: 'transparent',
    marginRight: 8,
  },
  radioInnerCircle: {
    width: 12,
    height: 12,
    borderRadius: 10,
    backgroundColor: Colors.blue,
    alignSelf: 'center',
    marginTop: 2,
  },
});

export default GroupEventsSearchBar;