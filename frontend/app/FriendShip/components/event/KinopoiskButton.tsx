import { Colors } from '@/constants/Colors';
import React from 'react';
import { ActivityIndicator, Image, StyleSheet, TouchableOpacity } from 'react-native';

interface KinopoiskButtonProps {
  onPress: () => void;
  disabled?: boolean;
  loading?: boolean;
}

const KinopoiskButton: React.FC<KinopoiskButtonProps> = ({
  onPress,
  disabled = false,
  loading = false,
}) => {
  return (
    <TouchableOpacity
      style={[
        styles.button,
        (disabled || loading) && styles.buttonDisabled
      ]}
      onPress={onPress}
      disabled={disabled || loading}
      activeOpacity={0.7}
    >
      {loading ? (
        <ActivityIndicator size="small" color={Colors.lightBlue} />
      ) : (
        <Image
          source={require('@/assets/images/event_card/kinopoisk.png')}
          style={styles.icon}
          resizeMode="contain"
        />
      )}
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  button: {
    width: 40,
    height: 40,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 8,
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  icon: {
    width: 30,
    height: 30,
    borderRadius: 10
  },
});

export default KinopoiskButton;
