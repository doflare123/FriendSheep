'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { registerSession } from '../../api/register_sessions';
import { PageProtection, PAGE_KEYS } from '../../api/pageProtection';

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
    const [isLoading, setIsLoading] = useState(false);
    const [errors, setErrors] = useState({
        email: '',
        userName: '',
        password: '',
        general: ''
    });
    const router = useRouter();

    const validateForm = () => {
        const newErrors = {
            email: '',
            userName: '',
            password: '',
            general: ''
        };

        // Валидация email
        if (!email.trim()) {
            newErrors.email = 'Поле "Почта" обязательно для заполнения';
        } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
            newErrors.email = 'Введите корректный email адрес';
        }

        // Валидация имени пользователя
        if (!userName.trim()) {
            newErrors.userName = 'Поле "Имя пользователя" обязательно для заполнения';
        } else if (userName.trim().length < 2) {
            newErrors.userName = 'Имя пользователя должно содержать минимум 2 символа';
        }

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

        setErrors(newErrors);
        return !newErrors.email && !newErrors.userName && !newErrors.password;
    };

    const clearErrors = () => {
        setErrors({
            email: '',
            userName: '',
            password: '',
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
                `/register/confirm?username=${userName}&email=${email}&password=${password}&sessionId=${sessionId}`, 
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
                    id="username"
                    label="Имя пользователя"
                    placeholder="Имя пользователя"
                    value={userName}
                    onChange={handleUserNameChange}
                    required
                    disabled={isLoading}
                    className={errors.userName ? 'error' : ''}
                />
                {errors.userName && (
                    <span className="errorMessage">{errors.userName}</span>
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
                {errors.password && (
                    <span className="errorMessage">{errors.password}</span>
                )}
            </div>

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
                {errors.email && (
                    <span className="errorMessage">{errors.email}</span>
                )}
            </div>

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

            <FormButton type="submit" disabled={isLoading}>
                {isLoading ? 'Регистрируем...' : 'Зарегистрироваться'}
            </FormButton>
        
        </FormContainer>
    );
}