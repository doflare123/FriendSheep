import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React from 'react';
import { Image, ImageSourcePropType, StyleSheet, Text, TouchableOpacity } from 'react-native';

const CategoryButton = ({ title, onPress, imageSource }: { title: string; imageSource?: ImageSourcePropType; onPress: () => void }) => (
 <TouchableOpacity onPress={onPress} style={styles.button}>
    {imageSource && (
      <Image source={imageSource} style={styles.backgroundImage} />
    )}
    <Text style={styles.text}>{title}</Text>
  </TouchableOpacity>
);

const styles = StyleSheet.create({
  button: {
    backgroundColor: Colors.blue,
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
    color: Colors.white,
    fontSize: 20,
    fontFamily: inter.regular,
  },
  backgroundImage: {
    ...StyleSheet.absoluteFillObject,
    resizeMode: 'repeat',
  },
});

export default CategoryButton;
