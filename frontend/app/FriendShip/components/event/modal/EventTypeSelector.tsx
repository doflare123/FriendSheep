import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
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
  const colors = useThemedColors();

  return (
    <View>
      <Text style={[styles.sectionLabel, { color: colors.black }]}>
        Выберите тип события *
      </Text>
      <View style={styles.typeContainer}>
        <View style={{ flexDirection: 'column', alignItems: 'center' }}>
          <TouchableOpacity
            style={[
              styles.typeButton,
              { borderColor: colors.lightGrey },
              selected === 'offline' && { borderColor: colors.lightBlue }
            ]}
            onPress={() => onSelect('offline')}
          >
            <View style={[styles.typeIconContainer, { backgroundColor: colors.white }]}>
              <Image
                source={require('@/assets/images/event_card/offline.png')}
                style={[styles.typeIcon, {tintColor: colors.black}]}
              />
            </View>
          </TouchableOpacity>
          <Text style={[styles.typeText, { color: colors.black }]}>
            Оффлайн
          </Text>
        </View>

        <View style={{ flexDirection: 'column', alignItems: 'center' }}>
          <TouchableOpacity
            style={[
              styles.typeButton,
              { borderColor: colors.lightGrey },
              selected === 'online' && { borderColor: colors.lightBlue }
            ]}
            onPress={() => onSelect('online')}
          >
            <View style={[styles.typeIconContainer, { backgroundColor: colors.white }]}>
              <Image
                source={require('@/assets/images/event_card/online.png')}
                style={[styles.typeIcon, {tintColor: colors.black}]}
              />
            </View>
          </TouchableOpacity>
          <Text style={[styles.typeText, { color: colors.black }]}>
            Онлайн
          </Text>
        </View>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  sectionLabel: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
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
  },
  typeIconContainer: {
    width: 30,
    height: 30,
    borderRadius: 20,
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
    fontFamily: Montserrat.bold,
    fontSize: 10,
  },
});

export default EventTypeSelector;