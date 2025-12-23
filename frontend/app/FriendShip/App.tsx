import ProfilePage from '@/app/(tabs)/ProfilePage';
import { PushNotificationProvider } from '@/components/PushNotificationProvider';
import { ThemeProvider, useTheme } from '@/components/ThemeContext';
import { ToastProvider } from '@/components/ToastContext';
import { Montserrat_200ExtraLight, Montserrat_300Light, Montserrat_400Regular, Montserrat_500Medium, Montserrat_600SemiBold, Montserrat_700Bold, useFonts } from '@expo-google-fonts/montserrat';
import { MontserratAlternates_500Medium } from '@expo-google-fonts/montserrat-alternates';
import messaging from '@react-native-firebase/messaging';
import { DarkTheme, DefaultTheme, NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import * as SplashScreen from 'expo-splash-screen';
import React, { useEffect } from 'react';
import { StyleSheet, View } from 'react-native';
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

messaging().setBackgroundMessageHandler(async (remoteMessage) => {
  console.log('[Background] Получено уведомление:', remoteMessage);
});

function RootNavigator() {
  const { isAuthenticated } = useAuthContext();
  const { isDark } = useTheme();

  const backgroundColor = isDark ? '#1A1A1A' : '#FFFFFF';

  const customCardStyleInterpolator = ({ current, layouts }: any) => {
    return {
      cardStyle: {
        opacity: current.progress.interpolate({
          inputRange: [0, 0.5, 1],
          outputRange: [0, 0.5, 1],
        }),
        transform: [
          {
            scale: current.progress.interpolate({
              inputRange: [0, 1],
              outputRange: [0.95, 1],
            }),
          },
        ],
        backgroundColor,
      },
    };
  };

  return (
    <Stack.Navigator
      screenOptions={{
        gestureEnabled: true,
        headerShown: false,
        cardStyleInterpolator: customCardStyleInterpolator,
        transitionSpec: {
          open: {
            animation: 'timing',
            config: {
              duration: 300,
            },
          },
          close: {
            animation: 'timing',
            config: {
              duration: 300,
            },
          },
        },
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
            <ThemedNavigationContainer />
          </SafeAreaProvider>
        </ToastProvider>   
      </AuthProvider>
    </ThemeProvider>
  );
}

function ThemedNavigationContainer() {
  const { isDark } = useTheme();

  const lightTheme = {
    ...DefaultTheme,
    colors: {
      ...DefaultTheme.colors,
      background: '#FFFFFF',
      card: '#FFFFFF',
      text: '#000000',
      border: '#E2EAEF',
      primary: '#408DD2',
    },
  };

  const darkTheme = {
    ...DarkTheme,
    colors: {
      ...DarkTheme.colors,
      background: '#1A1A1A',
      card: '#2F2F2F',
      text: '#EDEDED',
      border: '#3A3A3A',
      primary: '#408DD2',
    },
  };

  return (
    <View style={[styles.container, { backgroundColor: isDark ? '#1A1A1A' : '#FFFFFF' }]}>
      <NavigationContainer theme={isDark ? darkTheme : lightTheme}>
        <AuthWrapper />
      </NavigationContainer>
    </View>
  );
}

function AuthWrapper() {
  const { isAuthenticated } = useAuthContext();
  
  return (
    <PushNotificationProvider isAuthenticated={isAuthenticated}>
      <RootNavigator />
    </PushNotificationProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
});