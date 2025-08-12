'use client';

import React, { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { notFound } from 'next/navigation';
import GroupAdminComponent from '../../../../components/Groups/GroupAdminComponent';
import { GroupData } from '../../../../types/Groups';
import { getGroupInfo } from '../../../../api/get_group_info';

const GroupAdminPage: React.FC = () => {
    const params = useParams();
    const groupId = params.id as string;
    
    const [groupData, setGroupData] = useState<GroupData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchGroupData = async () => {
        try {
            const accessToken = localStorage.getItem('access_token');
            
            if (!accessToken) {
            setError('Токен доступа не найден');
            setLoading(false);
            return;
            }

            if (!groupId || isNaN(parseInt(groupId))) {
            setError('Некорректный ID группы');
            setLoading(false);
            return;
            }

            const data = await getGroupInfo(accessToken, parseInt(groupId));
            
            if (!data) {
            // Если данных нет, показываем 404
            notFound();
            } else {
            setGroupData(data);
            }
        } catch (err: any) {
            console.error('Ошибка при загрузке группы:', err);
            
            // Обработка различных типов ошибок согласно API документации
            if (err.response?.status === 400) {
            setError('Некорректный ID группы');
            } else if (err.response?.status === 401) {
            setError('Пользователь не авторизован');
            } else if (err.response?.status === 403) {
            setError('Доступ к приватной группе запрещен');
            } else if (err.response?.status === 404) {
            // Показываем стандартную 404 страницу Next.js
            notFound();
            } else if (err.response?.status === 500) {
            setError('Внутренняя ошибка сервера');
            } else {
            setError('Ошибка при загрузке группы');
            }
        } finally {
            setLoading(false);
        }
        };

        fetchGroupData();
    }, [groupId]);

    if (loading) {
        return (
        <div className="flex justify-center items-center h-64">
            <div className="text-lg">Загрузка...</div>
        </div>
        );
    }

    if (error) {
        notFound();
    }

    if (!groupData) {
        // Если данных нет после загрузки, показываем 404
        notFound();
    }

    return (
        <GroupAdminComponent 
            groupId={groupId}
            groupData={groupData}
        />
    );
};

export default GroupAdminPage;