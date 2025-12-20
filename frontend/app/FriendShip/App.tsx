import pushNotificationService from '@/api/services/pushNotificationService';
import ProfilePage from '@/app/(tabs)/ProfilePage';
import { ThemeProvider } from '@/components/ThemeContext';
import { ToastProvider } from '@/components/ToastContext';
import { Montserrat_200ExtraLight, Montserrat_300Light, Montserrat_400Regular, Montserrat_500Medium, Montserrat_600SemiBold, Montserrat_700Bold, useFonts } from '@expo-google-fonts/montserrat';
import { MontserratAlternates_500Medium } from '@expo-google-fonts/montserrat-alternates';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import * as Notifications from 'expo-notifications';
import * as SplashScreen from 'expo-splash-screen';
import React, { useEffect, useRef } from 'react';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import AllEventsPage from './app/(tabs)/AllEventsPage';
import AllGroupsPage from './app/(tabs)/AllGroupsPage';
import CategoryPage from './app/(tabs)/CategoryPage';
import Confirm from './app/(tabs)/ConfirmPage';
import Done from './app/(tabs)/DonePage';
import ForgotPassword from './app/(tabs)/ForgotPassword';
import GroupManagePage from './app/(tabs)/GroupManagePage';
import GroupPage from './app/(tabs)/GroupPage';
import GroupSearchPage from './app/(tabs)/GroupSearchPage';
import GroupsPage from './app/(tabs)/GroupsPage';
import Login from './app/(tabs)/LoginPage';
import MainPage from './app/(tabs)/MainPage';
import Register from './app/(tabs)/RegisterPage';
import ResetPassword from './app/(tabs)/ResetPassword';
import SettingsPage from './app/(tabs)/SettingsPage';
import UserSearchPage from './app/(tabs)/UserSearchPage';
import { AuthProvider, useAuthContext } from './components/auth/AuthContext';
import { RootStackParamList } from './navigation/types';

const Stack = createStackNavigator<RootStackParamList>();

SplashScreen.preventAutoHideAsync();

function RootNavigator() {
  const { isAuthenticated } = useAuthContext();
  const notificationListener = useRef<Notifications.Subscription | undefined>(undefined);
  const responseListener = useRef<Notifications.Subscription | undefined>(undefined);

  useEffect(() => {
    if (isAuthenticated) {
      initializePushNotifications();
    } else {
      notificationListener.current?.remove();
      responseListener.current?.remove();
    }

    return () => {
      notificationListener.current?.remove();
      responseListener.current?.remove();
    };
  }, [isAuthenticated]);

  const initializePushNotifications = async () => {
    try {
      console.log('[App] 🔔 Инициализация push-уведомлений...');

      await pushNotificationService.setupBackgroundHandler();

      const expoPushToken = await pushNotificationService.registerForPushNotifications();

      const fcmToken = await pushNotificationService.getFCMToken();

      if (expoPushToken) {
        await pushNotificationService.sendTokenToServer(expoPushToken, fcmToken);
      }

      notificationListener.current = pushNotificationService.addNotificationListener(
        (notification) => {
          console.log('📨 Получено уведомление:', notification);
        }
      );

      responseListener.current = pushNotificationService.addNotificationResponseListener(
        (response) => {
          console.log('👆 Нажатие на уведомление:', response);
          
          const data = response.notification.request.content.data;
          
          if (data.type === 'group_invite') {
            console.log('[App] Переход к приглашениям в группу');
          } else if (data.type === 'event_update') {
            console.log('[App] Переход к событию:', data.eventId);
          } else if (data.type === 'notification') {
            console.log('[App] Общее уведомление');
          }
        }
      );

      console.log('[App] ✅ Push-уведомления инициализированы успешно');
    } catch (error) {
      console.error('[App] ❌ Ошибка инициализации push-уведомлений:', error);
    }
  };

  return (
    <Stack.Navigator
      screenOptions={{
        gestureEnabled: false,
        headerShown: false,
        animation: 'scale_from_center',
        cardStyle: { backgroundColor: 'transparent' },
      }}
    >
      {isAuthenticated ? (
        <>
          <Stack.Screen name="MainPage" component={MainPage} />
          <Stack.Screen 
            name="CategoryPage" 
            component={CategoryPage}
            options={{ headerShown: false }}
          />
          <Stack.Screen name="GroupsPage" component={GroupsPage} />
          <Stack.Screen 
            name="GroupPage" 
            component={GroupPage}
            options={{ headerShown: false }}
          />
          <Stack.Screen 
            name="GroupManagePage" 
            component={GroupManagePage}
            options={{ headerShown: false }}
          />
          <Stack.Screen name="GroupSearchPage" component={GroupSearchPage} />
          <Stack.Screen name="ProfilePage" component={ProfilePage} />
          <Stack.Screen 
            name="UserSearchPage" 
            component={UserSearchPage}
            options={{ headerShown: false }}
          />
          <Stack.Screen name="SettingsPage" component={SettingsPage} />
          <Stack.Screen 
            name="AllEventsPage" 
            component={AllEventsPage}
            options={{ headerShown: false }}
          />
          <Stack.Screen name="AllGroupsPage" component={AllGroupsPage} />
        </>
      ) : (
        <>
          <Stack.Screen name="Login" component={Login} />
          <Stack.Screen name="Register" component={Register} />
          <Stack.Screen name="ForgotPassword" component={ForgotPassword} />
          <Stack.Screen name="ResetPassword" component={ResetPassword} />
          <Stack.Screen name="Confirm" component={Confirm} />
          <Stack.Screen name="Done" component={Done} />
        </>
      )}
    </Stack.Navigator>
  );
}

export default function App() {
  const [fontsLoaded] = useFonts({
    Montserrat_400Regular,
    Montserrat_700Bold,
    MontserratAlternates_500Medium,
    Montserrat_500Medium,
    Montserrat_600SemiBold,
    Montserrat_300Light,
    Montserrat_200ExtraLight
  });
  
  useEffect(() => {
    if (fontsLoaded) {
      SplashScreen.hideAsync();
    }
  }, [fontsLoaded]);

  if (!fontsLoaded) {
    return null;
  }

  return (
    <ThemeProvider>
      <AuthProvider>
        <ToastProvider>
          <SafeAreaProvider>
            <NavigationContainer>
              <RootNavigator />
            </NavigationContainer>
          </SafeAreaProvider>
        </ToastProvider>   
      </AuthProvider>
    </ThemeProvider>
  );
}