'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import styles from '../styles/MainPage.module.css';
import { getAccesToken } from '@/Constants';
import MainHome from '@/components/MainHome';
import GuestHome from '@/components/GuestHome';
import LoadingIndicator from '@/components/LoadingIndicator';

export default function Home() {
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isMobile, setIsMobile] = useState(false);
  const router = useRouter();

  useEffect(() => {
    const checkAuth = async () => {
      const accessToken = await getAccesToken();
      setToken(accessToken);
      setIsLoading(false);
    };

    checkAuth();
  }, []);

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth <= 768);
    };

    checkMobile();
    window.addEventListener('resize', checkMobile);

    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  if (isLoading) {
    return (
      <div className={styles.pageWrapper}>
        <div className='bgPage'>
          <div className={styles.contentWrapper}>
            <LoadingIndicator text="Загрузка..." />
          </div>
        </div>
      </div>
    );
  }

  if (!token || isMobile) {
    return <GuestHome />;
  }

  return (
    <div className={styles.pageWrapper}>
      <div className='bgPage'>
        <MainHome token={token} />
      </div>
    </div>
  );
}