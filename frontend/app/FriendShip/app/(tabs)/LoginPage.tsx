import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import { Platform, Text, TouchableOpacity, View } from 'react-native';
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
      contentContainerStyle={{ flexGrow: 1 }}
      enableOnAndroid
      keyboardShouldPersistTaps="handled"
      extraScrollHeight={Platform.OS === 'ios' ? 20 : 100}
    >
      <View style={[authorizeStyle.container, { flex: 1, justifyContent: 'space-between', paddingBottom: 30 }]}>
        <View>
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

        </View>
      </View>
      
      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </KeyboardAwareScrollView>
  );
};

export default Login;
