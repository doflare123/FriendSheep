import authService from '@/api/services/authService';
import { useAuthContext } from '@/components/auth/AuthContext';
import { useToast } from '@/components/ToastContext';
import { Colors } from '@/constants/Colors';
import { useThemedColors } from '@/hooks/useThemedColors';
import profanityFilter from '@/utils/profanityFilter';
import { hasRussianChars, validatePassword, validateUsername } from '@/utils/validators';
import { useNavigation } from '@react-navigation/native';
import React, { useMemo, useState } from 'react';
import { ActivityIndicator, Linking, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Input from '../../components/auth/Input';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Register = () => {
  const colors = useThemedColors();
  const { setTempRegData } = useAuthContext();
  const navigation = useNavigation();
  const { showToast } = useToast();
  
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const usernameValidation = useMemo(() => validateUsername(username), [username]);
  const passwordValidation = useMemo(() => validatePassword(password), [password]);
  const showRussianWarning = useMemo(() => hasRussianChars(password), [password]);
  
  const isEmailValid = useMemo(() => {
    if (email.length === 0) return true;
    return validateEmail(email.trim().toLowerCase());
  }, [email]);

  const openURL = (url: string) => {
    Linking.openURL(url).catch(err =>
      console.error('Не удалось открыть ссылку:', err),
    );
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

    if (!usernameValidation.isValid) {
      showToast({
        type: 'warning',
        title: 'Внимание',
        message: 'Имя пользователя не соответствует требованиям',
      });
      return;
    }

    if (!passwordValidation.isValid) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Пароль не соответствует требованиям',
      });
      return;
    }

    if (password !== confirmPassword) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Пароли не совпадают',
      });
      return;
    }

    setLoading(true);

    try {
      const response = await authService.createRegistrationSession(sanitizedEmail);

      const cleanedUsername = profanityFilter.clean(username.trim());

      setTempRegData({
        email: sanitizedEmail,
        username: cleanedUsername,
        password: password,
      });

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: 'Код подтверждения отправлен на почту',
      });

      (navigation.navigate as any)('Confirm', {
        sessionId: response.session_id,
        type: 'register',
      });
    } catch {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось создать аккаунт',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={[authorizeStyle.container, { backgroundColor: colors.white }]}>
      <View style={authorizeStyle.topContainer}>
        <Logo />
        <Text style={[authorizeStyle.title, { color: Colors.lightBlue }]}>
          Регистрация
        </Text>
      </View>

      <KeyboardAwareScrollView
        enableOnAndroid
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={authorizeStyle.scrollContent}
        extraScrollHeight={20}
        showsVerticalScrollIndicator={false}>

        <Text style={[authorizeStyle.label, { color: colors.black }]}>
          Имя пользователя 
          {username.length > 0 && (
            <Text style={{ color: usernameValidation.length > 40 ? colors.red : colors.grey }}>
              {' '}({usernameValidation.length}/{usernameValidation.maxLength})
            </Text>
          )}
        </Text>
        <Input
          placeholder="От 5 до 40 символов"
          value={username}
          onChangeText={setUsername}
          editable={!loading}
          isValid={username.length === 0 || usernameValidation.isValid}
          maxLength={40}
        />

        {username.length > 0 && !usernameValidation.isValid && (
          <Text style={[authorizeStyle.passwordValidation, { color: colors.red }]}>
            {usernameValidation.missingRequirements.join(', ')}
          </Text>
        )}

        <Text style={[authorizeStyle.label, { color: colors.black }]}>
          Пароль
        </Text>
        <Input
          placeholder="Минимум 10 символов"
          secureTextEntry
          value={password}
          onChangeText={setPassword}
          editable={!loading}
          isValid={password.length === 0 || (passwordValidation.isValid && !showRussianWarning)}
          borderColor={showRussianWarning ? colors.orange : undefined}
        />

        {showRussianWarning && (
          <Text style={[authorizeStyle.passwordValidation, { color: colors.orange }]}>
            Включена русская раскладка!
          </Text>
        )}

        {password.length > 0 && !passwordValidation.isValid && !showRussianWarning && (
          <Text style={[authorizeStyle.passwordValidation, { color: colors.red }]}>
            Необходимо: {passwordValidation.missingRequirements.join(', ')}
          </Text>
        )}

        <Text style={[authorizeStyle.label, { color: colors.black }]}>
          Подтвердите пароль
        </Text>
        <Input
          placeholder="Повторите пароль"
          secureTextEntry
          value={confirmPassword}
          onChangeText={setConfirmPassword}
          editable={!loading}
          isValid={confirmPassword.length === 0 || password === confirmPassword}
        />

        {confirmPassword.length > 0 && password !== confirmPassword && (
          <Text style={[authorizeStyle.passwordValidation, { color: colors.red }]}>
            Пароли не совпадают
          </Text>
        )}

        <Text style={[authorizeStyle.label, { color: colors.black }]}>
          Почта
        </Text>
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
          <Text style={[authorizeStyle.passwordValidation, { color: colors.red }]}>
            Введите корректный email (должен содержать @ и домен)
          </Text>
        )}

        <View style={[authorizeStyle.account, {justifyContent: 'flex-end'}]}>
          <TouchableOpacity 
            onPress={() => navigation.navigate('Login' as never)}
            disabled={loading}
          >
            <Text style={[authorizeStyle.account, { color: Colors.lightBlue }]}>
              Есть аккаунт?
            </Text>
          </TouchableOpacity>
        </View>

        <Text style={[authorizeStyle.terms, { color: colors.grey }]}>
          При создании аккаунта вы соглашаетесь с условиями{' '}
          <Text
            style={[authorizeStyle.link, { color: Colors.lightBlue }]}
            onPress={() =>
              openURL('https://friendsheep.ru/info/privacy')
            }
          >
            Пользовательского соглашения
          </Text>{' '}
          и{' '}
          <Text
            style={[authorizeStyle.link, { color: Colors.lightBlue }]}
            onPress={() =>
              openURL(
                'https://friendsheep.ru/info/privacy',
              )
            }
          >
            Политики конфиденциальности
          </Text>
        </Text>

        <Button
          title={loading ? 'Отправка...' : 'Зарегистрироваться'}
          onPress={handleRegister}
          disabled={loading || !usernameValidation.isValid || !passwordValidation.isValid || password !== confirmPassword || !isEmailValid}
        />

        {loading && (
          <ActivityIndicator 
            size="large" 
            color={colors.blue2}
            style={{ marginTop: 16 }} 
          />
        )}

        <Text style={[authorizeStyle.footer, { color: colors.grey }]}>
          ©NecroDwarf
        </Text>
      </KeyboardAwareScrollView>
    </View>
  );
};

export default Register;