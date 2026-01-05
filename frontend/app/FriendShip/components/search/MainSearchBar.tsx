import barsStyle from '@/app/styles/barsStyle';
import FilterModal from '@/components/search/modal/FilterModal';
import SearchTypeModal from '@/components/search/modal/SearchTypeModal';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useGroupSearchState } from '@/hooks/useGroupSearchState';
import { SortingActions, SortingState } from '@/hooks/useSearchState';
import { useThemedColors } from '@/hooks/useThemedColors';
import { useUserSearchState } from '@/hooks/useUserSearchState';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import React, { useRef, useState } from 'react';
import { Image, StyleSheet, TextInput, TouchableOpacity, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

interface MainSearchBarProps {
  sortingState: SortingState;
  sortingActions: SortingActions;
}

type MainSearchBarNavigationProp = NativeStackNavigationProp<
  RootStackParamList,
  'MainPage'
>;

type MainSearchBarRouteProp = RouteProp<RootStackParamList, 'MainPage'>;

const MainSearchBar: React.FC<MainSearchBarProps> = ({ sortingState, sortingActions }) => {
  const colors = useThemedColors();
  const navigation = useNavigation<MainSearchBarNavigationProp>();
  const route = useRoute<MainSearchBarRouteProp>();

  const { searchState: groupSearchState, searchActions: groupSearchActions } = useGroupSearchState();
  const { searchState: userSearchState, searchActions: userSearchActions } = useUserSearchState();

  const insets = useSafeAreaInsets();
  const [categoryModalVisible, setCategoryModalVisible] = useState(false);
  const [filterModalVisible, setFilterModalVisible] = useState(false);

  const [categoryModalPos, setCategoryModalPos] = useState({ top: 0, left: 0 });
  const [filterModalPos, setFilterModalPos] = useState({ top: 0, left: 0 });

  const categoryIconRef = useRef<View>(null);
  const filterIconRef = useRef<View>(null);

  const [filterIconLayout, setFilterIconLayout] = useState({ width: 0, height: 0 });

  const { searchQuery, activeSearchType } = sortingState;
  const { setSearchQuery, setActiveSearchType } = sortingActions;

  const getCurrentSearchQuery = () => {
    switch (activeSearchType) {
      case 'group':
        return groupSearchState.searchQuery;
      case 'profile':
        return userSearchState.searchQuery;
      case 'event':
      default:
        return searchQuery;
    }
  };

  const handleSearchTypeChange = (type: 'event' | 'profile' | 'group') => {
    setActiveSearchType(type);
    setCategoryModalVisible(false);
  };

  const getPlaceholderByType = (type: 'event' | 'profile' | 'group') => {
    switch (type) {
      case 'event':
        return 'Найти событие...';
      case 'profile':
        return 'Найти пользователя...';
      case 'group':
        return 'Найти группу...';
      default:
        return 'Поиск';
    }
  };

  const handleSearchInputChange = (text: string) => {
    switch (activeSearchType) {
      case 'group':
        groupSearchActions.setSearchQuery(text);
        break;
      case 'profile':
        userSearchActions.setSearchQuery(text);
        break;
      case 'event':
      default:
        setSearchQuery(text);
        break;
    }
  };

  const handleSearchSubmit = () => {
    if (activeSearchType === 'event') {
      if (navigation.getState().routes[navigation.getState().index].name !== 'MainPage') {
        navigation.navigate('MainPage', { searchQuery: searchQuery });
      }
    } else if (activeSearchType === 'group') {
      navigation.navigate('GroupSearchPage');
    } else if (activeSearchType === 'profile') {
      navigation.navigate('UserSearchPage');
    }
  };

  const getSearchTypeIcon = () => {
    switch (activeSearchType) {
      case 'event':
        return require('@/assets/images/top_bar/search_bar/event-arrow.png');
      case 'profile':
        return require('@/assets/images/top_bar/search_bar/profile-arrow.png');
      case 'group':
        return require('@/assets/images/top_bar/search_bar/group-arrow.png');
    }
  };

  return (
    <>
      <View style={[styles.container, { backgroundColor: colors.veryLightGrey }]}>
        <TouchableOpacity
          ref={categoryIconRef}
          onPress={() => {
            categoryIconRef.current?.measureInWindow((x, y) => {
              setCategoryModalPos({ top: y - 15, left: x - 10 });
              setCategoryModalVisible(true);
            });
          }}
        >
          <Image style={[barsStyle.switch, {tintColor: colors.lightGreyBlue}]} source={getSearchTypeIcon()} />
        </TouchableOpacity>

        <TextInput
          placeholder={getPlaceholderByType(activeSearchType)}
          placeholderTextColor={Colors.lightGrey}
          value={getCurrentSearchQuery()}
          onChangeText={handleSearchInputChange}
          onSubmitEditing={handleSearchSubmit}
          returnKeyType="search"
          style={[styles.input, { color: colors.black }]}
        />

        {activeSearchType !== 'profile' && (
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
            style={[
              {
                opacity:
                  activeSearchType === 'event' || activeSearchType === 'group' ? 1 : 0.5,
              },
            ]}
            disabled={activeSearchType !== 'event' && activeSearchType !== 'group'}
          >
            <Image
              style={[barsStyle.options, {tintColor: colors.lightGreyBlue}]}
              source={require('@/assets/images/top_bar/search_bar/options.png')}
            />
          </TouchableOpacity>
        )}
      </View>

      <SearchTypeModal
        visible={categoryModalVisible}
        onClose={() => setCategoryModalVisible(false)}
        position={categoryModalPos}
        activeType={activeSearchType}
        onTypeChange={handleSearchTypeChange}
      />

      <FilterModal
        visible={filterModalVisible}
        onClose={() => setFilterModalVisible(false)}
        position={filterModalPos}
        searchType={activeSearchType}
        eventFilters={{ sortingState, sortingActions }}
        groupFilters={{ groupSearchState, groupSearchActions }}
        topInset={insets.top}
      />
    </>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
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
});

export default MainSearchBar;