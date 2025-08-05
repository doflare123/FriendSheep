'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { login as loginAPI } from '../../api/login';
import { useAuth } from '../../contexts/AuthContext';
import FormContainer from '../../components/FormContainer';
import FormInput from '../../components/FormInput';
import FormButton from '../../components/FormButton';
import FormLink from '../../components/FormLink';
import LinkNote from '../../components/LinkNote';

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
  const { login } = useAuth(); // Используем функцию login из контекста

  const validateForm = () => {
    const newErrors = {
      email: false,
      password: false,
      general: ''
    };

    // Валидация email
    if (!email.trim()) {
      newErrors.email = true;
    }

    // Валидация пароля
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Очищаем предыдущие ошибки
    clearErrors();
    
    // Валидируем форму
    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    
    try {
      const response = await loginAPI(email, password);
      
      if (response.access_token && response.refresh_token) {
        // Используем функцию login из контекста вместо ручного сохранения
        login(response.access_token, response.refresh_token);
        
        console.log('Успешная авторизация');
        
        // Перенаправляем на главную страницу
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