import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import { Linking, Platform, Text, TouchableOpacity, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/Button';
import Input from '../../components/Input';
import Logo from '../../components/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Register = () => {
  const navigation = useNavigation();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');

  const openURL = (url: string) => {
    Linking.openURL(url).catch(err => console.error('Не удалось открыть ссылку:', err));
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
          <Text style={authorizeStyle.title}>Регистрация</Text>

          <Text style={authorizeStyle.label}>Имя пользователя</Text>
          <Input
            placeholder="Имя пользователя"
            value={username}
            onChangeText={setUsername}
          />

          <Text style={authorizeStyle.label}>Пароль</Text>
          <Input
            placeholder="Пароль"
            secureTextEntry
            value={password}
            onChangeText={setPassword}
          />

          <Text style={authorizeStyle.label}>Почта</Text>
          <Input
            placeholder="user_email@gmail.com"
            value={email}
            onChangeText={setEmail}
            keyboardType="email-address"
          />

          <View style={authorizeStyle.account}>
            <TouchableOpacity onPress={() => navigation.navigate('Login' as never)}>
              <Text>Есть аккаунт?</Text>
            </TouchableOpacity>
          </View>

          <Text style={authorizeStyle.terms}>
            При создании аккаунта вы соглашаетесь с условиями{' '}
            <Text style={authorizeStyle.link} onPress={() => openURL('https://example.com/terms')}>
              Пользовательского соглашения
            </Text>{' '}
            и{' '}
            <Text style={authorizeStyle.link} onPress={() => openURL('https://example.com/privacy')}>
              Политики конфиденциальности
            </Text>
          </Text>

        <Button title="Зарегистрироваться" onPress={() => navigation.navigate('Confirm' as never)} />
        </View>
      </View>
      
      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </KeyboardAwareScrollView>
  );
};

export default Register;
