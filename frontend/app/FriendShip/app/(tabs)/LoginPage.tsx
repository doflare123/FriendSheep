import authService from '@/api/services/authService';
import { clearTokens } from '@/api/storage/tokenStorage';
import { useAuthContext } from '@/components/auth/AuthContext';
import { useToast } from '@/components/ToastContext';
import { Colors } from '@/constants/Colors';
import { useNavigation } from '@react-navigation/native';
import React, { useMemo, useState } from 'react';
import { ActivityIndicator, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Input from '../../components/auth/Input';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Login = () => {
  const navigation = useNavigation();
  const { showToast } = useToast();
  const { setIsAuthenticated } = useAuthContext();
  
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);

  const hasRussianChars = useMemo(() => {
    return /[а-яА-ЯёЁ]/.test(password);
  }, [password]);

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const isEmailValid = useMemo(() => {
    if (email.length === 0) return true;
    return validateEmail(email.trim().toLowerCase());
  }, [email]);

  const handleLogin = async () => {
    const sanitizedEmail = email.trim().toLowerCase();
    
    if (!sanitizedEmail) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Введите email',
      });
      return;
    }

    if (!validateEmail(sanitizedEmail)) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Введите корректный email',
      });
      return;
    }

    if (!password.trim()) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Введите пароль',
      });
      return;
    }

    setLoading(true);

    try {
      await clearTokens();
      console.log('[Login] ✅ Старые токены очищены');

      await authService.login(sanitizedEmail, password);

      console.log('[Login] ✅ Успешная авторизация');

      setIsAuthenticated(true);

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: 'Вы вошли в систему',
      });

    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка входа',
        message: error.message || 'Не удалось войти в систему',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={authorizeStyle.container}>
      <View style={authorizeStyle.topContainer}>
        <Logo />
        <Text style={authorizeStyle.title}>Вход</Text>
      </View>

      <KeyboardAwareScrollView
        enableOnAndroid
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={authorizeStyle.scrollContent}
        extraScrollHeight={20}
        showsVerticalScrollIndicator={false}>

        <Text style={authorizeStyle.label}>Почта</Text>
        <Input
          placeholder="user_email@gmail.com"
          value={email}
          onChangeText={setEmail}
          keyboardType="email-address"
          autoCapitalize="none"
          editable={!loading}
          isValid={isEmailValid}
        />

        {email.length > 0 && !isEmailValid && (
          <Text style={authorizeStyle.passwordValidation}>
            Введите корректный email (должен содержать @ и домен)
          </Text>
        )}

        <Text style={authorizeStyle.label}>Пароль</Text>
        <Input
          placeholder="Пароль"
          secureTextEntry
          value={password}
          onChangeText={setPassword}
          editable={!loading}
          borderColor={hasRussianChars ? Colors.orange : undefined}
        />

        {hasRussianChars && (
          <Text style={[authorizeStyle.passwordValidation, {color: Colors.orange}]}>
            Включена русская раскладка!
          </Text>
        )}

        <View style={authorizeStyle.account}>
          <TouchableOpacity 
            onPress={() => navigation.navigate('ForgotPassword' as never)}
            disabled={loading}
          >
            <Text style={authorizeStyle.password}>Забыли пароль?</Text>
          </TouchableOpacity>
          <TouchableOpacity 
            onPress={() => navigation.navigate('Register' as never)}
            disabled={loading}
          >
            <Text style={authorizeStyle.account}>Нет аккаунта?</Text>
          </TouchableOpacity>
        </View>

        <Button 
          title={loading ? 'Вход...' : 'Войти'} 
          onPress={handleLogin}
          disabled={loading || !isEmailValid || email.length === 0}
        />

        {loading && (
          <ActivityIndicator 
            size="large" 
            color={Colors.blue} 
            style={{ marginTop: 16 }} 
          />
        )}
      </KeyboardAwareScrollView>

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </View>
  );
};

export default Login;