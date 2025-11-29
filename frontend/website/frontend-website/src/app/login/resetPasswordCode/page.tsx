'use client';

export const dynamic = 'force-dynamic';

import { useSearchParams, useRouter } from 'next/navigation';
import { sendEmail } from '@/api/resetPswd/sendEmail';
import { confirm_code } from '@/api/confirm_code';
import CodeVerification from '@/components/CodeVerification';

export default function ResetPasswordCode() {
    const searchParams = useSearchParams();
    const router = useRouter();

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