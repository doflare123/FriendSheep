'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { registerSession } from '../../api/sessions';

export default function RegisterPage() {
    const [email, setEmail] = useState('');
    const [userName, setUserName] = useState('');
    const [password, setPassword] = useState('');
    const router = useRouter();

    const handleSubmit = async (e: React.FormEvent) => {
        const isValid =
            password.length >= 10 &&
            /[A-Z]/.test(password) &&
            /[a-z]/.test(password) &&
            /[^A-Za-z0-9]/.test(password);

        if (!isValid){
            alert("Вы ввели пароль неправильно!");
            return;
        }

        e.preventDefault();
        try {
            const sessionId = await registerSession(email);
            const expiry = Math.floor(Date.now() / 1000) + 10 * 60;
            localStorage.setItem('codeExpiryTime', expiry.toString());
            router.push(`/register/confirm?username=${userName}&email=${email}&password=${password}&sessionId=${sessionId}`);
        } catch (error) {
            console.error('Ошибка при регистрации:', error);
            alert('Что-то пошло не так при регистрации');
        }
    };

    return (
        <div className="page-wrapper">
            <main className="main-center">
                <div className="form-container">
                    <h1 className="heading-title">Регистрация</h1>

                    <form className="space-y-4" onSubmit={handleSubmit}>

                        <div>
                            <label className="label-style" htmlFor="username">
                                Имя пользователя
                            </label>
                            <input
                                id="username"
                                className="input-style"
                                placeholder="Имя пользователя"
                                value={userName}
                                onChange={(e) => setUserName(e.target.value)}
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

                        <div className="link-small-note">
                            <a href="#" className="hover:underline">
                                Есть аккаунт?
                            </a>
                        </div>

                        <div className="text-left text-sm text-gray-600 mt-2">
                            При создании аккаунта вы соглашаетесь с условиями{" "}
                            <a href="#" className="text-[#37A2E6] hover:underline">
                                Пользовательского соглашения
                            </a>{" "}
                            и{" "}
                            <a href="#" className="text-[#37A2E6] hover:underline">
                                Политики конфиденциальности
                            </a>.
                        </div>

                        <button
                            className="button-primary"
                            type="submit"
                            >
                            Зарегистрироваться
                        </button>
                    </form>
                </div>
            </main>
        </div>
    );
}

