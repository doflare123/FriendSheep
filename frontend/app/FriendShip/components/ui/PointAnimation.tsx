import React, { useEffect, useRef } from 'react';
import { Animated, Easing, StyleSheet, View } from 'react-native';
import { Colors } from '../../constants/Colors';

const POINT_SIZE = 20;
const POINT_SPACING = 12;

const PointAnimation = () => {
  const dot1Anim = useRef(new Animated.Value(0)).current;
  const dot2Anim = useRef(new Animated.Value(0)).current;
  const dot3Anim = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    const jump = (anim: Animated.Value, delay: number) => {
      return Animated.loop(
        Animated.sequence([
          Animated.delay(delay),
          Animated.timing(anim, {
            toValue: -15,
            duration: 300,
            easing: Easing.inOut(Easing.ease),
            useNativeDriver: true,
          }),
          Animated.timing(anim, {
            toValue: 0,
            duration: 300,
            easing: Easing.inOut(Easing.ease),
            useNativeDriver: true,
          }),
        ])
      );
    };

    jump(dot1Anim, 0).start();
    jump(dot2Anim, 150).start();
    jump(dot3Anim, 300).start();
  }, []);

  return (
    <View style={styles.container}>
      {[dot1Anim, dot2Anim, dot3Anim].map((anim, i) => (
        <Animated.View
          key={i}
          style={[
            styles.dot,
            {
              transform: [{ translateY: anim }],
            },
          ]}
        />
      ))}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    marginTop: 5,
    marginBottom: 30,
    justifyContent: 'center',
    alignItems: 'center',
    gap: POINT_SPACING,
  },
  dot: {
    width: POINT_SIZE,
    height: POINT_SIZE,
    borderRadius: POINT_SIZE / 2,
    backgroundColor: Colors.lightBlue,
    marginHorizontal: 6,
  },
});

export default PointAnimation;
