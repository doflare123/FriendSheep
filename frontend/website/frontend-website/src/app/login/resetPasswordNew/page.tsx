'use client';

export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import FormContainer from '@/components/FormContainer';
import FormInput from '@/components/FormInput';
import FormButton from '@/components/FormButton';
import { showNotification } from '@/utils';
import { resetPassword } from '@/api/resetPassword';
import {checkDeviceAndRedirect} from '@/Constants';

export default function ResetPasswordNewPage() {
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState({
    password: '',
    confirmPassword: '',
    general: ''
  });
  
  const router = useRouter();
  const searchParams = useSearchParams();

  useEffect(() => {
      checkDeviceAndRedirect(router);
  }, [router]);
  
  const email = searchParams.get('email') || '';
  const sessionId = searchParams.get('sessionId') || '';

  const validateForm = () => {
    const newErrors = {
      password: '',
      confirmPassword: '',
      general: ''
    };

    // Валидация пароля
    if (!password.trim()) {
      newErrors.password = 'Поле "Пароль" обязательно для заполнения';
    } else {
      const isValidPassword =
        password.length >= 10 &&
        /[A-Z]/.test(password) &&
        /[a-z]/.test(password) &&
        /[^A-Za-z0-9]/.test(password);

      if (!isValidPassword) {
        newErrors.password = 'Пароль должен содержать минимум 10 символов, включая заглавные и строчные буквы, а также специальные символы';
      }
    }

    // Валидация подтверждения пароля
    if (!confirmPassword.trim()) {
      newErrors.confirmPassword = 'Подтвердите пароль';
    } else if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Пароли не совпадают';
    }

    setErrors(newErrors);
    return !newErrors.password && !newErrors.confirmPassword;
  };

  const clearErrors = () => {
    setErrors({
      password: '',
      confirmPassword: '',
      general: ''
    });
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
    if (errors.password || errors.general) {
      setErrors(prev => ({ ...prev, password: '', general: '' }));
    }
  };

  const handleConfirmPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setConfirmPassword(e.target.value);
    if (errors.confirmPassword || errors.general) {
      setErrors(prev => ({ ...prev, confirmPassword: '', general: '' }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    clearErrors();
    
    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    
    try {
      await resetPassword(email, password, sessionId);
      
      showNotification(200, 'Пароль успешно изменен', 'success');
      
      // Небольшая задержка для отображения уведомления
      setTimeout(() => {
        console.log("LOGIN5");
        router.push('/login');
      }, 1000);
    } catch (error: any) {
      console.error('Ошибка при сбросе пароля:', error);
      
      const errorMessage = error?.response?.data?.message || 
                          error?.message || 
                          'Ошибка при изменении пароля';
      
      const statusCode = error?.response?.status || 500;
      
      showNotification(statusCode, errorMessage, 'error');
      setErrors(prev => ({ ...prev, general: errorMessage }));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <FormContainer title="Сброс пароля" onSubmit={handleSubmit} conteinerSize='max-w-xl'>
      <div>
        <FormInput 
          id="password" 
          label="Пароль" 
          type="password" 
          placeholder="Новый пароль" 
          value={password} 
          onChange={handlePasswordChange}
          required 
          disabled={isLoading}
          className={errors.password ? 'error' : ''}
        />
        {errors.password && (
          <span className="errorMessage">{errors.password}</span>
        )}
      </div>

      <div>
        <FormInput 
          id="confirmPassword" 
          label="Повторите пароль" 
          type="password" 
          placeholder="Повторите новый пароль" 
          value={confirmPassword} 
          onChange={handleConfirmPasswordChange}
          required 
          disabled={isLoading}
          className={errors.confirmPassword ? 'error' : ''}
        />
        {errors.confirmPassword && (
          <span className="errorMessage">{errors.confirmPassword}</span>
        )}
      </div>

      {errors.general && (
        <div className="errorMessage" style={{ textAlign: 'center' }}>
          {errors.general}
        </div>
      )}

      <FormButton type="submit" disabled={isLoading}>
        {isLoading ? 'Сохраняем...' : 'Сменить'}
      </FormButton>
    </FormContainer>
  );
}