'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import styles from '@/styles/ThemeToggle.module.css';

export default function ThemeToggle() {
  const [theme, setTheme] = useState<'light' | 'dark'>('light');

  useEffect(() => {
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark' | null;
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const initialTheme = savedTheme || (prefersDark ? 'dark' : 'light');
    setTheme(initialTheme);
  }, []);

  const toggleTheme = () => {
    const newTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
    localStorage.setItem('theme', newTheme);
    document.documentElement.setAttribute('data-theme', newTheme);
  };

  return (
    <button
      className={styles.themeToggle}
      onClick={toggleTheme}
      title={theme === 'light' ? 'Переключить на тёмную тему (в разработке, могут быть баги)' : 'Переключить на светлую тему'}
      aria-label="Переключить тему"
    >
      <Image
        src={theme === 'light' ? '/icons/dark_theme.png' : '/icons/light_theme.png'}
        alt="Переключить тему"
        width={48}
        height={48}
        priority
      />
    </button>
  );
}