'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { registerSession } from '../../api/register_sessions';
import { PageProtection, PAGE_KEYS } from '../../api/pageProtection';

import FormContainer from '../../components/FormContainer';
import FormInput from '../../components/FormInput';
import FormButton from '../../components/FormButton';
import FormLink from '../../components/FormLink';
import FormText from '../../components/FormText';
import LinkNote from '../../components/LinkNote';

import {filterProfanity} from '@/utils';

import {checkDeviceAndRedirect} from '@/Constants';

export default function RegisterPage() {
    const [email, setEmail] = useState('');
    const [userName, setUserName] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [errors, setErrors] = useState({
        email: '',
        userName: '',
        password: '',
        confirmPassword: '',
        general: ''
    });
    const router = useRouter();

    useEffect(() => {
        checkDeviceAndRedirect(router);
    }, [router]);

    const validateForm = () => {
        const newErrors = {
            email: '',
            userName: '',
            password: '',
            confirmPassword: '',
            general: ''
        };

        // Валидация email
        if (!email.trim()) {
            newErrors.email = 'Поле "Почта" обязательно для заполнения';
        } else if (!email.includes('@') || !email.includes('.')) {
            newErrors.email = 'Email должен содержать символы @ и .';
        } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
            newErrors.email = 'Введите корректный email адрес';
        }

        // Валидация имени пользователя
        if (!userName.trim()) {
            newErrors.userName = 'Поле "Имя пользователя" обязательно для заполнения';
        } else if (userName.trim().length < 5) {
            newErrors.userName = 'Имя пользователя должно содержать минимум 5 символов';
        } else if (userName.trim().length > 40) {
            newErrors.userName = 'Имя пользователя должно содержать максимум 40 символов';
        }

        // Валидация пароля
        if (!password.trim()) {
            newErrors.password = 'Поле "Пароль" обязательно для заполнения';
        } else if (/[А-Яа-яЁё]/.test(password)) {
            newErrors.password = 'Пароль не должен содержать русские буквы';
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
            newErrors.confirmPassword = 'Поле "Подтверждение пароля" обязательно для заполнения';
        } else if (password !== confirmPassword) {
            newErrors.confirmPassword = 'Пароли не совпадают';
        }

        setErrors(newErrors);
        return !newErrors.email && !newErrors.userName && !newErrors.password && !newErrors.confirmPassword;
    };

    const clearErrors = () => {
        setErrors({
            email: '',
            userName: '',
            password: '',
            confirmPassword: '',
            general: ''
        });
    };

    const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setEmail(e.target.value);
        if (errors.email) {
            setErrors(prev => ({ ...prev, email: '' }));
        }
    };

    const handleUserNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setUserName(e.target.value);
        if (errors.userName) {
            setErrors(prev => ({ ...prev, userName: '' }));
        }
    };

    const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setPassword(e.target.value);
        if (errors.password) {
            setErrors(prev => ({ ...prev, password: '' }));
        }
    };

    const handleConfirmPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setConfirmPassword(e.target.value);
        if (errors.confirmPassword) {
            setErrors(prev => ({ ...prev, confirmPassword: '' }));
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
            const sessionId = await registerSession(email);
            const expiry = Math.floor(Date.now() / 1000) + 10 * 60;
            localStorage.setItem('codeExpiryTime', expiry.toString());
            PageProtection.navigateWithAccess(
                router, 
                `/register/confirm?username=${filterProfanity(userName)}&email=${email}&password=${password}&sessionId=${sessionId}`, 
                PAGE_KEYS.CODE_VERIFY,
                180
            );
        } catch (error: any) {
            console.error('Ошибка при регистрации:', error);
            
            let errorMessage = 'Что-то пошло не так при регистрации';
            
            if (error.response?.status === 400) {
                errorMessage = 'Проверьте введенные данные';
            } else if (error.response?.status >= 500) {
                errorMessage = 'Ошибка сервера. Попробуйте позже';
            }
            
            setErrors(prev => ({ ...prev, general: errorMessage }));
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <FormContainer title="Регистрация" onSubmit={handleSubmit}>
            {errors.general && (
                <div className="errorMessage" style={{ marginBottom: '16px', textAlign: 'center' }}>
                    {errors.general}
                </div>
            )}
        
            <div>
                <FormInput
                    id="email"
                    name="email"
                    label="Почта"
                    type="email"
                    placeholder="user_email@gmail.com"
                    value={email}
                    onChange={handleEmailChange}
                    autoComplete="email"
                    required
                    disabled={isLoading}
                    className={errors.email ? 'error' : ''}
                />
                {errors.email && (
                    <span className="errorMessage">{errors.email}</span>
                )}
            </div>

            <div>
                <FormInput
                    id="password"
                    name="password"
                    label="Пароль"
                    type="password"
                    placeholder="Пароль"
                    value={password}
                    onChange={handlePasswordChange}
                    autoComplete="new-password"
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
                    name="confirm-password"
                    label="Подтверждение пароля"
                    type="password"
                    placeholder="Подтвердите пароль"
                    value={confirmPassword}
                    onChange={handleConfirmPasswordChange}
                    autoComplete="new-password"
                    required
                    disabled={isLoading}
                    className={errors.confirmPassword ? 'error' : ''}
                />
                {errors.confirmPassword && (
                    <span className="errorMessage">{errors.confirmPassword}</span>
                )}
            </div>

            <div>
                <FormInput
                    id="username"
                    name="username"
                    label="Имя пользователя"
                    placeholder="Имя пользователя"
                    value={userName}
                    onChange={handleUserNameChange}
                    autoComplete="off"
                    required
                    disabled={isLoading}
                    className={errors.userName ? 'error' : ''}
                />
                {errors.userName && (
                    <span className="errorMessage">{errors.userName}</span>
                )}
            </div>

            <LinkNote>
                <FormLink href="/login" color="var(--color-text-primary)">
                    Есть аккаунт?
                </FormLink>
            </LinkNote>

            <FormText>
                При создании аккаунта вы соглашаетесь с условиями{" "}
                <FormLink href="/info/agreement">
                    Пользовательского соглашения
                </FormLink>{" "}
                и{" "}
                <FormLink href="/info/privacy">
                    Политики конфиденциальности
                </FormLink>.
            </FormText>

            <FormButton type="submit" disabled={isLoading}>
                {isLoading ? 'Регистрируем...' : 'Зарегистрироваться'}
            </FormButton>
        
        </FormContainer>
    );
}