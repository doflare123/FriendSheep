import { Colors } from "@/constants/Colors";
import { inter } from "@/constants/Inter";
import React, { useEffect, useRef } from "react";
import { Animated, Easing, Image, StyleSheet, Text, TouchableWithoutFeedback, View } from "react-native";

interface ToastProps {
  visible: boolean;
  type?: "success" | "error" | "warning";
  title: string;
  message: string;
  onHide: () => void;
}

const icons = {
  success: require("../assets/images/toast/success.png"),
  error: require("../assets/images/toast/error.png"),
  warning: require("../assets/images/toast/warning.png"),
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

    const slideAnim = useRef(new Animated.Value(300)).current;
    const hideCalled = useRef(false);
    const timerRef = useRef<number | null>(null);

    const hideToast = () => {
    if (hideCalled.current) return;
    hideCalled.current = true;

    if (timerRef.current !== null) {
        clearTimeout(timerRef.current);
    }

    Animated.timing(slideAnim, {
        toValue: 300,
        duration: 300,
        easing: Easing.in(Easing.cubic),
        useNativeDriver: true,
    }).start(() => onHide());
    };

    useEffect(() => {
    hideCalled.current = false;

    if (visible) {
        Animated.timing(slideAnim, {
        toValue: 0,
        duration: 400,
        easing: Easing.out(Easing.cubic),
        useNativeDriver: true,
        }).start();

        timerRef.current = setTimeout(hideToast, 3000);

        return () => {
        if (timerRef.current !== null) {
            clearTimeout(timerRef.current);
        }
        };
    }
    }, [visible]);

  if (!visible) return null;

  return (
    <TouchableWithoutFeedback onPress={hideToast}>
      <Animated.View
        style={[
          styles.container,
          {
            transform: [{ translateX: slideAnim }],
            backgroundColor: Colors.white,
            borderColor: borderColors[type],
          },
        ]}
      >
        <View style={{ flex: 1 }}>
          <View style={{ flexDirection: 'row', alignItems: 'center' }}>
            <Text style={styles.title}>{title}</Text>
            <Image source={icons[type]} style={styles.icon} />
          </View>
          <Text style={styles.message}>{message}</Text>
        </View>
      </Animated.View>
    </TouchableWithoutFeedback>
  );
};

const styles = StyleSheet.create({
  container: {
    position: "absolute",
    top: 40,
    right: 10,
    flexDirection: "row",
    alignItems: "center",
    padding: 12,
    borderRadius: 12,
    borderWidth: 1,
    shadowColor: "#000",
    shadowOpacity: 0.1,
    shadowOffset: { width: 0, height: 2 },
    shadowRadius: 4,
    elevation: 3,
    maxWidth: "90%",
  },
  icon: { width: 24, height: 24, marginLeft: 4, resizeMode: "contain" },
  title: { fontFamily: inter.bold, fontSize: 14, color: Colors.black },
  message: { fontFamily: inter.regular, fontSize: 13, color: Colors.black },
});

export default Toast;
