import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import React from 'react';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import Confirm from './app/(tabs)/ConfirmPage';
import Done from './app/(tabs)/DonePage';
import Login from './app/(tabs)/LoginPage';
import MainPage from './app/(tabs)/MainPage';
import Register from './app/(tabs)/RegisterPage';


const Stack = createStackNavigator();

export default function App() {
  return (
    <SafeAreaProvider>
      <NavigationContainer>
        <Stack.Navigator initialRouteName="Register" screenOptions={{ headerShown: false }}>
          <Stack.Screen name="Register" component={Register} />
          <Stack.Screen name="Login" component={Login} />
          <Stack.Screen name="Confirm" component={Confirm} />
          <Stack.Screen name="Done" component={Done} />
          <Stack.Screen name="MainPage" component={MainPage} />
        </Stack.Navigator>
      </NavigationContainer>
    </SafeAreaProvider>
  );
}
