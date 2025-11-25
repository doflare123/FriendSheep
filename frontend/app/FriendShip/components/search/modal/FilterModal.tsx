import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { GroupSearchActions, GroupSearchState } from '@/hooks/useGroupSearchState';
import { SortingActions, SortingState } from '@/hooks/useSearchState';
import React from 'react';
import { Modal, Pressable, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

interface FilterModalProps {
  visible: boolean;
  onClose: () => void;
  position: { top: number; left: number };
  searchType: 'event' | 'profile' | 'group';
  eventFilters: {
    sortingState: SortingState;
    sortingActions: SortingActions;
  };
  groupFilters: {
    groupSearchState: GroupSearchState;
    groupSearchActions: GroupSearchActions;
  };
  topInset: number;
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

const FilterModal: React.FC<FilterModalProps> = ({
  visible,
  onClose,
  position,
  searchType,
  eventFilters,
  groupFilters,
  topInset,
}) => {
  const { sortingState, sortingActions } = eventFilters;
  const { groupSearchState, groupSearchActions } = groupFilters;

  return (
    <View style={styles.shadowWrapper}>
      <Modal
        transparent
        animationType="fade"
        visible={visible}
        onRequestClose={onClose}
      >
        <Pressable
          style={[styles.modalOverlay, { paddingTop: topInset }]}
          onPress={onClose}
        >
          <Pressable
            style={[
              styles.dropdown,
              { position: 'absolute', top: position.top, left: position.left },
            ]}
            onPress={() => {}}
          >
            {searchType === 'event' && (
              <>
                <Text style={styles.dropdownTitle}>Фильтрация по категориям</Text>
                {['Все', 'Игры', 'Фильмы', 'Настолки', 'Другое'].map((cat) => (
                  <TouchableOpacity
                    key={cat}
                    style={styles.radioItem}
                    onPress={() => sortingActions.toggleCategoryCheckbox(cat)}
                  >
                    <View style={styles.radioCircleEmpty}>
                      {sortingState.checkedCategories.includes(cat) && (
                        <View style={styles.radioInnerCircle} />
                      )}
                    </View>
                    <Text style={styles.radioLabel}>{cat}</Text>
                  </TouchableOpacity>
                ))}

                <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>
                  Сортировка по дате
                </Text>
                {(['none', 'asc', 'desc'] as const).map((order) => (
                  <TouchableOpacity
                    key={order}
                    style={styles.radioItem}
                    onPress={() => sortingActions.setSortByDate(order)}
                  >
                    <View style={styles.radioCircleEmpty}>
                      {sortingState.sortByDate === order && (
                        <View style={styles.radioInnerCircle} />
                      )}
                    </View>
                    <Text style={styles.radioLabel}>{getSortingLabel(order)}</Text>
                  </TouchableOpacity>
                ))}

                <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>
                  Сортировка по участникам
                </Text>
                {(['none', 'asc', 'desc'] as const).map((order) => (
                  <TouchableOpacity
                    key={order}
                    style={styles.radioItem}
                    onPress={() => sortingActions.setSortByParticipants(order)}
                  >
                    <View style={styles.radioCircleEmpty}>
                      {sortingState.sortByParticipants === order && (
                        <View style={styles.radioInnerCircle} />
                      )}
                    </View>
                    <Text style={styles.radioLabel}>{getSortingLabel(order)}</Text>
                  </TouchableOpacity>
                ))}
              </>
            )}

          {searchType === 'group' && (
            <>
              <Text style={styles.dropdownTitle}>Сортировка по участникам</Text>
              {(['none', 'asc', 'desc'] as const).map((order) => (
                <TouchableOpacity
                  key={order}
                  style={styles.radioItem}
                  onPress={() => groupSearchActions.setSortByParticipants(order)}
                >
                  <View style={styles.radioCircleEmpty}>
                    {groupSearchState.sortByParticipants === order && (
                      <View style={styles.radioInnerCircle} />
                    )}
                  </View>
                  <Text style={styles.radioLabel}>{getSortingLabel(order)}</Text>
                </TouchableOpacity>
              ))}

              <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>
                Сортировка по категориям
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
                      groupSearchActions.setCheckedCategories([]);
                    } else {
                      groupSearchActions.toggleCategoryCheckbox(value);
                    }
                  }}
                >
                  <View style={styles.radioCircleEmpty}>
                    {value === null
                      ? groupSearchState.checkedCategories.length === 0 && (
                          <View style={styles.radioInnerCircle} />
                        )
                      : groupSearchState.checkedCategories.includes(value) && (
                          <View style={styles.radioInnerCircle} />
                        )}
                  </View>
                  <Text style={styles.radioLabel}>{label}</Text>
                </TouchableOpacity>
              ))}

              <Text style={[styles.dropdownTitle, { marginTop: 12 }]}>
                Сортировка по регистрации
              </Text>
              {(['none', 'asc', 'desc'] as const).map((order) => (
                <TouchableOpacity
                  key={order}
                  style={styles.radioItem}
                  onPress={() => groupSearchActions.setSortByRegistration(order)}
                >
                  <View style={styles.radioCircleEmpty}>
                    {groupSearchState.sortByRegistration === order && (
                      <View style={styles.radioInnerCircle} />
                    )}
                  </View>
                  <Text style={styles.radioLabel}>{getSortingLabel(order)}</Text>
                </TouchableOpacity>
              ))}
            </>
          )}
          </Pressable>
        </Pressable>
      </Modal>
    </View>
  );
};

const styles = StyleSheet.create({
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
    marginRight: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  radioInnerCircle: {
    width: 12,
    height: 12,
    borderRadius: 10,
    backgroundColor: Colors.blue2,
    alignSelf: 'center',
    justifyContent: 'center',
  },
  shadowWrapper:{
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 6,
  },
});

export default FilterModal;