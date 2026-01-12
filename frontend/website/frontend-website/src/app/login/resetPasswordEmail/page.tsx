'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import FormContainer from '@/components/FormContainer';
import FormInput from '@/components/FormInput';
import FormButton from '@/components/FormButton';
import { showNotification } from '@/utils';
import { sendEmail } from '@/api/resetPswd/sendEmail';
import {checkDeviceAndRedirect} from '@/Constants';

export default function ResetPasswordEmailPage() {
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  
  const router = useRouter();

  useEffect(() => {
      checkDeviceAndRedirect(router);
  }, [router]);

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!email.trim()) {
      showNotification(400, 'Введите email', 'warning');
      return;
    }

    setIsLoading(true);
    
    try {
      const sessionId = await sendEmail(email);
      
      showNotification(200, 'Код отправлен на почту', 'success');
      
      router.push(`/login/resetPasswordCode?email=${email}&sessionId=${sessionId}`);
    } catch (error: any) {
      console.error('Ошибка при отправке email:', error);
      
      const statusCode = error?.response?.status || 500;
      
      const errorMessage = statusCode === 400 
        ? 'Такого аккаунта нет'
        : error?.response?.data?.message || error?.message || 'Ошибка при отправке кода';
      
      showNotification(statusCode, errorMessage, 'error');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <FormContainer title="Сброс пароля" onSubmit={handleSubmit} conteinerSize='max-w-xl'>
      <div className="flex flex-col" style={{ minHeight: '300px' }}>
        <div className="flex-grow">
          <FormInput 
            id="email" 
            label="Почта" 
            type="email" 
            placeholder="user_email@gmail.com" 
            value={email} 
            onChange={handleEmailChange}
            required 
            disabled={isLoading}
          />
        </div>

        <div className="mt-auto pt-4">
          <FormButton type="submit" disabled={isLoading}>
            {isLoading ? 'Отправляем...' : 'Отправить код'}
          </FormButton>
        </div>
      </div>
    </FormContainer>
  );
}