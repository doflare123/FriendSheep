import authService from '@/api/services/authService';
import { useToast } from '@/components/ToastContext';
import { Colors } from '@/constants/Colors';
import { hasRussianChars, validatePassword } from '@/utils/validators';
import { useNavigation, useRoute } from '@react-navigation/native';
import React, { useMemo, useState } from 'react';
import { ActivityIndicator, Text, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Input from '../../components/auth/Input';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

interface RouteParams {
  sessionId: string;
  email: string;
  code: string;
}

const ResetPassword = () => {
  const navigation = useNavigation();
  const route = useRoute();
  const { showToast } = useToast();
  
  const { sessionId, email, code } = route.params as RouteParams;
  
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const passwordValidation = useMemo(() => validatePassword(newPassword), [newPassword]);
  const showRussianWarning = useMemo(() => hasRussianChars(newPassword), [newPassword]);

  const handleResetPassword = async () => {
    if (!passwordValidation.isValid) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Пароль не соответствует требованиям',
      });
      return;
    }

    if (newPassword !== confirmPassword) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Пароли не совпадают',
      });
      return;
    }

    setLoading(true);

    try {
      console.log('[ResetPassword] Confirming password reset...');
      console.log('[ResetPassword] Email:', email);
      console.log('[ResetPassword] SessionId:', sessionId);
      console.log('[ResetPassword] Code:', code);

      await authService.confirmPasswordReset(
        email,
        sessionId,
        code,
        newPassword
      );

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: 'Пароль успешно изменен',
      });

      setTimeout(() => {
        navigation.navigate('Login' as never);
      }, 1500);
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось изменить пароль',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={authorizeStyle.container}>
      <View style={authorizeStyle.topContainer}>
        <Logo />
        <Text style={authorizeStyle.title}>Новый пароль</Text>
      </View>

      <KeyboardAwareScrollView
        enableOnAndroid
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={authorizeStyle.scrollContent}
        extraScrollHeight={20}
        showsVerticalScrollIndicator={false}>

        <Text style={authorizeStyle.terms}>
          Установите новый пароль для {email}
        </Text>

        <Text style={authorizeStyle.label}>Новый пароль</Text>
        <Input
          placeholder="Минимум 10 символов"
          secureTextEntry
          value={newPassword}
          onChangeText={setNewPassword}
          editable={!loading}
          isValid={newPassword.length === 0 || (passwordValidation.isValid && !showRussianWarning)}
          borderColor={showRussianWarning ? Colors.orange : undefined}
        />

        {showRussianWarning && (
          <Text style={[authorizeStyle.passwordValidation, {color: Colors.orange}]}>
            Включена русская раскладка!
          </Text>
        )}

        {newPassword.length > 0 && !passwordValidation.isValid && !showRussianWarning && (
          <Text style={authorizeStyle.passwordValidation}>
            Необходимо: {passwordValidation.missingRequirements.join(', ')}
          </Text>
        )}

        <Text style={authorizeStyle.label}>Подтвердите пароль</Text>
        <Input
          placeholder="Повторите пароль"
          secureTextEntry
          value={confirmPassword}
          onChangeText={setConfirmPassword}
          editable={!loading}
          isValid={confirmPassword.length === 0 || newPassword === confirmPassword}
        />

        {confirmPassword.length > 0 && newPassword !== confirmPassword && (
          <Text style={authorizeStyle.passwordValidation}>
            Пароли не совпадают
          </Text>
        )}

        <Button
          title={loading ? 'Сохранение...' : 'Сохранить пароль'}
          onPress={handleResetPassword}
          disabled={loading || !passwordValidation.isValid || newPassword !== confirmPassword}
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

export default ResetPassword;