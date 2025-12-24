import { Colors } from "@/constants/Colors";
import { Montserrat } from "@/constants/Montserrat";
import { useThemedColors } from "@/hooks/useThemedColors";
import React, { useEffect, useRef } from "react";
import { Animated, Easing, Image, PanResponder, StyleSheet, Text, View } from "react-native";

interface ToastProps {
  visible: boolean;
  type?: "success" | "error" | "warning";
  title: string;
  message: string;
  onHide: () => void;
}

const icons = {
  success: require("@/assets/images/toast/success.png"),
  error: require("@/assets/images/toast/error.png"),
  warning: require("@/assets/images/toast/warning.png"),
};

const borderColors = {
  success: Colors.green,
  error: Colors.red,
  warning: Colors.orange
};

const Toast: React.FC<ToastProps> = ({
  visible,
  type = "success",
  title,
  message,
  onHide,
}) => {
  const colors = useThemedColors();
  const translateX = useRef(new Animated.Value(300)).current;
  const hideCalled = useRef(false);
  const timerRef = useRef<number | null>(null);

  const hideToast = () => {
    if (hideCalled.current) return;
    hideCalled.current = true;

    if (timerRef.current !== null) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }

    Animated.timing(translateX, {
      toValue: 300,
      duration: 300,
      easing: Easing.in(Easing.cubic),
      useNativeDriver: true,
    }).start(() => {
      onHide();
    });
  };

  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => true,
      onMoveShouldSetPanResponder: (_, gestureState) => {
        return Math.abs(gestureState.dx) > 5;
      },
      onPanResponderGrant: () => {
        if (timerRef.current !== null) {
          clearTimeout(timerRef.current);
          timerRef.current = null;
        }
      },
      onPanResponderMove: (_, gestureState) => {
        if (gestureState.dx > 0) {
          translateX.setValue(gestureState.dx);
        }
      },
      onPanResponderRelease: (_, gestureState) => {
        if (gestureState.dx > 80) {
          hideCalled.current = true;
          
          if (timerRef.current !== null) {
            clearTimeout(timerRef.current);
            timerRef.current = null;
          }

          Animated.timing(translateX, {
            toValue: 400,
            duration: 200,
            easing: Easing.out(Easing.ease),
            useNativeDriver: true,
          }).start(() => {
            onHide();
          });
        } else {
          Animated.spring(translateX, {
            toValue: 0,
            useNativeDriver: true,
            friction: 8,
          }).start();
          
          timerRef.current = setTimeout(hideToast, 3000) as unknown as number;
        }
      },
    })
  ).current;

  useEffect(() => {
    if (visible) {
      hideCalled.current = false;
      translateX.setValue(300);
      
      Animated.timing(translateX, {
        toValue: 0,
        duration: 400,
        easing: Easing.out(Easing.cubic),
        useNativeDriver: true,
      }).start();

      timerRef.current = setTimeout(hideToast, 3000) as unknown as number;

      return () => {
        if (timerRef.current !== null) {
          clearTimeout(timerRef.current);
          timerRef.current = null;
        }
      };
    }
  }, [visible]);

  if (!visible) return null;

  return (
    <Animated.View
      {...panResponder.panHandlers}
      style={[
        styles.container,
        {
          transform: [{ translateX }],
          backgroundColor: colors.white,
          borderColor: borderColors[type],
        },
      ]}
    >
      <View style={styles.content}>
        <View style={styles.header}>
          <Image source={icons[type]} style={styles.icon} />
          <Text 
            style={[styles.title, { color: colors.black }]} 
            numberOfLines={2}
            ellipsizeMode="tail"
          >
            {title}
          </Text>
        </View>
        <Text 
          style={[styles.message, { color: colors.black }]} 
          numberOfLines={3}
          ellipsizeMode="tail"
        >
          {message}
        </Text>
      </View>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  container: {
    position: "absolute",
    top: 100,
    right: 10,
    padding: 12,
    borderRadius: 12,
    borderWidth: 1,
    shadowColor: "#000",
    shadowOpacity: 0.1,
    shadowOffset: { width: 0, height: 2 },
    shadowRadius: 4,
    elevation: 10,
    zIndex: 9999,
    maxWidth: "90%",
  },
  content: {
    width: '100%',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 2,
    width: '100%',
  },
  icon: { 
    width: 24, 
    height: 24, 
    marginRight: 4,
    resizeMode: "contain",
  },
  title: { 
    fontFamily: Montserrat.bold, 
    fontSize: 14,
    flex: 1,
  },
  message: { 
    fontFamily: Montserrat.regular, 
    fontSize: 13,
    width: '100%',
  },
});

export default Toast;