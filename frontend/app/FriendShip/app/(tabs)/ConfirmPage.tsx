import authService from '@/api/services/authService';
import { useAuthContext } from '@/components/auth/AuthContext';
import InputSmall from '@/components/auth/InputSmall';
import { useToast } from '@/components/ToastContext';
import { Colors } from '@/constants/Colors';
import { useThemedColors } from '@/hooks/useThemedColors';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import React, { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, AppState, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Logo from '../../components/auth/Logo';
import PointAnimation from '../../components/ui/PointAnimation';
import authorizeStyle from '../styles/authorizeStyle';

const RESEND_TIMEOUT_MINUTES = 15;

type ConfirmRouteParams = {
  sessionId: string;
  type: 'register' | 'reset';
  email?: string;
};

const Confirm = () => {
  const colors = useThemedColors();
  const navigation = useNavigation();
  const { showToast } = useToast();
  const { tempRegData, setTempRegData } = useAuthContext();
  const route = useRoute<RouteProp<{ Confirm: ConfirmRouteParams }, 'Confirm'>>();
  
  const { sessionId, type = 'register', email: routeEmail } = route.params || {};

  const email = type === 'register' ? (tempRegData?.email || '') : (routeEmail || '');
  const username = tempRegData?.username || '';
  const password = tempRegData?.password || '';

  useEffect(() => {
    if (type === 'register' && (!tempRegData || !sessionId)) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Данные регистрации не найдены',
      });
      navigation.navigate('Register' as never);
    } else if (type === 'reset' && (!sessionId || !routeEmail)) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Данные для сброса пароля не найдены',
      });
      navigation.navigate('ForgotPassword' as never);
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

    setLoading(true);

    try {
      if (type === 'register') {
        if (!username || !password) {
          showToast({
            type: 'error',
            title: 'Ошибка',
            message: 'Отсутствуют данные регистрации',
          });
          setLoading(false);
          return;
        }

        console.log('[Confirm] Starting registration process...');
        console.log('[Confirm] Email:', email);
        console.log('[Confirm] Username:', username);
        console.log('[Confirm] SessionId:', sessionId);
        console.log('[Confirm] Code:', trimmedCode);

        console.log('[Confirm] Step 1: Verifying session...');
        const verifyResponse = await authService.verifySession(sessionId, trimmedCode, 'register');
        console.log('[Confirm] Verify response:', verifyResponse);

        await new Promise(resolve => setTimeout(resolve, 500));

        console.log('[Confirm] Step 2: Creating user...');
        await authService.createUser(email, username, password, sessionId);
        console.log('[Confirm] User created successfully!');

        setTempRegData(null);

        showToast({
          type: 'success',
          title: 'Регистрация завершена!',
          message: 'Пожалуйста, войдите в систему',
        });

        setTimeout(() => {
          navigation.navigate('Login' as never);
        }, 1500);

      } else if (type === 'reset') {
        console.log('[Confirm] Password reset - verifying code...');
        console.log('[Confirm] Email:', email);
        console.log('[Confirm] SessionId:', sessionId);
        console.log('[Confirm] Code:', trimmedCode);

        const verifyResponse = await authService.verifySession(sessionId, trimmedCode, 'reset_password');
        console.log('[Confirm] Verify response:', verifyResponse);

        showToast({
          type: 'success',
          title: 'Код подтвержден!',
          message: 'Теперь установите новый пароль',
        });

        setTimeout(() => {
          (navigation.navigate as any)('ResetPassword', {
            sessionId,
            email,
            code: trimmedCode,
          });
        }, 1000);
      }

    } catch (error: any) {
      setAttempts(prev => prev + 1);
      const errorMessage = error.message || 'Произошла ошибка';
      
      console.error('[Confirm] Error during confirmation:', error);
      console.error('[Confirm] Error message:', errorMessage);
      console.error('[Confirm] Current attempts:', attempts + 1);
      
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
          title: type === 'register' ? 'Ошибка регистрации' : 'Ошибка сброса пароля',
          message: errorMessage,
        });
      }
    } finally {
      setLoading(false);
    }
  };
  
  useEffect(() => {
    return () => {
      if (navigation.isFocused() && type === 'register') {
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
      let response;
      
      if (type === 'register') {
        response = await authService.resendCode(email);
      } else {
        response = await authService.requestPasswordReset(email);
      }

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
      setAttempts(0);
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

  const getTitle = () => {
    return type === 'register' ? 'Код подтверждения' : 'Подтверждение сброса';
  };

  const getButtonTitle = () => {
    if (loading) {
      return type === 'register' ? 'Регистрация...' : 'Проверка...';
    }
    return type === 'register' ? 'Подтвердить' : 'Продолжить';
  };

  return (
    <View style={[authorizeStyle.container, { backgroundColor: colors.white }]}>
      <View style={[authorizeStyle.topContainer, { marginBottom: 0 }]}>
        <Logo />
        <Text style={[authorizeStyle.title, { color: Colors.lightBlue }]}>
          {getTitle()}
        </Text>
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
          style={[
            authorizeStyle.label,
            { fontSize: 16, textAlign: 'center', color: colors.black }
          ]}
        >
          Введите код подтверждения из письма, отправленного на почту {user_email_name}@{user_email_domain}
        </Text>

        <InputSmall 
          value={code}
          onChangeText={setCode}
          editable={!loading && !resending}
          length={6}
        />

        {attempts >= MAX_ATTEMPTS && (
          <Text style={[
            authorizeStyle.hintLabel,
            { color: colors.red, marginBottom: 10 }
          ]}>
            Превышено количество попыток. Запросите новый код.
          </Text>
        )}

        {!resendAvailable ? (
          <Text style={[authorizeStyle.hintLabel, { color: colors.grey }]}>
            Повторно код можно отправить через {minutes}:{seconds.toString().padStart(2, '0')}
          </Text>
        ) : (
          <TouchableOpacity 
            onPress={handleResend}
            disabled={resending}
          >
            <Text style={[authorizeStyle.hintLabel, { color: colors.lightBlue }]}>
              {resending ? 'Отправка...' : 'Отправить код повторно'}
            </Text>
          </TouchableOpacity>
        )}

        {type === 'reset' && (
          <TouchableOpacity 
            onPress={() => navigation.navigate('ForgotPassword' as never)}
            disabled={loading || resending}
            style={{ marginTop: 10 }}
          >
            <Text style={[authorizeStyle.hintLabel, { color: colors.lightBlue }]}>
              Изменить email
            </Text>
          </TouchableOpacity>
        )}

        <Button 
          title={getButtonTitle()} 
          onPress={handleConfirm}
          disabled={loading || resending || code.length < 6 || attempts >= MAX_ATTEMPTS}
        />

        {(loading || resending) && (
          <ActivityIndicator 
            size="large" 
            color={colors.blue} 
            style={{ marginTop: 16 }} 
          />
        )}
      </KeyboardAwareScrollView>

      <Text style={[authorizeStyle.footer, { color: colors.grey }]}>
        ©NecroDwarf
      </Text>
    </View>
  );
};

export default Confirm;