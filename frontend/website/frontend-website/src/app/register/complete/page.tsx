'use client';

import { useRouter } from 'next/navigation';

import FormContainer from '../../../components/FormContainer';
import FormButton from '../../../components/FormButton';
import Image from 'next/image'

export default function CompleteReg() {
    const router = useRouter();

    const handleSubmit = () => {
        router.push(`/login`);
    }

    return (
        <FormContainer title="Аккаунт создан!" onSubmit={handleSubmit}>
            
            <Image src="/accept.png" alt="Успешно" width={96} height={96} className="mb-4 mx-auto" />

            <p className="label-style text-center text-xl">
                Ваш аккаунт успешно подтвержден и активирован.<br />
                Теперь вы можете приступить к пользованию сервисом.<br />
                Спасибо, что выбрали нас!
            </p>

            <Image src="/logo.png" alt="Логотип" width={208} height={208} className="mb-4 mx-auto mt-8 rounded-3xl" />

            <FormButton type="submit">
                Перейти ко входу
            </FormButton>
        
        </FormContainer>
    );
}
