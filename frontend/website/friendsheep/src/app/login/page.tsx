// src/app/login/page.tsx
import Image from 'next/image';

export default function LoginPage() {
  return (
    <div className="min-h-screen bg-[url('/bg_login.png')] bg-cover bg-center bg-no-repeat">
      {/* Header */}
      <header className="flex justify-between items-center px-6 py-4 bg-white shadow relative">
        {/* Логотип слева */}
        <div className="flex items-center gap-2">
            <Image src="/logo.png" alt="Логотип" width={32} height={32} />
            <span className="text-sky-500 font-semibold text-lg">FriendShip</span>
        </div>

        {/* Центрированные ссылки */}
        <nav className="absolute left-1/2 -translate-x-1/2 flex gap-6 text-sm text-black">
            <a href="#" className="hover:underline">Главная</a>
            <a href="#" className="hover:underline">Новости</a>
        </nav>

        {/* Пустой блок или кнопка справа */}
        <div className="text-sm text-black">
            <a href="#" className="hover:underline">Войти / Регистрация</a>
        </div>
      </header>

      {/* Login Form */}
      <main className="flex items-center justify-center min-h-[calc(100vh-80px)]">
        <div className="bg-white rounded-xl p-8 shadow-lg max-w-sm w-full">
          <h1 className="text-2xl font-semibold text-center text-sky-600 mb-4">Вход</h1>
          <form className="space-y-4">
            <input
              className="w-full rounded border p-2"
              placeholder="user_email@gmail.com"
              type="email"
            />
            <input
              className="w-full rounded border p-2"
              placeholder="Пароль"
              type="password"
            />
            <div className="text-right text-sm text-gray-600">
              <a href="#" className="hover:underline">Нет аккаунта?</a>
            </div>
            <button
              className="w-full rounded bg-blue-500 p-2 text-white hover:bg-blue-600 transition"
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
