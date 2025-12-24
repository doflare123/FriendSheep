import { getTokens, initializeTokenRefresh } from '@/api/storage/tokenStorage';
import { Colors } from '@/constants/Colors';
import React, { createContext, useContext, useEffect, useState } from 'react';
import { ActivityIndicator, View } from 'react-native';

interface TempRegistrationData {
  username: string;
  password: string;
  email: string;
}

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  setIsAuthenticated: (value: boolean) => void;

  tempRegData: TempRegistrationData | null;
  setTempRegData: (data: TempRegistrationData | null) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [tempRegData, setTempRegData] = useState<TempRegistrationData | null>(null);

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      console.log('[AuthProvider] Проверка статуса авторизации...');
      
      const tokens = await getTokens();
      
      if (tokens) {
        console.log('[AuthProvider] ✅ Пользователь авторизован');
        setIsAuthenticated(true);

        await initializeTokenRefresh();
      } else {
        console.log('[AuthProvider] ❌ Пользователь не авторизован');
        setIsAuthenticated(false);
      }
    } catch (error) {
      console.error('[AuthProvider] Ошибка проверки авторизации:', error);
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: Colors.white }}>
        <ActivityIndicator size="large" color={Colors.blue2} />
      </View>
    );
  }

  return (
    <AuthContext.Provider 
      value={{ 
        isAuthenticated, 
        isLoading, 
        setIsAuthenticated,
        tempRegData, 
        setTempRegData 
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuthContext = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuthContext must be used within AuthProvider');
  return context;
};

export const useAuth = useAuthContext;