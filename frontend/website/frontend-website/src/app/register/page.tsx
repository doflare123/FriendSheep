'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { registerSession } from '../../api/sessions';

import FormContainer from '../../components/FormContainer';
import FormInput from '../../components/FormInput';
import FormButton from '../../components/FormButton';
import FormLink from '../../components/FormLink';
import FormText from '../../components/FormText';
import LinkNote from '../../components/LinkNote';

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
        <FormContainer title="Регистрация" onSubmit={handleSubmit}>
        
            <FormInput
                id="username"
                label="Имя пользователя"
                placeholder="Имя пользователя"
                value={userName}
                onChange={(e) => setUserName(e.target.value)}
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

            <FormInput
                id="email"
                label="Почта"
                type="email"
                placeholder="user_email@gmail.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
            />

            <LinkNote>
                <FormLink href="/login" color="#000000">
                    Есть аккаунт?
                </FormLink>
            </LinkNote>

            <FormText>
                При создании аккаунта вы соглашаетесь с условиями{" "}
                <FormLink href="#">
                    Пользовательского соглашения
                </FormLink>{" "}
                и{" "}
                <FormLink href="#">
                    Политики конфиденциальности
                </FormLink>.
            </FormText>

            <FormButton type="submit">
                Зарегистрироваться
            </FormButton>
        
        </FormContainer>
    );
}

