'use client';

import { useEffect } from 'react';
import { initTheme } from '@/utils';

export default function ThemeInitializer() {
  useEffect(() => {
    initTheme();
  }, []);

  return null;
}