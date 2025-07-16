import { useNavigation } from '@react-navigation/native';
import React from 'react';
import { Image, Platform, Text, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/Button';
import Logo from '../../components/Logo';
import authorizeStyle from '../styles/authorizeStyle';

const Done = () => {
  const navigation = useNavigation();

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

          <Text style={[authorizeStyle.title, {marginBottom: 20}]}>Аккаунт создан!</Text>
          
          <Image style={authorizeStyle.doneImage} source={require('../../assets/images/done.png')} />

          <Text style={[authorizeStyle.label, {fontSize: 18, textAlign: 'center', marginBottom: 40}]}>Ваш аккаунт успешно подтвержден и активирован. Теперь вы можете приступить к пользованию сервисом.{"\n\n"}Спасибо, что выбрали нас!</Text>

          <Button title="Перейти ко входу" onPress={() => navigation.navigate('Login' as never)} />
        </View>
      </View>

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </KeyboardAwareScrollView>
  );
};

export default Done;
