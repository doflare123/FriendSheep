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
  black: Platform.select({
    ios: 'Inter-Black',
    android: 'Inter-Black',
  }),
  medium: Platform.select({
    ios: 'Inter-Medium',
    android: 'Inter-Medium',
  }),
  bold_italic: Platform.select({
    ios: 'Inter-BoldItalic',
    android: 'Inter-BoldItalic',
  }),
};