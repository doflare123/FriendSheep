'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { login } from '../../api/login';
import { setCookie } from '../../api/auth'; // Импортируем утилиту
import FormContainer from '../../components/FormContainer';
import FormInput from '../../components/FormInput';
import FormButton from '../../components/FormButton';
import FormLink from '../../components/FormLink';
import LinkNote from '../../components/LinkNote';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  const saveTokens = (accessToken: string, refreshToken: string) => {
    // Access token в localStorage (короткое время жизни)
    localStorage.setItem('access_token', accessToken);
    
    // Refresh token в cookies (более безопасно, долгое время жизни)
    setCookie('refresh_token', refreshToken, 7); // 7 дней
    
    // Или оба в localStorage (проще в разработке)
    // localStorage.setItem('refresh_token', refreshToken);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    
    try {
      const response = await login(email, password);
      
      if (response.access_token && response.refresh_token) {
        saveTokens(response.access_token, response.refresh_token);
        
        console.log('Успешная авторизация');
        router.push('/');
        
        // Обновляем состояние приложения вместо перезагрузки
        window.dispatchEvent(new Event('auth-change'));
      } else {
        throw new Error('Токены не получены');
      }
    } catch (error: any) {
      console.error('Ошибка при авторизации:', error);
      
      let errorMessage = 'Что-то пошло не так при авторизации';
      
      if (error.response?.status === 401) {
        errorMessage = 'Неверный email или пароль';
      } else if (error.response?.status === 400) {
        errorMessage = 'Проверьте введенные данные';
      } else if (error.response?.status >= 500) {
        errorMessage = 'Ошибка сервера. Попробуйте позже';
      }
      
      alert(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <FormContainer title="Вход" onSubmit={handleSubmit} conteinerSize='max-w-xl'>
      <FormInput 
        id="email" 
        label="Почта" 
        type="email" 
        placeholder="user_email@gmail.com" 
        value={email} 
        onChange={(e) => setEmail(e.target.value)} 
        required 
        disabled={isLoading}
      />

      <FormInput 
        id="password" 
        label="Пароль" 
        type="password" 
        placeholder="Пароль" 
        value={password} 
        onChange={(e) => setPassword(e.target.value)} 
        required 
        disabled={isLoading}
      />

      <LinkNote>
        <FormLink href="/register" color="#000000">
          Нет аккаунта?
        </FormLink>
      </LinkNote>

      <FormButton type="submit" disabled={isLoading}>
        {isLoading ? 'Входим...' : 'Войти'}
      </FormButton>
    </FormContainer>
  );
}