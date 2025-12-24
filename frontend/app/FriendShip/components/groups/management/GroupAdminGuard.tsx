import groupService from '@/api/services/group/groupService';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useEffect, useState } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';

interface GroupAdminGuardProps {
  groupId: string;
  children: React.ReactNode;
}

export const GroupAdminGuard: React.FC<GroupAdminGuardProps> = ({ 
  groupId, 
  children 
}) => {
  const colors = useThemedColors();
  const [isAuthorized, setIsAuthorized] = useState<boolean | null>(null);
  const navigation = useNavigation<StackNavigationProp<any>>();

  useEffect(() => {
    checkAdminAccess();
  }, [groupId]);

  const checkAdminAccess = async () => {
    try {
      await groupService.getGroupDetail(groupId);
      setIsAuthorized(true);
    } catch (error: any) {
      console.error('[GroupAdminGuard] Доступ запрещён');
      setIsAuthorized(false);
      
      navigation.replace('GroupPage', { groupId });
    }
  };

  if (isAuthorized === null) {
    return (
      <View style={[styles.container, { backgroundColor: colors.white }]}>
        <ActivityIndicator size="large" color={colors.blue2} />
        <Text style={[styles.text, { color: colors.grey }]}>
          Проверка прав доступа...
        </Text>
      </View>
    );
  }

  if (!isAuthorized) {
    return null;
  }

  return <>{children}</>;
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  text: {
    marginTop: 10,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
});