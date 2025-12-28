import { Colors } from '@/constants/Colors';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as NavigationBar from 'expo-navigation-bar';
import React, { createContext, ReactNode, useContext, useEffect, useState } from 'react';
import { ActivityIndicator, Platform, StatusBar, StyleSheet, View } from 'react-native';

type Theme = 'light' | 'dark';

interface ThemeContextType {
  theme: Theme;
  toggleTheme: () => void;
  isDark: boolean;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

const THEME_STORAGE_KEY = '@app_theme';

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [theme, setTheme] = useState<Theme>('light');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadTheme();
  }, []);

  useEffect(() => {
    const isDark = theme === 'dark';
    const backgroundColor = isDark ? '#1A1A1A' : '#FFFFFF';

    StatusBar.setBarStyle(isDark ? 'light-content' : 'dark-content', true);
    StatusBar.setBackgroundColor(backgroundColor, true);
    
    if (Platform.OS === 'android') {
      NavigationBar.setBackgroundColorAsync(backgroundColor);
      NavigationBar.setButtonStyleAsync(isDark ? 'light' : 'dark');
    }
  }, [theme]);

  const loadTheme = async () => {
    try {
      const savedTheme = await AsyncStorage.getItem(THEME_STORAGE_KEY);
      if (savedTheme === 'dark' || savedTheme === 'light') {
        setTheme(savedTheme);
        
        const isDark = savedTheme === 'dark';
        const backgroundColor = isDark ? '#1A1A1A' : '#FFFFFF';
        
        StatusBar.setBarStyle(isDark ? 'light-content' : 'dark-content', false);
        StatusBar.setBackgroundColor(backgroundColor, false);
        
        if (Platform.OS === 'android') {
          NavigationBar.setBackgroundColorAsync(backgroundColor);
          NavigationBar.setButtonStyleAsync(isDark ? 'light' : 'dark');
        }
      }
    } catch (error) {
      console.error('[ThemeContext] Ошибка загрузки темы:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const toggleTheme = async () => {
    try {
      const newTheme: Theme = theme === 'light' ? 'dark' : 'light';
      setTheme(newTheme);
      await AsyncStorage.setItem(THEME_STORAGE_KEY, newTheme);
      console.log(`[ThemeContext] Тема изменена на: ${newTheme}`);
    } catch (error) {
      console.error('[ThemeContext] Ошибка сохранения темы:', error);
    }
  };

  if (isLoading) {
    return (
      <View style={[
        styles.loadingContainer, 
        { backgroundColor: theme === 'dark' ? '#1A1A1A' : '#FFFFFF' }
      ]}>
        <ActivityIndicator size="large" color={Colors.lightBlue} />
      </View>
    );
  }

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, isDark: theme === 'dark' }}>
      {children}
    </ThemeContext.Provider>
  );
};

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme должен использоваться внутри ThemeProvider');
  }
  return context;
};

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
});