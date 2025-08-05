'use client';

import React from 'react';
import { useParams } from 'next/navigation';
import GroupProfile from '../../../../components/Groups/profile/GroupProfile';

const GroupProfilePage: React.FC = () => {
  const params = useParams();
  const groupId = params.id as string;

  return <GroupProfile groupId={groupId} />;
};

export default GroupProfilePage;