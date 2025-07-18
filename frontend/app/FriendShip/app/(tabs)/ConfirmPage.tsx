import InputSmall from '@/components/InputSmall';
import { useNavigation } from '@react-navigation/native';
import React from 'react';
import { Text, View } from 'react-native';
import { KeyboardAwareScrollView } from 'react-native-keyboard-aware-scroll-view';
import Button from '../../components/Button';
import Logo from '../../components/Logo';
import PointAnimation from '../../components/ui/PointAnimation';
import authorizeStyle from '../styles/authorizeStyle';

const Confirm = () => {
  const navigation = useNavigation();

  const handleConfirm = () => {
  };

  const user_email_name = "pochta";
  const user_email_domain = "gmail.com";
  var time = 15;

  return (
    <View style={authorizeStyle.container}>
      <View style={authorizeStyle.topContainer}>
        <Logo />
        <Text style={authorizeStyle.title}>Код подтверждения</Text>
        <PointAnimation />
      </View>

      <KeyboardAwareScrollView
        enableOnAndroid
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={authorizeStyle.scrollContent}
        extraScrollHeight={20}
        showsVerticalScrollIndicator={false}>

        <Text style={[authorizeStyle.label, {fontSize: 16, textAlign: 'center'}]}>Введите код подтверждения из письма, отправленного на почту {user_email_name}@{user_email_domain}</Text>
        <InputSmall/>
        
        <Text style={authorizeStyle.hintLabel}>Повторно код можно отправить через {time} мин</Text>
        
        <Button title="Подтвердить" onPress={() => navigation.navigate('Done' as never)} />
      </KeyboardAwareScrollView>

    <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </View>
  );
};

export default Confirm;
