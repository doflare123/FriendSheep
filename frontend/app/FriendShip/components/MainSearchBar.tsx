import barsStyle from '@/app/styles/barsStyle';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React, { useState } from 'react';
import { Image, Modal, Pressable, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';

const MainSearchBar = () => {
  const [search, setSearch] = useState('');
  const [filterVisible, setFilterVisible] = useState(false);

  const [categoryModalVisible, setCategoryModalVisible] = useState(false);
  const [filterModalVisible, setFilterModalVisible] = useState(false);

  const [activeSearchType, setActiveSearchType] = useState<'event' | 'profile' | 'group'>('event');
  const [selectedCategory, setSelectedCategory] = useState<'Все' | 'Игры' | 'Фильмы' | 'Другое'>('Все');
  const handleSearchTypeChange = (type: 'event' | 'profile' | 'group') => {
    setActiveSearchType(type);
    setCategoryModalVisible(false);
  };

  const [checkedCategories, setCheckedCategories] = useState<string[]>(['Все']);
  const [sortByDate, setSortByDate] = useState<'asc' | 'desc'>('asc');
  const [sortByParticipants, setSortByParticipants] = useState<'asc' | 'desc'>('asc');

  const [sortByRegistration, setSortByRegistration] = useState<'asc' | 'desc'>('asc');


  const toggleCategoryCheckbox = (category: string) => {
    if (category === 'Все') {
      setCheckedCategories(['Все']);
    } else {
      const filtered = checkedCategories.filter(c => c !== 'Все');
      if (checkedCategories.includes(category)) {
        setCheckedCategories(filtered.filter(c => c !== category));
      } else {
        setCheckedCategories([...filtered, category]);
      }
    }
  };

  const getPlaceholderByType = (type: 'event' | 'profile' | 'group') => {
    switch (type) {
      case 'event':
        return 'Найти событие...';
      case 'profile':
        return 'Найти профиль...';
      case 'group':
        return 'Найти группу...';
      default:
        return 'Поиск';
    }
  };

  return (
    <>
      <View style={styles.container}>

        <TouchableOpacity onPress={() => setCategoryModalVisible(true)}>
          <Image
            style={barsStyle.switch}
            source={
              activeSearchType === 'event'
                ? require('../assets/images/top_bar/search_bar/event-arrow.png')
                : activeSearchType === 'profile'
                  ? require('../assets/images/top_bar/search_bar/profile-arrow.png')
                  : require('../assets/images/top_bar/search_bar/group-arrow.png')
            }
          />
        </TouchableOpacity>

        <TextInput
          placeholder={getPlaceholderByType(activeSearchType)}
          value={search}
          onChangeText={setSearch}
          style={styles.input}
        />

        {activeSearchType !== 'profile' && (
          <TouchableOpacity onPress={() => setFilterModalVisible(true)}>
            <Image style={barsStyle.options} source={require("../assets/images/top_bar/search_bar/options.png")} />
          </TouchableOpacity>
        )}

      </View>

      <Modal
        transparent
        animationType="fade"
        visible={categoryModalVisible}
        onRequestClose={() => setCategoryModalVisible(false)}
      >
        <Pressable style={styles.modalOverlay} onPress={() => setCategoryModalVisible(false)}>
          <View style={styles.categoryModal}>
            <View>
              <Image
                source={
                  activeSearchType === 'event'
                    ? require('../assets/images/top_bar/search_bar/event-arrow.png')
                    : activeSearchType === 'profile'
                      ? require('../assets/images/top_bar/search_bar/profile-arrow.png')
                      : require('../assets/images/top_bar/search_bar/group-arrow.png')
                }
                style={styles.iconOnly}
                resizeMode="contain"
              />
            </View>

            {(['event', 'profile', 'group'] as const)
              .filter(type => type !== activeSearchType)
              .map((type) => {
                let iconSource;
                if (type === 'event') {
                  iconSource = require('../assets/images/top_bar/search_bar/event-bar.png');
                } else if (type === 'profile') {
                  iconSource = require('../assets/images/top_bar/search_bar/profile.png');
                } else {
                  iconSource = require('../assets/images/top_bar/search_bar/group.png');
                }

                return (
                  <TouchableOpacity key={type} onPress={() => handleSearchTypeChange(type)}>
                    <Image source={iconSource} style={styles.iconOnly} resizeMode="contain" />
                  </TouchableOpacity>
                );
              })}
          </View>
        </Pressable>
      </Modal>

      <Modal
        transparent
        animationType="fade"
        visible={filterModalVisible}
        onRequestClose={() => setFilterModalVisible(false)}
      >
       <Pressable style={styles.modalOverlay} onPress={() => setFilterModalVisible(false)}>
          <Pressable style={styles.dropdown} onPress={() => {}}>
              {activeSearchType === 'event' && (
                <>
                  <Text style={styles.dropdownTitle}>Сортировка по категориям</Text>
                  {['Все', 'Игры', 'Фильмы', 'Настолки', 'Другое'].map((cat) => (
                    <TouchableOpacity
                      key={cat}
                      style={styles.radioItem}
                      onPress={() => toggleCategoryCheckbox(cat)}
                    >
                      <View style={styles.radioCircleEmpty}>
                        {checkedCategories.includes(cat) && <View style={styles.radioInnerCircle} />}
                      </View>
                      <Text style={styles.radioLabel}>{cat}</Text>
                    </TouchableOpacity>
                  ))}

                  <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>Сортировка по дате</Text>
                  {['asc', 'desc'].map((order) => (
                    <TouchableOpacity
                      key={order}
                      style={styles.radioItem}
                      onPress={() => setSortByDate(order as 'asc' | 'desc')}
                    >
                      <View style={styles.radioCircleEmpty}>
                        {sortByDate === order && <View style={styles.radioInnerCircle} />}
                      </View>
                      <Text style={styles.radioLabel}>{order === 'asc' ? 'По возрастанию' : 'По убыванию'}</Text>
                    </TouchableOpacity>
                  ))}

                  <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>Сортировка по участникам</Text>
                  {['asc', 'desc'].map((order) => (
                    <TouchableOpacity
                      key={order}
                      style={styles.radioItem}
                      onPress={() => setSortByParticipants(order as 'asc' | 'desc')}
                    >
                      <View style={styles.radioCircleEmpty}>
                        {sortByParticipants === order && <View style={styles.radioInnerCircle}/>}
                      </View>
                      <Text style={styles.radioLabel}>{order === 'asc' ? 'По возрастанию' : 'По убыванию'}</Text>
                    </TouchableOpacity>
                  ))}
                </>
              )}

              {activeSearchType === 'group' && (
                <>
                  <Text style={styles.dropdownTitle}>Сортировка по участникам</Text>
                  {['asc', 'desc'].map((order) => (
                    <TouchableOpacity
                      key={order}
                      style={styles.radioItem}
                      onPress={() => setSortByParticipants(order as 'asc' | 'desc')}
                    >
                      <View style={styles.radioCircleEmpty}>
                        {sortByParticipants === order && <View style={styles.radioInnerCircle} />}
                      </View>
                      <Text style={styles.radioLabel}>{order === 'asc' ? 'По возрастанию' : 'По убыванию'}</Text>
                    </TouchableOpacity>
                  ))}

                  <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>Сортировка по категориям</Text>
                  {['Все', 'Игры', 'Фильмы', 'Настолки', 'Другое'].map((cat) => (
                    <TouchableOpacity
                      key={cat}
                      style={styles.radioItem}
                      onPress={() => toggleCategoryCheckbox(cat)}
                    >
                      <View style={styles.radioCircleEmpty}>
                        {checkedCategories.includes(cat) && <View style={styles.radioInnerCircle} />}
                      </View>
                      <Text style={styles.radioLabel}>{cat}</Text>
                    </TouchableOpacity>
                  ))}

                  <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>Сортировка по регистрации</Text>
                  {['asc', 'desc'].map((order) => (
                    <TouchableOpacity
                      key={order}
                      style={styles.radioItem}
                      onPress={() => setSortByRegistration(order as 'asc' | 'desc')}
                    >
                      <View style={styles.radioCircleEmpty}>
                        {sortByRegistration === order && <View style={styles.radioInnerCircle} />}
                      </View>
                      <Text style={styles.radioLabel}>{order === 'asc' ? 'По возрастанию' : 'По убыванию'}</Text>
                    </TouchableOpacity>
                  ))}
                </>
              )}
          </Pressable>
        </Pressable>
      </Modal>
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
    borderWidth: 1.5,
    borderColor: Colors.blue,
    marginTop: 30
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
    alignItems: 'flex-end',
    paddingTop: 30,
    paddingRight: 80,
    fontFamily: inter.regular
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
  radioCircleSelected: {
    height: 16,
    width: 16,
    borderRadius: 8,
    borderWidth: 2,
    borderColor: Colors.blue,
    backgroundColor: Colors.blue,
    marginRight: 8,
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
  categoryModal: {
    position: 'absolute',
    top: 5,
    left: 21,
    backgroundColor: Colors.white,
    borderRadius: 10,
    padding: 10,
    elevation: 5,
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 2,
  },
  iconOnly: {
    width: 25,
    height: 25,
    marginVertical: 4,
  },
});

export default MainSearchBar;
