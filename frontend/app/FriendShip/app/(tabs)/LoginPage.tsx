import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import { Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Input from '../../components/auth/Input';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Login = () => {
  const navigation = useNavigation();
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');

  const handleLogin = () => {
    navigation.navigate('MainPage' as never);
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
          />

          <Text style={authorizeStyle.label}>Пароль</Text>
          <Input
            placeholder="Пароль"
            secureTextEntry
            value={password}
            onChangeText={setPassword}
          />

          <View style={authorizeStyle.account}>
            <TouchableOpacity onPress={() => navigation.navigate('Register' as never)}>
              <Text>Нет аккаунта?</Text>
            </TouchableOpacity>
          </View>

          <Button title="Войти" onPress={handleLogin} />
      </KeyboardAwareScrollView>

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </View>
  );
};

export default Login;
