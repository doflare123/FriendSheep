import React, { createContext, ReactNode, useContext, useState } from 'react';

interface HeaderContextType {
  hasPlayedAnimation: boolean;
  markAnimationPlayed: () => void;
  resetAnimation: () => void;
}

const HeaderContext = createContext<HeaderContextType | undefined>(undefined);

export const HeaderProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [hasPlayedAnimation, setHasPlayedAnimation] = useState(false);

  const markAnimationPlayed = () => {
    console.log('[HeaderContext] ✅ Анимация отмечена как проигранная');
    setHasPlayedAnimation(true);
  };

  const resetAnimation = () => {
    console.log('[HeaderContext] 🔄 Сброс анимации');
    setHasPlayedAnimation(false);
  };

  return (
    <HeaderContext.Provider value={{ hasPlayedAnimation, markAnimationPlayed, resetAnimation }}>
      {children}
    </HeaderContext.Provider>
  );
};

export const useHeaderAnimation = () => {
  const context = useContext(HeaderContext);
  if (!context) {
    throw new Error('useHeaderAnimation must be used within HeaderProvider');
  }
  return context;
};