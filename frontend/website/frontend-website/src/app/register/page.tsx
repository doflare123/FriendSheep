export default function RegisterPage() {
return (
    <div className="page-wrapper">
        <main className="main-center">
            <div className="form-container">
                <h1 className="heading-title">Регистрация</h1>

                <form className="space-y-4">

                    <div>
                        <label className="label-style" htmlFor="username">
                            Имя пользователя
                        </label>
                        <input
                            id="username"
                            className="input-style"
                            placeholder="Имя пользователя"
                        />
                    </div>

                    <div>
                        <label className="label-style" htmlFor="password">
                            Пароль
                        </label>
                        <input
                            id="password"
                            className="input-style"
                            placeholder="Пароль"
                            type="password"
                        />
                    </div>

                    <div>
                        <label className="label-style" htmlFor="email">
                            Почта
                        </label>
                        <input
                            id="email"
                            className="input-style"
                            placeholder="user_email@gmail.com"
                            type="email"
                        />
                    </div>

                    <div className="link-small-note">
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
                        className="button-primary"
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

