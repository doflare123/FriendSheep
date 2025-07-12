'use client';

import { useState, useRef, useEffect } from 'react';
import "../../ConfirmCode.css";
import { useRouter } from 'next/router';

export default function ConfirmCode() {
    const [values, setValues] = useState(['', '', '', '', '', '']);
    const inputsRef = useRef([]);

    const [timeLeft, setTimeLeft] = useState(15 * 60);
    const [resendAvailable, setResendAvailable] = useState(false);

    const router = useRouter();
    const { email, username, session_id } = router.query;

    useEffect(() => {
        if (timeLeft <= 0) {
            setResendAvailable(true);
            return;
        }

        const interval = setInterval(() => {
            setTimeLeft((prev) => prev - 1);
        }, 1000);

        return () => clearInterval(interval);
    }, [timeLeft]);

    const handleResendCode = () => {
        // Здесь добавьте ваш функционал для повторной отправки кода
        console.log("Код отправлен повторно");
        setTimeLeft(15 * 60); // Сбросить на 15 минут
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

    return (
        <div className="page-wrapper">
            <main className="main-center">
                <div className="max-w-xl min-h-[800px]">
                    <div className="form-container">
                        <h1 className="heading-title">Код подтверждения</h1>

                        <div className="dots-animation">
                            <div className="dot dot1" />
                            <div className="dot dot2" />
                            <div className="dot dot3" />
                        </div>

                        <form className="space-y-4 mt-12">
                            <div className="text-center text-base text-black font-medium leading-tight">
                                Введите код подтверждения из письма, отправленного на почту <br />
                                <span className="font-bold">{email}</span>
                            </div>

                            <div className="flex gap-3 justify-center">
                                {values.map((val, index) => (
                                <input
                                    key={index}
                                    ref={(el) => (inputsRef.current[index] = el)}
                                    type="text"
                                    inputMode="numeric"
                                    maxLength={1}
                                    className="w-16 h-16 border-2 border-[#316BC2] rounded-lg text-center text-2xl focus:outline-none focus:ring-2 focus:ring-blue-400"
                                    value={val}
                                    onChange={(e) => handleChange(index, e.target.value)}
                                    onKeyDown={(e) => handleKeyDown(index, e)}
                                />
                                ))}
                            </div>

                            <div className="text-center text-sm text-black/50 mt-2">
                                {resendAvailable ? (
                                <button
                                    type="button"
                                    className="text-blue-600 font-bold hover:underline text-xl"
                                    onClick={handleResendCode}
                                >
                                    Отправить код
                                </button>
                                ) : (
                                formatTime()
                                )}
                            </div>

                            <div className="mt-15">
                                <button className="button-primary" type="submit">
                                Подтвердить
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </main>
        </div>
    );
}
