// src/app/login/page.tsx
'use client';

import { useState, useRef, useEffect } from 'react';
import { login } from '../../api/login';

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
    <div className="page-wrapper">
      <main className="main-center">
        <div className="form-container">
          <h1 className="heading-title">Вход</h1>

          <form className="space-y-4" onSubmit={handleSubmit}>
            <div>
              <label className="label-style" htmlFor="email">
                Почта
              </label>
              <input
                id="email"
                className="input-style"
                placeholder="user_email@gmail.com"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>

            <div>
              <label className="label-style" htmlFor="password">
                Пароль
              </label>
              <input
                id="password"
                className="input-style"
                placeholder="Пароль"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>

            <div className="link-small-note">
              <a href="#" className="hover:underline">
                Нет аккаунта?
              </a>
            </div>

            <button
              className="button-primary"
              type="submit"
            >
              Войти
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}

