import { useAuthContext } from '@/api/services/AuthContext';
import authService from '@/api/services/authService';
import InputSmall from '@/components/auth/InputSmall';
import { useToast } from '@/components/ToastContext';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import React, { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, AppState, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Logo from '../../components/auth/Logo';
import PointAnimation from '../../components/ui/PointAnimation';
import { Colors } from '../../constants/Colors';
import authorizeStyle from '../styles/authorizeStyle';

const RESEND_TIMEOUT_MINUTES = 15;

type ConfirmRouteParams = {
  sessionId: string;
};

const Confirm = () => {
  const navigation = useNavigation();
  const { showToast } = useToast();
  const { tempRegData, setTempRegData } = useAuthContext();
  const route = useRoute<RouteProp<{ Confirm: ConfirmRouteParams }, 'Confirm'>>();
  
  const { sessionId } = route.params || {};

  const email = tempRegData?.email || '';
  const username = tempRegData?.username || '';
  const password = tempRegData?.password || '';

  useEffect(() => {
    if (!tempRegData || !sessionId) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Данные регистрации не найдены',
      });
      navigation.navigate('Register' as never);
    }
  }, []);

  const [code, setCode] = useState('');
  const [remainingSeconds, setRemainingSeconds] = useState(RESEND_TIMEOUT_MINUTES * 60);
  const [resendAvailable, setResendAvailable] = useState(false);
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);

  const appState = useRef(AppState.currentState);
  const endTimeRef = useRef(Date.now() + remainingSeconds * 1000);

  const [attempts, setAttempts] = useState(0);
  const MAX_ATTEMPTS = 5;

  const handleConfirm = async () => {

    if (attempts >= MAX_ATTEMPTS) {
      showToast({
        type: 'error',
        title: 'Превышено количество попыток',
        message: 'Запросите новый код',
      });
      return;
    }

    const trimmedCode = code.trim().toUpperCase().replace(/[^A-Z0-9]/g, '');

    if (!trimmedCode || trimmedCode.length < 6) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Введите полный код подтверждения (6 символов)',
      });
      return;
    }

    if (!sessionId) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Отсутствует ID сессии',
      });
      return;
    }

    if (!username || !password) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Отсутствуют данные регистрации',
      });
      return;
    }

    setLoading(true);

    try {
      await authService.verifySession(sessionId, trimmedCode, 'register');
      await authService.createUser(email, username, password, sessionId);

      setTempRegData(null)

      showToast({
        type: 'success',
        title: 'Регистрация завершена!',
        message: 'Пожалуйста, войдите в систему',
      });

      setTimeout(() => {
        navigation.navigate('Login' as never);
      }, 1500);

    } catch (error: any) {
      setAttempts(prev => prev + 1);
      const errorMessage = error.message || 'Произошла ошибка';
      
      if (errorMessage.includes('Неверный код')) {
        showToast({
          type: 'error',
          title: 'Неверный код',
          message: 'Проверьте код из письма и попробуйте снова',
        });
      } else if (errorMessage.includes('Сессия не найдена')) {
        showToast({
          type: 'error',
          title: 'Сессия истекла',
          message: 'Запросите новый код подтверждения',
        });
      } else if (errorMessage.includes('уже существует') || errorMessage.includes('already exists')) {
        showToast({
          type: 'success',
          title: 'Аккаунт существует',
          message: 'Попробуйте войти',
        });
        
        setTimeout(() => {
          navigation.navigate('Login' as never);
        }, 1000);
      } else {
        showToast({
          type: 'error',
          title: 'Ошибка регистрации',
          message: errorMessage,
        });
      }
    } finally {
      setLoading(false);
    }
  };
  
    useEffect(() => {
      return () => {
        if (navigation.isFocused()) {
          setTempRegData(null);
        }
      };
    }, []);

  const handleResend = async () => {
    if (!email) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Email не найден',
      });
      return;
    }

    setResending(true);

    try {
      await authService.resendCode(email);

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: 'Код отправлен повторно',
      });

      const newEndTime = Date.now() + RESEND_TIMEOUT_MINUTES * 60 * 1000;
      endTimeRef.current = newEndTime;
      setRemainingSeconds(RESEND_TIMEOUT_MINUTES * 60);
      setResendAvailable(false);
      
      setCode('');
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось отправить код',
      });
    } finally {
      setResending(false);
    }
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

  const emailParts = email?.split('@') || ['', ''];
  const user_email_name = emailParts[0] || 'unknown';
  const user_email_domain = emailParts[1] || 'gmail.com';

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

        <InputSmall 
          value={code}
          onChangeText={setCode}
          editable={!loading && !resending}
          length={6}
        />

        {!resendAvailable ? (
          <Text style={authorizeStyle.hintLabel}>
            Повторно код можно отправить через {minutes}:{seconds.toString().padStart(2, '0')}
          </Text>
        ) : (
          <TouchableOpacity 
            onPress={handleResend}
            disabled={resending}
          >
            <Text style={[authorizeStyle.hintLabel, { color: Colors.lightBlue }]}>
              {resending ? 'Отправка...' : 'Отправить код повторно'}
            </Text>
          </TouchableOpacity>
        )}

        <Button 
          title={loading ? 'Регистрация...' : 'Подтвердить'} 
          onPress={handleConfirm}
          disabled={loading || resending || code.length < 6}
        />

        {(loading || resending) && (
          <ActivityIndicator 
            size="large" 
            color="#0000ff" 
            style={{ marginTop: 16 }} 
          />
        )}
      </KeyboardAwareScrollView>

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </View>
  );
};

export default Confirm;