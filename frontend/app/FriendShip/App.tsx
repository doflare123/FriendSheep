import React from 'react';
import { SafeAreaView, StatusBar } from 'react-native';
import Register from './app/(tabs)/Register';

export default function App() {
  return (
    <SafeAreaView style={{ flex: 1 }}>
      <StatusBar barStyle="dark-content" />
      <Register />
    </SafeAreaView>
  );
}
