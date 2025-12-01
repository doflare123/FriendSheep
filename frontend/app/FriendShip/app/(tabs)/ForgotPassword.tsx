import authService from '@/api/services/authService';
import { useToast } from '@/components/ToastContext';
import { Colors } from '@/constants/Colors';
import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import { ActivityIndicator, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Input from '../../components/auth/Input';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const ForgotPassword = () => {
  const navigation = useNavigation();
  const { showToast } = useToast();
  
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleRequestReset = async () => {
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

    setLoading(true);

    try {
      const response = await authService.requestPasswordReset(sanitizedEmail);

      showToast({
        type: 'success',
        title: 'Успешно!',
        message: 'Код для сброса пароля отправлен на почту',
      });

      (navigation.navigate as any)('ResetPassword', {
        sessionId: response.session_id,
        email: sanitizedEmail,
      });
    } catch (error: any) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: error.message || 'Не удалось отправить код',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={authorizeStyle.container}>
      <View style={authorizeStyle.topContainer}>
        <Logo />
        <Text style={authorizeStyle.title}>Восстановление пароля</Text>
      </View>

      <KeyboardAwareScrollView
        enableOnAndroid
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={authorizeStyle.scrollContent}
        extraScrollHeight={20}
        showsVerticalScrollIndicator={false}>

        <Text style={authorizeStyle.label}>Email</Text>
        <Input
          placeholder="user_email@gmail.com"
          value={email}
          onChangeText={setEmail}
          keyboardType="email-address"
          autoCapitalize="none"
          editable={!loading}
        />

        <Text style={authorizeStyle.terms}>
          Мы отправим код подтверждения на указанный email.
          После этого вы сможете установить новый пароль
        </Text>

        <Button
          title={loading ? 'Отправка...' : 'Отправить код'}
          onPress={handleRequestReset}
          disabled={loading}
        />

        {loading && (
          <ActivityIndicator 
            size="large" 
            color={Colors.blue}
            style={{ marginTop: 16 }} 
          />
        )}

        <View style={[authorizeStyle.account, {marginTop: 20}]}>
          <TouchableOpacity 
            onPress={() => navigation.goBack()}
            disabled={loading}
          >
            <Text style={authorizeStyle.account}>Вернуться к входу</Text>
          </TouchableOpacity>
        </View>
      </KeyboardAwareScrollView>

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </View>
  );
};

export default ForgotPassword;