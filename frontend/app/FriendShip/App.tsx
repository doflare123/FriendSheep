import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import React from 'react';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import CategoryPage from './app/(tabs)/CategoryPage';
import Confirm from './app/(tabs)/ConfirmPage';
import Done from './app/(tabs)/DonePage';
import GroupsPage from './app/(tabs)/GroupsPage';
import Login from './app/(tabs)/LoginPage';
import MainPage from './app/(tabs)/MainPage';
import Register from './app/(tabs)/RegisterPage';
import { RootStackParamList } from './navigation/types';


const Stack = createStackNavigator<RootStackParamList>();

export default function App() {
  return (
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
          <Stack.Screen name="Confirm" component={Confirm} />
          <Stack.Screen name="Done" component={Done} />
          <Stack.Screen name="MainPage" component={MainPage} />
          <Stack.Screen 
            name="CategoryPage" 
            component={CategoryPage}
            options={{ headerShown: false }}
          />
          <Stack.Screen name="GroupsPage" component={GroupsPage} />
        </Stack.Navigator>
      </NavigationContainer>
    </SafeAreaProvider>
  );
}
