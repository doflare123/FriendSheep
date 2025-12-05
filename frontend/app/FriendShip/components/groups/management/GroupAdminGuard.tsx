import groupService from '@/api/services/group/groupService';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
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
      <View style={styles.container}>
        <ActivityIndicator size="large" color={Colors.blue} />
        <Text style={styles.text}>Проверка прав доступа...</Text>
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
    backgroundColor: Colors.white,
  },
  text: {
    marginTop: 10,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
});