'use client';

export const dynamic = 'force-dynamic';

import { useEffect } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { sendEmail } from '@/api/resetPswd/sendEmail';
import { confirm_code } from '@/api/confirm_code';
import CodeVerification from '@/components/CodeVerification';
import {checkDeviceAndRedirect} from '@/Constants';

export default function ResetPasswordCode() {
    const searchParams = useSearchParams();
    const router = useRouter();

    useEffect(() => {
        checkDeviceAndRedirect(router);
    }, [router]);

    const email: string = searchParams.get('email') || "";
    const sessionId: string = searchParams.get('sessionId') || "";

    const handleSuccess = (sessionId: string) => {
        router.push(`/login/resetPasswordNew?email=${email}&sessionId=${sessionId}`);
    };

    return (
        <CodeVerification
            email={email}
            sessionId={sessionId}
            mode="reset_password"
            onSuccess={handleSuccess}
            onResendCode={sendEmail}
            onVerifyCode={confirm_code}
        />
    );
}