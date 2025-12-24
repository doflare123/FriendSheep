import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import { Image, ImageSourcePropType, StyleSheet, Text, TouchableOpacity } from 'react-native';

const CategoryButton = ({ title, onPress, imageSource }: { title: string; imageSource?: ImageSourcePropType; onPress: () => void }) => {
  const colors = useThemedColors();

  return (
    <TouchableOpacity 
      onPress={onPress} 
      style={[styles.button, { backgroundColor: colors.blue2 }]}
    >
      {imageSource && (
        <Image source={imageSource} style={styles.backgroundImage} />
      )}
      <Text style={[styles.text, { color: colors.white }]}>
        {title}
      </Text>
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  button: {
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    overflow: 'hidden',
    margin: 16,
    height: 40,
    width: '90%',
    alignSelf: 'center',
  },
  text: {
    fontSize: 20,
    fontFamily: Montserrat.regular,
  },
  backgroundImage: {
    ...StyleSheet.absoluteFillObject,
    resizeMode: 'repeat',
  },
});

export default CategoryButton;