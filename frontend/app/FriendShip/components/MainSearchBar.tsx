import barsStyle from '@/app/styles/barsStyle';
import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React, { useState } from 'react';
import { Image, Modal, Pressable, StyleSheet, Text, TextInput, TouchableOpacity, View } from 'react-native';

const MainSearchBar = () => {
  const [search, setSearch] = useState('');
  const [filterVisible, setFilterVisible] = useState(false);

  return (
    <>
      <View style={styles.container}>

        <Image style={barsStyle.switch} source={require("../assets/images/top_bar/search_bar/event-arrow.png")}/>

        <TextInput
          placeholder="Найти событие"
          value={search}
          onChangeText={setSearch}
          style={styles.input}
        />

        <TouchableOpacity onPress={() => setFilterVisible(true)}>
          <Image style={barsStyle.options} source={require("../assets/images/top_bar/search_bar/options.png")} />
        </TouchableOpacity>
      </View>

      <Modal
        transparent
        animationType="fade"
        visible={filterVisible}
        onRequestClose={() => setFilterVisible(false)}
      >
        <Pressable style={styles.modalOverlay} onPress={() => setFilterVisible(false)}>
          <View style={styles.dropdown}>
            <Text style={styles.dropdownTitle}>Сортировка по категориям</Text>
            <TouchableOpacity style={styles.radioItem}>
              <View style={styles.radioCircleSelected} />
              <Text style={styles.radioLabel}>Все</Text>
            </TouchableOpacity>
          </View>
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
    paddingRight: 35,
    backgroundColor: Colors.lightBlack,
    fontFamily: inter.regular
  },
  dropdown: {
    width: 230,
    backgroundColor: '#fff',
    borderRadius: 10,
    padding: 12,
    elevation: 5,
    borderColor: '#1d74f5',
    borderWidth: 1,
  },
  dropdownTitle: {
    fontWeight: 'bold',
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
    borderColor: '#1d74f5',
    backgroundColor: '#1d74f5',
    marginRight: 8,
  },
  radioLabel: {
    fontSize: 14,
  },
});

export default MainSearchBar;
