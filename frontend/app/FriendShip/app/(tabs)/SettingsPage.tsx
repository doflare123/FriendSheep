import BottomBar from '@/components/BottomBar';
import ConfirmationModal from '@/components/ConfirmationModal';
import PageHeader from '@/components/PageHeader';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { useNavigation } from '@react-navigation/native';
import React, { useState } from 'react';
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

const GroupPage = () => {
  const { sortingState, sortingActions } = useSearchState();
  const navigation = useNavigation();
  const [showLogoutModal, setShowLogoutModal] = useState(false);

  const handleLogoutPress = () => {
    setShowLogoutModal(true);
  };

  const handleConfirmLogout = () => {
    setShowLogoutModal(false);
    navigation.navigate('Register' as never);
  };

  const handleCancelLogout = () => {
    setShowLogoutModal(false);
  };

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <PageHeader title="Настройки" showWave />
        <View style={styles.shadowWrapper}>
            <TouchableOpacity style={styles.badge} onPress={handleLogoutPress}>
                <Text style={styles.logout}>Выйти из аккаунта</Text>
                <Image
                    source={require('@/assets/images/settings/logout.png')}
                    style={styles.icon}
                />
            </TouchableOpacity>
        </View>
      </ScrollView>
      <BottomBar />

      <ConfirmationModal
        visible={showLogoutModal}
        title="Выход из аккаунта"
        message="Вы уверены, что хотите выйти из аккаунта?"
        onConfirm={handleConfirmLogout}
        onCancel={handleCancelLogout}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  scrollView: {
    flex: 1,
  },
  logout:{
    fontFamily: Montserrat.regular,
    fontSize: 20,
    color: Colors.red,
    flex: 1,
  },
  shadowWrapper:{
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 6,
  },
  badge:{
    padding: 16,
    backgroundColor: Colors.white,
    elevation: 2,
    borderRadius: 20,
    margin: 16,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center'
  },
  icon:{
    resizeMode: 'contain',
    width: 30,
    height: 30,
  }
});

export default GroupPage;