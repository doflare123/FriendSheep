'use client';

import Image from 'next/image';
import Link from 'next/link';

export default function Header() {
    return (
        <header className="flex justify-between items-center px-6 py-5 bg-white shadow relative">
            {/* Логотип слева */}
            <div className="flex items-center gap-2">
                <Image src="/logo.png" alt="Логотип" width={64} height={64} />
                <span className="text-sky-500 text-3xl font-inter">FriendShip</span>
            </div>

            {/* Центрированные ссылки */}
            <nav className="absolute left-1/2 -translate-x-1/2 flex gap-6 text-2xl text-black font-inter">
                <Link href="#" className="hover:underline">
                Главная
                </Link>
                <Link href="#" className="hover:underline">
                Новости
                </Link>
            </nav>

            {/* Кнопки справа */}
            <div className="text-2xl text-black font-inter flex items-center gap-2">
                <Link href="#" className="hover:underline px-1">
                Войти
                </Link>
                <span className="select-none">/</span>
                <Link href="#" className="hover:underline px-1">
                Регистрация
                </Link>
            </div>
        </header>
    );
}
