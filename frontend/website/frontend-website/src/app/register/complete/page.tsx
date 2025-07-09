export default function CompleteReg() {
    return (
        <div className="page-wrapper">
            <main className="main-center">
                <div className="form-container">
                    <h1 className="heading-title">Аккаунт создан!</h1>

                    <form className="space-y-4">

                        <img src="/accept.png" alt="Успешно" className="w-24 h-24 mb-4 mx-auto" />

                        <p className="label-style text-center text-xl">
                            Ваш аккаунт успешно подтвержден и активирован.<br />
                            Теперь вы можете приступить к пользованию сервисом.<br />
                            Спасибо, что выбрали нас!
                        </p>

                        <img src="/logo.png" alt="Логотип" className="w-52 h-52 mb-4 mx-auto mt-8 rounded-3xl" />

                        <button
                            className="button-primary"
                            type="submit"
                            >
                            Перейти ко входу
                        </button>
                    </form>
                </div>
            </main>
        </div>
    );
}
