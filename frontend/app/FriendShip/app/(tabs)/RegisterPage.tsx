import { useAuthContext } from '@/api/services/AuthContext';
import authService from '@/api/services/authService';
import { useToast } from '@/components/ToastContext';
import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import { ActivityIndicator, Linking, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Input from '../../components/auth/Input';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Register = () => {
  const { setTempRegData } = useAuthContext();
  const navigation = useNavigation();
  const { showToast } = useToast();
  
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);

  const openURL = (url: string) => {
    Linking.openURL(url).catch(err =>
      console.error('Не удалось открыть ссылку:', err),
    );
  };

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleRegister = async () => {
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

    if (!username.trim()) {
      showToast({
        type: 'warning',
        title: 'Внимание',
        message: 'Введите имя пользователя',
      });
      return;
    }

    if (!password.trim() || password.length < 6) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Пароль должен содержать минимум 6 символов',
      });
      return;
    }

    setLoading(true);

    try {
      const response = await authService.createRegistrationSession(sanitizedEmail);

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: 'Код подтверждения отправлен на почту',
      });

      (navigation.navigate as any)('Confirm', {
        sessionId: response.session_id,
        email: sanitizedEmail,
        username: username.trim(),
        password,
      });
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось создать аккаунт',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={authorizeStyle.container}>
      <View style={authorizeStyle.topContainer}>
        <Logo />
        <Text style={authorizeStyle.title}>Регистрация</Text>
      </View>

      <KeyboardAwareScrollView
        enableOnAndroid
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={authorizeStyle.scrollContent}
        extraScrollHeight={20}
        showsVerticalScrollIndicator={false}>

        <Text style={authorizeStyle.label}>Имя пользователя</Text>
        <Input
          placeholder="Имя пользователя"
          value={username}
          onChangeText={setUsername}
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

        <Text style={authorizeStyle.label}>Почта</Text>
        <Input
          placeholder="user_email@gmail.com"
          value={email}
          onChangeText={setEmail}
          keyboardType="email-address"
          autoCapitalize="none"
          editable={!loading}
        />

        <View style={[authorizeStyle.account, {justifyContent: 'flex-end'}]}>
          <TouchableOpacity 
            onPress={() => navigation.navigate('Login' as never)}
            disabled={loading}
          >
            <Text>Есть аккаунт?</Text>
          </TouchableOpacity>
        </View>

        <Text style={authorizeStyle.terms}>
          При создании аккаунта вы соглашаетесь с условиями{' '}
          <Text
            style={authorizeStyle.link}
            onPress={() =>
              openURL('https://www.youtube.com/@adventurekabanchikov')
            }
          >
            Пользовательского соглашения
          </Text>{' '}
          и{' '}
          <Text
            style={authorizeStyle.link}
            onPress={() =>
              openURL(
                'https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=RDdQw4w9WgXcQ&start_radio=1',
              )
            }
          >
            Политики конфиденциальности
          </Text>
        </Text>

        <Button
          title={loading ? 'Отправка...' : 'Зарегистрироваться'}
          onPress={handleRegister}
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

export default Register;