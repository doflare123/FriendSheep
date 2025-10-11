import InputSmall from '@/components/auth/InputSmall';
import { useNavigation } from '@react-navigation/native';
import React, { useEffect, useRef, useState } from 'react';
import { AppState, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Logo from '../../components/auth/Logo';
import PointAnimation from '../../components/ui/PointAnimation';
import { Colors } from '../../constants/Colors';
import authorizeStyle from '../styles/authorizeStyle';

const RESEND_TIMEOUT_MINUTES = 15;

const Confirm = () => {
  const navigation = useNavigation();
  const [remainingSeconds, setRemainingSeconds] = useState(RESEND_TIMEOUT_MINUTES * 60);
  const [resendAvailable, setResendAvailable] = useState(false);
  const appState = useRef(AppState.currentState);
  const endTimeRef = useRef(Date.now() + remainingSeconds * 1000);

  const handleConfirm = () => {
    navigation.navigate('Done' as never);
  };

  const handleResend = () => {
    const newEndTime = Date.now() + RESEND_TIMEOUT_MINUTES * 60 * 1000;
    endTimeRef.current = newEndTime;
    setRemainingSeconds(RESEND_TIMEOUT_MINUTES * 60);
    setResendAvailable(false);
  };

  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      const diff = Math.max(0, Math.floor((endTimeRef.current - now) / 1000));
      setRemainingSeconds(diff);
      setResendAvailable(diff <= 0);
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const subscription = AppState.addEventListener('change', (nextAppState) => {
      if (appState.current.match(/inactive|background/) && nextAppState === 'active') {
        const now = Date.now();
        const diff = Math.max(0, Math.floor((endTimeRef.current - now) / 1000));
        setRemainingSeconds(diff);
        setResendAvailable(diff <= 0);
      }
      appState.current = nextAppState;
    });

    return () => subscription.remove();
  }, []);

  const minutes = Math.floor(remainingSeconds / 60);
  const seconds = remainingSeconds % 60;

  const user_email_name = 'pochta';
  const user_email_domain = 'gmail.com';

  return (
    <View style={authorizeStyle.container}>
      <View style={[authorizeStyle.topContainer, { marginBottom: 0 }]}>
        <Logo />
        <Text style={authorizeStyle.title}>Код подтверждения</Text>
        <PointAnimation />
      </View>

      <KeyboardAwareScrollView
        enableOnAndroid
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={authorizeStyle.scrollContent}
        extraScrollHeight={20}
        showsVerticalScrollIndicator={false}
      >
        <Text
          style={[authorizeStyle.label, { fontSize: 16, textAlign: 'center' }]}
        >
          Введите код подтверждения из письма, отправленного на почту {user_email_name}@{user_email_domain}
        </Text>

        <InputSmall />

        {!resendAvailable ? (
          <Text style={authorizeStyle.hintLabel}>
            Повторно код можно отправить через {minutes}:{seconds.toString().padStart(2, '0')}
          </Text>
        ) : (
          <TouchableOpacity onPress={handleResend}>
            <Text style={[authorizeStyle.hintLabel, { color: Colors.lightBlue }]}>
              Отправить код повторно
            </Text>
          </TouchableOpacity>
        )}

        <Button title="Подтвердить" onPress={handleConfirm} />
      </KeyboardAwareScrollView>

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </View>
  );
};

export default Confirm;
