import { Platform } from 'react-native';

export const inter = {
  regular: Platform.select({
    ios: 'Inter-Regular',
    android: 'Inter-Regular',
  }),
  bold: Platform.select({
    ios: 'Inter-Bold', 
    android: 'Inter-Bold',
  }),
};