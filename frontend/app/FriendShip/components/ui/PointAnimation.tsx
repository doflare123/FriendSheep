import React, { useEffect, useRef } from 'react';
import { Animated, Easing, StyleSheet, View } from 'react-native';
import { Colors } from '../../constants/Colors';

const POINT_SIZE = 22;
const POINT_SPACING = 18;
const POINT_DISTANCE = POINT_SIZE + POINT_SPACING;
const JUMP_HEIGHT = 36;

const PointAnimation = () => {
  const positions = [0, 1, 2].map(i => i * POINT_DISTANCE);

  const translateXs = useRef([
    new Animated.Value(positions[0]),
    new Animated.Value(positions[1]),
    new Animated.Value(positions[2]),
  ]).current;

  const translateYs = useRef([
    new Animated.Value(0),
    new Animated.Value(0),
    new Animated.Value(0),
  ]).current;

  const orderRef = useRef([0, 1, 2]);

  const animate = () => {
    const [a, b, c] = orderRef.current;
    const nextOrder = [b, c, a];

    const jumpToB = Animated.parallel([
      Animated.timing(translateXs[a], {
        toValue: positions[1],
        duration: 800,
        easing: Easing.inOut(Easing.ease),
        useNativeDriver: true,
      }),
      Animated.timing(translateXs[b], {
        toValue: positions[0],
        duration: 800,
        easing: Easing.inOut(Easing.ease),
        useNativeDriver: true,
      }),
      Animated.sequence([
        Animated.timing(translateYs[a], {
          toValue: -JUMP_HEIGHT,
          duration: 300,
          easing: Easing.out(Easing.quad),
          useNativeDriver: true,
        }),
        Animated.timing(translateYs[a], {
          toValue: 0,
          duration: 300,
          easing: Easing.in(Easing.quad),
          useNativeDriver: true,
        }),
      ]),
    ]);

    const jumpToC = Animated.parallel([
      Animated.timing(translateXs[a], {
        toValue: positions[2],
        duration: 800,
        easing: Easing.inOut(Easing.ease),
        useNativeDriver: true,
      }),
      Animated.timing(translateXs[c], {
        toValue: positions[1],
        duration: 800,
        easing: Easing.inOut(Easing.ease),
        useNativeDriver: true,
      }),
      Animated.sequence([
        Animated.timing(translateYs[a], {
          toValue: -JUMP_HEIGHT,
          duration: 300,
          easing: Easing.out(Easing.quad),
          useNativeDriver: true,
        }),
        Animated.timing(translateYs[a], {
          toValue: 0,
          duration: 300,
          easing: Easing.in(Easing.quad),
          useNativeDriver: true,
        }),
      ]),
    ]);

    Animated.sequence([jumpToB, jumpToC]).start(() => {
      orderRef.current = nextOrder;
      animate();
    });
  };

  useEffect(() => {
    animate();
  }, []);

  return (
    <View style={styles.container}>
      {[0, 1, 2].map((i) => (
        <Animated.View
          key={i}
          style={[
            styles.dot,
            {
              transform: [
                { translateX: translateXs[i] },
                { translateY: translateYs[i] },
              ],
              position: 'absolute',
              left: 0,
            },
          ]}
        />
      ))}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    width: POINT_DISTANCE * 3 + 20,
    height: POINT_SIZE * 3,
    justifyContent: 'center',
    alignItems: 'flex-start',
    position: 'relative',
    marginStart: 30,
  },
  dot: {
    width: POINT_SIZE,
    height: POINT_SIZE,
    borderRadius: POINT_SIZE / 2,
    backgroundColor: Colors.lightBlue3,
  },
});

export default PointAnimation;
