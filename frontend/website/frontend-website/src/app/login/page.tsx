// src/app/login/page.tsx
'use client';

import { useState, useRef, useEffect } from 'react';
import { login } from '../../api/login';

import FormContainer from '../../components/FormContainer';
import FormInput from '../../components/FormInput';
import FormButton from '../../components/FormButton';
import FormLink from '../../components/FormLink';
import LinkNote from '../../components/LinkNote';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
          const responce = await login(email, password);
          if (responce.access_token && responce.refresh_token){
              localStorage.setItem('access_token', responce.access_token);
              localStorage.setItem('refresh_token', responce.refresh_token);
          }
      } catch (error) {
          console.error('Ошибка при авторизации:', error);
          alert('Что-то пошло не так при авторизации');
      }
    };

  return (
		<FormContainer title="Вход" onSubmit={handleSubmit}>
			
			<FormInput
				id="email"
				label="Почта"
				type="email"
				placeholder="user_email@gmail.com"
				value={email}
				onChange={(e) => setEmail(e.target.value)}
				required
			/>

			<FormInput
				id="password"
				label="Пароль"
				type="password"
				placeholder="Пароль"
				value={password}
				onChange={(e) => setPassword(e.target.value)}
				required
			/>

			<LinkNote>
				<FormLink href="#">
					Нет аккаунта?
				</FormLink>
			</LinkNote>

			<FormButton type="submit">
				Войти
			</FormButton>
			
		</FormContainer>
	);
}

