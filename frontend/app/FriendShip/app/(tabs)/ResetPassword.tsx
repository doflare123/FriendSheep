import authService from '@/api/services/authService';
import { useToast } from '@/components/ToastContext';
import { Colors } from '@/constants/Colors';
import { useNavigation, useRoute } from '@react-navigation/native';
import React, { useState } from 'react';
import { ActivityIndicator, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Input from '../../components/auth/Input';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

interface RouteParams {
  sessionId: string;
  email: string;
}

const ResetPassword = () => {
  const navigation = useNavigation();
  const route = useRoute();
  const { showToast } = useToast();
  
  const { sessionId, email } = route.params as RouteParams;
  
  const [code, setCode] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [resendLoading, setResendLoading] = useState(false);

  const handleResetPassword = async () => {
    if (!code.trim()) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Введите код подтверждения',
      });
      return;
    }

    if (!newPassword.trim() || newPassword.length < 6) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Пароль должен содержать минимум 6 символов',
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
      await authService.confirmPasswordReset(
        email,
        sessionId,
        code.trim(),
        newPassword
      );

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: 'Пароль успешно изменен',
      });

      navigation.navigate('Login' as never);
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось сбросить пароль',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleResendCode = async () => {
    setResendLoading(true);

    try {
      const response = await authService.requestPasswordReset(email);

      showToast({
        type: 'success',
        title: 'Код отправлен',
        message: 'Новый код отправлен на вашу почту',
      });

      (navigation.setParams as any)({ sessionId: response.session_id });
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось отправить код повторно',
      });
    } finally {
      setResendLoading(false);
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
          Код подтверждения отправлен на {email}
        </Text>

        <Text style={authorizeStyle.label}>Код подтверждения</Text>
        <Input
          placeholder="000000"
          value={code}
          onChangeText={setCode}
          keyboardType="number-pad"
          maxLength={6}
          editable={!loading}
        />

        <Text style={authorizeStyle.label}>Новый пароль</Text>
        <Input
          placeholder="Минимум 6 символов"
          secureTextEntry
          value={newPassword}
          onChangeText={setNewPassword}
          editable={!loading}
        />

        <Text style={authorizeStyle.label}>Подтвердите пароль</Text>
        <Input
          placeholder="Повторите пароль"
          secureTextEntry
          value={confirmPassword}
          onChangeText={setConfirmPassword}
          editable={!loading}
        />

        <View style={[authorizeStyle.account, {justifyContent: 'space-between'}]}>
          <TouchableOpacity 
            onPress={handleResendCode}
            disabled={loading || resendLoading}
          >
            <Text style={{ color: resendLoading ? '#999' : '#000' }}>
              {resendLoading ? 'Отправка...' : 'Отправить код повторно'}
            </Text>
          </TouchableOpacity>
          
          <TouchableOpacity 
            onPress={() => navigation.goBack()}
            disabled={loading}
          >
            <Text>Изменить email</Text>
          </TouchableOpacity>
        </View>

        <Button
          title={loading ? 'Сохранение...' : 'Сбросить пароль'}
          onPress={handleResetPassword}
          disabled={loading || resendLoading}
        />

        {(loading || resendLoading) && (
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