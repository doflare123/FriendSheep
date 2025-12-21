'use client';

export const dynamic = 'force-dynamic';

import { useEffect } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { registerSession } from '../../../api/register_sessions';
import { confirm_code } from '../../../api/confirm_code';
import { createUser } from '../../../api/create_user';
import { usePageProtection, PAGE_KEYS } from '../../../api/pageProtection';
import CodeVerification from '../../../components/CodeVerification';
import {checkDeviceAndRedirect} from '@/Constants';

export default function ConfirmCode() {
    // usePageProtection({ 
    //     pageKey: PAGE_KEYS.CODE_VERIFY,
    //     redirectTo: '/register'
    // });

    const searchParams = useSearchParams();
    const router = useRouter();

    useEffect(() => {
        checkDeviceAndRedirect(router);
    }, [router]);

    const email: string = searchParams.get('email') || "";
    const userName: string = searchParams.get('username') || "";
    const password: string = searchParams.get('password') || "";
    const sessionId: string = searchParams.get('sessionId') || "";

    const handleSuccess = (sessionId: string) => {
        router.push(`/register/complete?username=${userName}&email=${email}&password=${password}&sessionId=${sessionId}`);
    };

    return (
        <CodeVerification
            email={email}
            sessionId={sessionId}
            mode="register"
            userName={userName}
            password={password}
            onSuccess={handleSuccess}
            onResendCode={registerSession}
            onVerifyCode={confirm_code}
            onCreateUser={createUser}
        />
    );
}