import { Colors } from '@/constants/Colors';
import React from 'react';
import { Image, StyleSheet, TouchableOpacity } from 'react-native';

interface CreateEventButtonProps {
  onPress: () => void;
}

const CreateEventButton: React.FC<CreateEventButtonProps> = ({ onPress }) => {
  return (
    <TouchableOpacity 
      style={styles.container} 
      onPress={onPress}
      activeOpacity={0.8}
    >
      <Image
        source={require('@/assets/images/bottom_bar/add.png')}
        style={styles.icon}
      />
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    bottom: 110,
    left: 16,
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: Colors.blue2,
    justifyContent: 'center',
    alignItems: 'center',
    elevation: 5,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 3.84,
  },
  icon: {
    width: 40,
    height: 40,
    tintColor: Colors.white,
  },
});

export default CreateEventButton;