import InputNums from '@/components/InputNums';
import { useNavigation } from '@react-navigation/native';
import React from 'react';
import { Platform, Text, View } from 'react-native';
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
    <KeyboardAwareScrollView
      contentContainerStyle={{ flexGrow: 1 }}
      enableOnAndroid
      keyboardShouldPersistTaps="handled"
      extraScrollHeight={Platform.OS === 'ios' ? 20 : 100}
    >
      <View style={[authorizeStyle.container, { flex: 1, justifyContent: 'space-between', paddingBottom: 30 }]}>
        <View>
        <Logo />
        <Text style={authorizeStyle.title}>Код подтверждения</Text>

        <PointAnimation />

        <Text style={[authorizeStyle.label, {fontSize: 16, textAlign: 'center'}]}>Введите код подтверждения из письма, отправленного на почту {user_email_name}@{user_email_domain}</Text>
        <InputNums/>

        <Text style={authorizeStyle.hintLabel}>Повторно код можно отправить через {time} мин</Text>

        <Button title="Подтвердить" onPress={() => navigation.navigate('Done' as never)} />
        </View>
      </View>

      <Text style={authorizeStyle.footer}>©NecroDwarf</Text>
    </KeyboardAwareScrollView>
  );
};

export default Confirm;
