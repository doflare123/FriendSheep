import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

interface EventTypeSelectorProps {
  selected: 'online' | 'offline';
  onSelect: (type: 'online' | 'offline') => void;
}

const EventTypeSelector: React.FC<EventTypeSelectorProps> = ({
  selected,
  onSelect,
}) => {
  return (
    <View>
      <Text style={styles.sectionLabel}>Выберите тип события *</Text>
      <View style={styles.typeContainer}>
        <View style={{ flexDirection: 'column', alignItems: 'center' }}>
          <TouchableOpacity
            style={[styles.typeButton, selected === 'offline' && styles.typeSelected]}
            onPress={() => onSelect('offline')}
          >
            <View style={styles.typeIconContainer}>
              <Image
                source={require('@/assets/images/event_card/offline.png')}
                style={styles.typeIcon}
              />
            </View>
          </TouchableOpacity>
          <Text style={styles.typeText}>Оффлайн</Text>
        </View>

        <View style={{ flexDirection: 'column', alignItems: 'center' }}>
          <TouchableOpacity
            style={[styles.typeButton, selected === 'online' && styles.typeSelected]}
            onPress={() => onSelect('online')}
          >
            <View style={styles.typeIconContainer}>
              <Image
                source={require('@/assets/images/event_card/online.png')}
                style={styles.typeIcon}
              />
            </View>
          </TouchableOpacity>
          <Text style={styles.typeText}>Онлайн</Text>
        </View>
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
  typeContainer: {
    paddingHorizontal: 10,
    justifyContent: 'center',
    flexDirection: 'row',
    gap: 16,
    marginBottom: 20,
  },
  typeButton: {
    flex: 1,
    borderRadius: 60,
    padding: 8,
    alignItems: 'center',
    marginHorizontal: 40,
    borderWidth: 3,
    borderColor: Colors.lightGrey,
  },
  typeSelected: {
    borderColor: Colors.lightBlue3,
  },
  typeIconContainer: {
    width: 30,
    height: 30,
    borderRadius: 20,
    backgroundColor: Colors.white,
    justifyContent: 'center',
    alignItems: 'center',
  },
  typeIcon: {
    width: 25,
    height: 25,
    resizeMode: 'contain',
  },
  typeText: {
    marginTop: 2,
    fontFamily: inter.medium,
    fontSize: 10,
    color: Colors.black,
  },
});

export default EventTypeSelector;