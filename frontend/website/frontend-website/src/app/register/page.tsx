export default function RegisterPage() {
return (
    <div className="h-full bg-[url('/bg_login.png')] bg-cover bg-center bg-no-repeat">
        <main className="flex items-center justify-center h-full">
            <div className="bg-white rounded-2xl p-10 w-full max-w-2xl font-alexandria border-2 border-[#316BC2] min-h-[475px] text-2xl">
                <h1 className="text-4xl text-center text-[#37A2E6] font-bold mb-6">Регистрация</h1>

                <form className="space-y-4">

                    <div>
                        <label className="block mb-1 text-gray-700" htmlFor="username">
                            Имя пользователя
                        </label>
                        <input
                            id="username"
                            className="w-full rounded-lg border-2 border-[#316BC2] p-3 text-xl"
                            placeholder="Имя пользователя"
                        />
                    </div>

                    <div>
                        <label className="block mb-1 text-gray-700" htmlFor="password">
                            Пароль
                        </label>
                        <input
                            id="password"
                            className="w-full rounded-lg border-2 border-[#316BC2] p-3 text-xl"
                            placeholder="Пароль"
                            type="password"
                        />
                    </div>

                    <div>
                        <label className="block mb-1 text-gray-700" htmlFor="email">
                            Почта
                        </label>
                        <input
                            id="email"
                            className="w-full rounded-lg border-2 border-[#316BC2] p-3 text-xl"
                            placeholder="user_email@gmail.com"
                            type="email"
                        />
                    </div>

                    <div className="text-right text-xl text-gray-600 -mt-2 font-bold">
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
                        className="w-full rounded-lg bg-[#37A2E6] border-2 border-[#316BC2] text-white py-3 hover:bg-[#2d93d5] transition mt-5"
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

