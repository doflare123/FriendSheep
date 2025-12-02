import ProfilePage from '@/app/(tabs)/ProfilePage';
import { ToastProvider } from '@/components/ToastContext';
import { Montserrat_200ExtraLight, Montserrat_300Light, Montserrat_400Regular, Montserrat_500Medium, Montserrat_600SemiBold, Montserrat_700Bold, useFonts } from '@expo-google-fonts/montserrat';
import { MontserratAlternates_500Medium } from '@expo-google-fonts/montserrat-alternates';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import * as SplashScreen from 'expo-splash-screen';
import React, { useEffect } from 'react';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { AuthProvider } from './api/services/AuthContext';
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
import { RootStackParamList } from './navigation/types';

const Stack = createStackNavigator<RootStackParamList>();

SplashScreen.preventAutoHideAsync();

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
    <AuthProvider>
      <ToastProvider>
        <SafeAreaProvider>
            <NavigationContainer>
              <Stack.Navigator
                screenOptions={{
                  gestureEnabled: false,
                  headerShown: false,
                  animation: 'scale_from_center'
                }}
              > 
                <Stack.Screen name="Register" component={Register} />
                <Stack.Screen name="Login" component={Login} />
                <Stack.Screen name="ForgotPassword" component={ForgotPassword} />
                <Stack.Screen name="ResetPassword" component={ResetPassword} />
                <Stack.Screen name="Confirm" component={Confirm} />
                <Stack.Screen name="Done" component={Done} />
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
              </Stack.Navigator>
            </NavigationContainer>
        </SafeAreaProvider>
      </ToastProvider>   
    </AuthProvider>
  );
}
