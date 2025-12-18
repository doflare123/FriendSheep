import { useTheme } from '@/components/ThemeContext';
import { getColors } from '@/constants/Colors';

export const useThemedColors = () => {
  const { isDark } = useTheme();
  return getColors(isDark);
};