'use client';

import { useState, useRef, useEffect } from 'react';
import "../../../styles/ConfirmCode.css";
import { useSearchParams, useRouter } from 'next/navigation';
import { registerSession } from '../../../api/register_sessions';
import { confirm_code } from '../../../api/confirm_code';
import { createUser } from '../../../api/create_user';
import { usePageProtection, PAGE_KEYS } from '../../../api/pageProtection';
import { showNotification } from "@/utils";

import FormButton from '../../../components/FormButton';
import VerifiContainer from '../../../components/VerifiContainer';

export default function ConfirmCode() {
    // usePageProtection({ 
    //     pageKey: PAGE_KEYS.CODE_VERIFY,
    //     redirectTo: '/register'
    // });

    const [values, setValues] = useState(['', '', '', '', '', '']);
    const inputsRef = useRef<(HTMLInputElement | null)[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [errors, setErrors] = useState({
        code: '',
        general: ''
    });

    const [timeLeft, setTimeLeft] = useState(1 * 60);
    const [resendAvailable, setResendAvailable] = useState(false);

    const searchParams = useSearchParams();
    const router = useRouter();

    const email: string = searchParams.get('email') || "";
    const userName: string = searchParams.get('username') || "";
    const password: string = searchParams.get('password') || "";
    const [sessionId, setSessionId] = useState<string>(searchParams.get('sessionId')  || "");


    
    useEffect(() => {
        const storedExpiry = localStorage.getItem('codeExpiryTime');
        if (storedExpiry) {
            const expiry = parseInt(storedExpiry, 10);
            const now = Math.floor(Date.now() / 1000);
            const remaining = expiry - now;
            if (remaining > 0) {
                setTimeLeft(remaining);
            } else {
                setTimeLeft(0);
                setResendAvailable(true);
            }
        } else {
            const expiry = Math.floor(Date.now() / 1000) + 1 * 60;
            localStorage.setItem('codeExpiryTime', expiry.toString());
        }
    }, []);

    useEffect(() => {
        if (timeLeft <= 0) {
            setResendAvailable(true);
            localStorage.removeItem('codeExpiryTime');
            return;
        }

        const interval = setInterval(() => {
            setTimeLeft((prev) => prev - 1);
        }, 1000);

        return () => clearInterval(interval);
    }, [timeLeft]);

    const clearErrors = () => {
        setErrors({
            code: '',
            general: ''
        });
    };

    const validateCode = () => {
        const code = values.join('');
        if (code.length !== 6) {
            setErrors(prev => ({ ...prev, code: 'Введите полный 6-значный код' }));
            return false;
        }
        if (!/^[a-zA-Z0-9]{6}$/.test(code)) {
            setErrors(prev => ({ ...prev, code: 'Код должен содержать только буквы и цифры' }));
            return false;
        }
        return true;
    };

    const handleResendCode = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        clearErrors();

        try {
            const newSession = await registerSession(email);
            setSessionId(newSession);
            
            const newExpiry = Math.floor(Date.now() / 1000) + 10 * 60;
            localStorage.setItem('codeExpiryTime', newExpiry.toString());
            setTimeLeft(10 * 60); 
            setResendAvailable(false);
            
            // Очищаем введенный код
            setValues(['', '', '', '', '', '']);
            inputsRef.current[0]?.focus();
            
        } catch (error: any) {
            console.error('Ошибка при отправке кода:', error);
            
            let errorMessage = 'Что-то пошло не так при повторной отправке кода';
            
            if (error.response?.status === 429) {
                errorMessage = 'Слишком много попыток. Попробуйте позже';
            } else if (error.response?.status >= 500) {
                errorMessage = 'Ошибка сервера. Попробуйте позже';
            }
            
            showNotification(error.status, errorMessage);
        } finally {
            setIsLoading(false);
        }
    };

    const formatTime = () => {
        if (timeLeft >= 60) {
            const minutes = Math.floor(timeLeft / 60);
            const seconds = timeLeft - minutes*60
            if (seconds > 0)
                return `Повторно код можно отправить через ${minutes} мин. ${seconds} сек.`;
            else
                return `Повторно код можно отправить через ${minutes} мин.`;
        } else {
            return `Повторно код можно отправить через ${timeLeft} сек.`;
        }
    };

    const handleChange = (index: number, value: string) => {
        // Разрешаем буквы и цифры (алфавитно-цифровые символы)
        if (value && !/^[a-zA-Z0-9]$/.test(value)) return;
        
        const newValues = [...values];
        newValues[index] = value.toUpperCase(); // Приводим к верхнему регистру для единообразия
        setValues(newValues);
        
        // Очищаем ошибку при вводе
        if (errors.code) {
            setErrors(prev => ({ ...prev, code: '' }));
        }
        
        if (value && index < values.length - 1) {
            inputsRef.current[index + 1]?.focus();
        }
    };

    const handleKeyDown = (index: number, event: React.KeyboardEvent<HTMLInputElement>) => {
        if (event.key === 'Backspace' && !values[index] && index > 0) {
            inputsRef.current[index - 1]?.focus();
        }
    };

    const handlePaste = (event: React.ClipboardEvent<HTMLInputElement>) => {
        event.preventDefault();
        const paste = event.clipboardData.getData('text').trim().slice(0, values.length);
        if (!paste) return;

        // Проверяем, что вставляемый текст содержит только буквы и цифры
        if (!/^[a-zA-Z0-9]+$/.test(paste)) return;

        const newValues = [...values];
        for (let i = 0; i < paste.length; i++) {
            newValues[i] = paste[i].toUpperCase(); // Приводим к верхнему регистру
        }
        setValues(newValues);

        // Очищаем ошибку при вставке
        if (errors.code) {
            setErrors(prev => ({ ...prev, code: '' }));
        }

        const nextIndex = paste.length < values.length ? paste.length : values.length - 1;
        inputsRef.current[nextIndex]?.focus();
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        
        clearErrors();
        
        if (!validateCode()) {
            return;
        }

        setIsLoading(true);

        try {
            const response = await confirm_code(values.join(''), sessionId, "register");
            
            if (response.verified) {
                try {
                    await createUser(email, userName, password, sessionId);
                    router.push(`/register/complete?username=${userName}&email=${email}&password=${password}&sessionId=${sessionId}`);
                } catch (error: any) {
                    console.error('Ошибка при создании аккаунта:', error);
                    
                    let errorMessage = 'Что-то пошло не так при создании аккаунта';
                    
                    if (error.status === 400) {
                        errorMessage = 'Пользователь с таким email уже существует';
                    } else if (error.status === 409) {
                        errorMessage = 'Пользователь с таким именем уже существует';
                    } else if (error.status >= 500) {
                        errorMessage = 'Ошибка сервера. Попробуйте позже';
                    }
                    
                    showNotification(error.status, errorMessage);
                }
            } else {
                showNotification(404, 'Неверный код подтверждения');
            }
        } catch (error: any) {
            console.error('Ошибка при проверке кода:', error);
            
            let errorMessage = 'Что-то пошло не так при проверке кода';
            
            if (error.status === 401) {
                errorMessage = 'Код некорректен';
            } else if (error.status === 404) {
                errorMessage = 'Неверный код подтверждения';
            } else if (error.status === 429) {
                errorMessage = 'Слишком много попыток. Попробуйте позже';
            } else if (error.status >= 500) {
                errorMessage = 'Ошибка сервера. Попробуйте позже';
            }
            
            showNotification(error.status, errorMessage);
        } finally {
            setIsLoading(false);
        }
    };

    return (
		<VerifiContainer title="Код подтверждения" onSubmit={handleSubmit}>
			
			<div className="text-center text-base text-black">
                Введите код подтверждения из письма, отправленного на почту <span className="font-bold">{email}</span>
            </div>

			<div className="flex gap-3 justify-center mt-8">
                {values.map((val, index) => (
                    <input
                        key={index}
                        ref={(el) => { inputsRef.current[index] = el; }}
                        type="text"
                        inputMode="text"
                        maxLength={1}
                        className={`w-16 h-17 border-2 rounded-lg text-center text-2xl focus:outline-none focus:ring-2 focus:ring-blue-400 ${
                            errors.code ? 'border-[#ff4444]' : 'border-[#316BC2]'
                        }`}
                        value={val}
                        onChange={(e) => handleChange(index, e.target.value)}
                        onKeyDown={(e) => handleKeyDown(index, e)}
                        onPaste={handlePaste}
                        disabled={isLoading}
                    />
                ))}
            </div>

            {errors.code && (
                <div className="errorMessage" style={{ textAlign: 'center', marginTop: '8px' }}>
                    {errors.code}
                </div>
            )}

			<div className="text-center text-sm text-black/50 mt-4">
                {resendAvailable ? (
                <button
                    type="button"
                    className="hover:underline text-xl text-black disabled:opacity-50 disabled:cursor-not-allowed"
                    onClick={handleResendCode}
                    disabled={isLoading}
                >
                    {isLoading ? 'Отправляем...' : 'Отправить код повторно'}
                </button>
                ) : (
                formatTime()
                )}
            </div>

			<div className="mt-25">
				<FormButton type="submit" disabled={isLoading}>
					{isLoading ? 'Подтверждаем...' : 'Подтвердить'}
				</FormButton>
			</div>
			
		</VerifiContainer>
	);
}