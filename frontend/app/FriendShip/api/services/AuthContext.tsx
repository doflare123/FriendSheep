import React, { createContext, useContext, useState } from 'react';

interface TempRegistrationData {
  username: string;
  password: string;
  email: string;
}

interface AuthContextType {
  tempRegData: TempRegistrationData | null;
  setTempRegData: (data: TempRegistrationData | null) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [tempRegData, setTempRegData] = useState<TempRegistrationData | null>(null);

  return (
    <AuthContext.Provider value={{ tempRegData, setTempRegData }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuthContext = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuthContext must be used within AuthProvider');
  return context;
};