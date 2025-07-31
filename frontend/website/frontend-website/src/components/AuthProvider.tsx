'use client';

import { useEffect } from 'react';
import { setupAxiosInterceptors } from '../api/auth';

export default function AuthProvider({ children }: { children: React.ReactNode }) {
    useEffect(() => {
        // Настройка перехватчиков axios для автоматического обновления токенов
        setupAxiosInterceptors();
    }, []);

    return <>{children}</>;
}