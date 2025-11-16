import authService from '@/api/services/authService';
import { getTokens } from '@/api/storage/tokenStorage';
import { useToast } from '@/components/ToastContext';
import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import { ActivityIndicator, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Input from '../../components/auth/Input';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Login = () => {
  const navigation = useNavigation();
  const { showToast } = useToast();
  
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleLogin = async () => {
    if (!email.trim()) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Введите email',
      });
      return;
    }

    if (!validateEmail(email)) {
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
      console.log("🔥 LOGIN STARTED");

      const data = await authService.login(email, password);

      console.log("🔥 LOGIN RESPONSE:", data);

      const savedTokens = await getTokens();
      console.log('[Login] Токены после логина:', savedTokens ? 'СОХРАНЕНЫ' : 'НЕ СОХРАНЕНЫ');

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: 'Вы вошли в систему',
      });

      setTimeout(() => {
        navigation.navigate('MainPage' as never);
      }, 500);

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
        />

        <Text style={authorizeStyle.label}>Пароль</Text>
        <Input
          placeholder="Пароль"
          secureTextEntry
          value={password}
          onChangeText={setPassword}
          editable={!loading}
        />

        <View style={authorizeStyle.account}>
          <TouchableOpacity 
            onPress={() => navigation.navigate('Register' as never)}
            disabled={loading}
          >
            <Text style={authorizeStyle.password}>Забыли пароль?</Text>
          </TouchableOpacity>
          <TouchableOpacity 
            onPress={() => navigation.navigate('Register' as never)}
            disabled={loading}
          >
            <Text>Нет аккаунта?</Text>
          </TouchableOpacity>
        </View>

        <Button 
          title={loading ? 'Вход...' : 'Войти'} 
          onPress={handleLogin}
          disabled={loading}
        />

        {loading && (
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

export default Login;