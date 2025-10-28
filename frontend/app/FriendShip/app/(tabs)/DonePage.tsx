import { useNavigation } from '@react-navigation/native';
import React from 'react';
import { Image, Text, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/auth/Button';
import Logo from '../../components/auth/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Done = () => {
  const navigation = useNavigation();

  return (
    <View style={authorizeStyle.container}>

      <View style={authorizeStyle.topContainer}>
        <Logo />
        <Text style={authorizeStyle.title}>Аккаунт создан!</Text>
        <Image style={authorizeStyle.doneImage} source={require('../../assets/images/done.png')} />
      </View>

      <KeyboardAwareScrollView
        enableOnAndroid
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={authorizeStyle.scrollContent}
        extraScrollHeight={20}
        showsVerticalScrollIndicator={false}>

        <Text style={[authorizeStyle.label, {fontSize: 16, textAlign: 'center', marginBottom: 40}]}>Ваш аккаунт успешно подтвержден и активирован. Теперь вы можете приступить к пользованию сервисом.{"\n\n"}Спасибо, что выбрали нас!</Text>

        <Button title="Перейти ко входу" onPress={() => navigation.navigate('Login' as never)} />
      </KeyboardAwareScrollView>

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </View>
  );
};

export default Done;
