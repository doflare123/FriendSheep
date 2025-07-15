'use client';

import { useRouter } from 'next/navigation';

import FormContainer from '../../../components/FormContainer';
import FormButton from '../../../components/FormButton';

export default function CompleteReg() {
    const router = useRouter();

    const handleSubmit = () => {
        router.push(`/login`);
    }

    return (
        <FormContainer title="Аккаунт создан!" onSubmit={handleSubmit}>
            
            <img src="/accept.png" alt="Успешно" className="w-24 h-24 mb-4 mx-auto" />

            <p className="label-style text-center text-xl">
                Ваш аккаунт успешно подтвержден и активирован.<br />
                Теперь вы можете приступить к пользованию сервисом.<br />
                Спасибо, что выбрали нас!
            </p>

            <img src="/logo.png" alt="Логотип" className="w-52 h-52 mb-4 mx-auto mt-8 rounded-3xl" />

            <FormButton type="submit">
                Перейти ко входу
            </FormButton>
        
        </FormContainer>
    );
}
