import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import React from 'react';

import Confirm from './app/(tabs)/ConfirmPage';
import Done from './app/(tabs)/DonePage';
import Login from './app/(tabs)/LoginPage';
import Register from './app/(tabs)/RegisterPage';


const Stack = createStackNavigator();

export default function App() {
  return (
    <NavigationContainer>
      <Stack.Navigator initialRouteName="Register" screenOptions={{ headerShown: false }}>
        <Stack.Screen name="Register" component={Register} />
        <Stack.Screen name="Login" component={Login} />
        <Stack.Screen name="Confirm" component={Confirm} />
        <Stack.Screen name="Done" component={Done} />
      </Stack.Navigator>
    </NavigationContainer>
  );
}
