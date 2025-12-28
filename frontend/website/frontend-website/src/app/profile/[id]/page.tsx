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
import {getAccesToken, updateUserData} from '@/Constants';
import { getGroups } from '@/api/get_groups';
import {SmallGroup} from '@/types/Groups';
import { useRouter } from 'next/navigation';

import { useAuth } from '@/contexts/AuthContext';

export default function Page() {
  const params = useParams();
  const userUs = params.id as string;
  const { userData } = useAuth();
  const router = useRouter();

  const [profileData, setProfileData] = useState<UserDataResponse | null>(null);
  const [subsData, setSubsData] = useState<SmallGroup[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [isOwn, setIsOwn] = useState(false);
  const [userId, setUserId] = useState<string>('');

  const [complete, setComplete] = useState(false); 

  useEffect(() => {

    if (!userData || complete) {
      return;
    }

    const loadProfile = async () => {
      try {
        const accessToken = await getAccesToken(router);
        
        // Конвертируем Us в Id
        const convertedUserId = await convertUstoId(accessToken, userUs);
        setUserId(convertedUserId);

        // Проверяем, это собственный профиль или чужой
        const ownUs = userData?.us || 'ZZZ';
        const isOwnProfile = userUs === ownUs;
        setIsOwn(isOwnProfile);

        let data: UserDataResponse;
        let subs: SmallGroup[];

        if (isOwnProfile) {
          data = await getUserInfo(accessToken);
          subs = await getGroups(accessToken);
        } else {
          data = await getOtherUserInfo(accessToken, convertedUserId);
          subs = await getGroups(accessToken, convertedUserId);
        }

        console.log("DATA", data)
        setProfileData(data);
        setSubsData(subs);
      } catch (err: any) {
        showNotification(
          err.status || 'error',
          'Ошибка при загрузке профиля: ' + (err.errorMessage || err.message || 'Неизвестная ошибка')
        );
      } finally {
        setLoading(false);
        setComplete(true);
      }
    };

    loadProfile();
  }, [userData]); // Зависимость только от userUs, чтобы не было бесконечного цикла

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

  return <ProfilePage params={{ id: userId, data: profileData, isOwn, subs: subsData}} />;
}