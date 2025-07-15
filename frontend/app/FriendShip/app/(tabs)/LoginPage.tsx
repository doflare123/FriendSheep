import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import { Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/Button';
import Input from '../../components/Input';
import Logo from '../../components/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Login = () => {
  const navigation = useNavigation();
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');

  const handleLogin = () => {
  };

  return (
    <KeyboardAwareScrollView
      contentContainerStyle={authorizeStyle.container}
      enableOnAndroid
      keyboardShouldPersistTaps="always"
    >
      <Logo />
      <Text style={authorizeStyle.title}>Вход</Text>

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

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </KeyboardAwareScrollView>
  );
};

export default Login;
