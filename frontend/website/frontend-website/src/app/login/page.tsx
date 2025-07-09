// src/app/login/page.tsx

export default function LoginPage() {
  return (
    <div className="page-wrapper">
      <main className="main-center">
        <div className="form-container">
          <h1 className="heading-title">Вход</h1>

          <form className="space-y-4">
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

            <div className="link-small-note">
              <a href="#" className="hover:underline">
                Нет аккаунта?
              </a>
            </div>

            <button
              className="button-primary"
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

