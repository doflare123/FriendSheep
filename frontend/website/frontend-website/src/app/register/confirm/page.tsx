'use client';

import { useState, useRef, useEffect } from 'react';
import "../../../styles/ConfirmCode.css";
import { useSearchParams, useRouter } from 'next/navigation';
import { registerSession } from '../../../api/sessions';
import { confirm_code } from '../../../api/confirm_code';
import { createUser } from '../../../api/create_user';

import FormButton from '../../../components/FormButton';
import VerifiContainer from '../../../components/VerifiContainer';

export default function ConfirmCode() {
    const [values, setValues] = useState(['', '', '', '', '', '']);
    const inputsRef = useRef<(HTMLInputElement | null)[]>([]);

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

    const handleResendCode = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const newSession = await registerSession(email);
            setSessionId(newSession);
        } catch (error) {
            console.error('Ошибка при отправке кода:', error);
            alert('Что-то пошло не так при повторной отправке кода');
        }
        const newExpiry = Math.floor(Date.now() / 1000) + 10 * 60;
        localStorage.setItem('codeExpiryTime', newExpiry.toString());
        setTimeLeft(10 * 60); 
        setResendAvailable(false);
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
        const newValues = [...values];
        newValues[index] = value;
        setValues(newValues);
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

        const newValues = [...values];
        for (let i = 0; i < paste.length; i++) {
            newValues[i] = paste[i];
        }
        setValues(newValues);

        const nextIndex = paste.length < values.length ? paste.length : values.length - 1;
        inputsRef.current[nextIndex]?.focus();
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const responce = await confirm_code(values.join(''), sessionId, "register");
            if (responce.verified){
                try {
                    await createUser(email, userName, password, sessionId);
                    router.push(`/register/complete?username=${userName}&email=${email}&password=${password}&sessionId=${sessionId}`);
                } catch (error) {
                    console.error('Ошибка при создании аккаунт:', error);
                    alert('Что-то пошло не так при создании аккаунт');
                }
            }
        } catch (error) {
            console.error('Ошибка при проверке кода:', error);
            alert('Что-то пошло не так при проверке кода');
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
                        inputMode="numeric"
                        maxLength={1}
                        className="w-16 h-17 border-2 border-[#316BC2] rounded-lg text-center text-2xl focus:outline-none focus:ring-2 focus:ring-blue-400"
                        value={val}
                        onChange={(e) => handleChange(index, e.target.value)}
                        onKeyDown={(e) => handleKeyDown(index, e)}
                        onPaste={handlePaste}
                    />
                ))}
            </div>

			<div className="text-center text-sm text-black/50 mt-4">
                {resendAvailable ? (
                <button
                    type="button"
                    className="hover:underline text-xl text-black"
                    onClick={handleResendCode}
                >
                    Отправить код повторно
                </button>
                ) : (
                formatTime()
                )}
            </div>

			<div className="mt-25">
				<FormButton type="submit">
					Подтвердить
				</FormButton>
			</div>
			
		</VerifiContainer>
	);
}
