import { Montserrat } from '@/constants/Montserrat';
import { SortingActions, SortingState } from '@/hooks/useSearchState';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useRef, useState } from 'react';
import { Image, StyleSheet, TextInput, TouchableOpacity, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import FilterModal from '../search/modal/FilterModal';

interface GroupEventsSearchBarProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  sortingState: SortingState;
  sortingActions: SortingActions;
}

const GroupEventsSearchBar: React.FC<GroupEventsSearchBarProps> = ({
  searchQuery,
  onSearchChange,
  sortingState,
  sortingActions,
}) => {
  const colors = useThemedColors();
  const insets = useSafeAreaInsets();
  const [filterModalVisible, setFilterModalVisible] = useState(false);
  const [filterModalPos, setFilterModalPos] = useState({ top: 0, left: 0 });
  
  const filterIconRef = useRef<View>(null);
  const [filterIconLayout, setFilterIconLayout] = useState({ width: 0, height: 0 });

  return (
    <>
      <View style={[styles.container, { backgroundColor: colors.veryLightGrey }]}>
        <TextInput
          placeholder="Найти событие..."
          value={searchQuery}
          onChangeText={onSearchChange}
          returnKeyType="search"
          style={[styles.input, { color: colors.black }]}
          placeholderTextColor={colors.grey}
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
            style={styles.optionsIcon} 
            source={require('@/assets/images/top_bar/search_bar/options.png')} 
          />
        </TouchableOpacity>
      </View>

      <FilterModal
        visible={filterModalVisible}
        onClose={() => setFilterModalVisible(false)}
        position={filterModalPos}
        searchType="event"
        eventFilters={{
          sortingState,
          sortingActions,
        }}
        groupFilters={{
          groupSearchState: {
            searchQuery: '',
            sortByParticipants: 'none',
            sortByRegistration: 'none',
            checkedCategories: [],
          },
          groupSearchActions: {
            setSearchQuery: () => {},
            setSortByParticipants: () => {},
            setSortByRegistration: () => {},
            toggleCategoryCheckbox: () => {},
            setCheckedCategories: () => {},
            resetFilters: () => {},
          },
        }}
        topInset={insets.top}
      />
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
    height: 40,
  },
  input: {
    flex: 1,
    paddingVertical: 8,
    paddingHorizontal: 10,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  optionsIcon: {
    width: 25,
    height: 25,
    resizeMode: 'contain',
  },
});

export default GroupEventsSearchBar;