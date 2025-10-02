'use client';

import { useEffect, useState } from 'react';
import ProfilePage from './ProfilePage';
import { notFound, useParams } from 'next/navigation';
import { UserDataResponse } from '../../../types/UserData';
import { getUserInfo } from '../../../api/profile/getOwnProfile';
import { getOtherUserInfo } from '../../../api/profile/getProfile';
import LoadingIndicator from '@/components/LoadingIndicator';
import { convertUstoId } from '@/api/search/convertUstoId';
import { showNotification } from '@/utils';
import {getAccesToken} from '@/Constants';

import { useAuth } from '@/contexts/AuthContext';

export default function Page() {
  const params = useParams();
  const userUs = params.id as string;
  const { userData } = useAuth();

  const [profileData, setProfileData] = useState<UserDataResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [isOwn, setIsOwn] = useState(false);
  const [userId, setUserId] = useState<string>('');

  useEffect(() => {

    if (!userData) {
      return;
    }

    const loadProfile = async () => {
      try {
        const accessToken = getAccesToken();
        
        // Конвертируем Us в Id
        const convertedUserId = await convertUstoId(accessToken, userUs);
        setUserId(convertedUserId);

        // Проверяем, это собственный профиль или чужой
        const ownUs = userData?.us || 'ZZZ';
        const isOwnProfile = userUs === ownUs;
        setIsOwn(isOwnProfile);

        console.log("ZXC", ownUs, userUs, convertedUserId, userData, isOwnProfile, accessToken);

        let data: UserDataResponse;

        if (isOwnProfile) {
          data = await getUserInfo(accessToken);
        } else {
          data = await getOtherUserInfo(accessToken, convertedUserId);
        }

        setProfileData(data);
      } catch (err: any) {
        showNotification(
          err.status || 'error',
          'Ошибка при загрузке профиля: ' + (err.errorMessage || err.message || 'Неизвестная ошибка')
        );
      } finally {
        setLoading(false);
      }
    };

    loadProfile();
  }, [userUs, userData]); // Зависимость только от userUs, чтобы не было бесконечного цикла

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-full">
        <LoadingIndicator />
      </div>
    );
  }

  if (!profileData) {
    notFound();
    return null;
  }

  return <ProfilePage params={{ id: userId, data: profileData, isOwn }} />;
}