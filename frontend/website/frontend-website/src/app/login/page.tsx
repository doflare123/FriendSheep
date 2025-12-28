'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { login as loginAPI } from '../../api/login';
import { useAuth } from '../../contexts/AuthContext';
import FormContainer from '../../components/FormContainer';
import FormInput from '../../components/FormInput';
import FormButton from '../../components/FormButton';
import FormLink from '../../components/FormLink';
import LinkNote from '../../components/LinkNote';
import {updateUserData, checkDeviceAndRedirect} from '@/Constants'
import styles from '@/styles/resetPassword.module.css';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState({
    email: false,
    password: false,
    general: ''
  });
  
  const router = useRouter();
  const { login  } = useAuth();

  useEffect(() => {
      checkDeviceAndRedirect(router);
  }, [router]);

  const validateForm = () => {
    const newErrors = {
      email: false,
      password: false,
      general: ''
    };

    if (!email.trim()) {
      newErrors.email = true;
    }

    if (!password.trim()) {
      newErrors.password = true;
    }

    setErrors(newErrors);
    return !newErrors.email && !newErrors.password;
  };

  const clearErrors = () => {
    setErrors({
      email: false,
      password: false,
      general: ''
    });
  };

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    if (errors.general) {
      setErrors(prev => ({ ...prev, general: '' }));
    }
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
    if (errors.general) {
      setErrors(prev => ({ ...prev, general: '' }));
    }
  };

  const handleForgotPassword = (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    router.push('/login/resetPasswordEmail/');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    clearErrors();
    
    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    
    try {
      const response = await loginAPI(email, password);
      
      if (response.access_token && response.refresh_token) {
        login(response.access_token, response.refresh_token);
        
        updateUserData();
        router.push('/');
      } else {
        throw new Error('Токены не получены');
      }
    } catch (error: any) {
      console.error('Ошибка при авторизации:', error);
      
      let errorMessage = 'Что-то пошло не так при авторизации';
      
      if (error.status === 401 || error.status === 404) {
        errorMessage = 'Неверный email или пароль';
      } else if (error.status === 400) {
        errorMessage = 'Проверьте введенные данные';
      } else if (error.status >= 500) {
        errorMessage = 'Ошибка сервера. Попробуйте позже';
      }
      
      setErrors(prev => ({ ...prev, general: errorMessage }));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <FormContainer title="Вход" onSubmit={handleSubmit} conteinerSize='max-w-xl'>

      <div>
        <FormInput 
          id="email" 
          label="Почта" 
          type="email" 
          placeholder="user_email@gmail.com" 
          value={email} 
          onChange={handleEmailChange}
          required 
          disabled={isLoading}
          className={errors.email ? 'error' : ''}
        />
        {errors.general && (
          <span className="errorMessage">{errors.general}</span>
        )}
      </div>

      <div>
        <FormInput 
          id="password" 
          label="Пароль" 
          type="password" 
          placeholder="Пароль" 
          value={password} 
          onChange={handlePasswordChange}
          required 
          disabled={isLoading}
          className={errors.password ? 'error' : ''}
        />
        {errors.general && (
          <span className="errorMessage">{errors.general}</span>
        )}
      </div>

      <LinkNote leftLink={
        <FormLink onClick={handleForgotPassword}>
          Забыли пароль?
        </FormLink>
      }>
        <FormLink href="/register" color="var(--color-text-primary)">
          Нет аккаунта?
        </FormLink>
      </LinkNote>

      <FormButton type="submit" disabled={isLoading}>
        {isLoading ? 'Входим...' : 'Войти'}
      </FormButton>
    </FormContainer>
  );
}