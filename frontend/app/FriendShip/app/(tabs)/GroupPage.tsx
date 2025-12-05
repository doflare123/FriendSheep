import { groupMemberService } from '@/api/services/group';
import groupService, { PublicGroupResponse } from '@/api/services/group/groupService';
import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import ConfirmationModal from '@/components/ConfirmationModal';
import EventCarousel from '@/components/event/EventCarousel';
import PrivateGroupPreview from '@/components/groups/PrivateGroupPreview';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { groupSessionsToEvents } from '@/utils/dataAdapters';
import { filterActiveSessions } from '@/utils/sessionStatusHelpers';
 
import { RouteProp, useFocusEffect, useNavigation, useRoute } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import React, { useCallback, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Image,
  Linking,
  Modal,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

type GroupPageRouteProp = RouteProp<RootStackParamList, 'GroupPage'>;

type GroupManagePageNavigationProp = NativeStackNavigationProp<RootStackParamList>;

type MembershipStatus = 'admin' | 'member' | 'pending' | 'not_member';

const categoryIcons: Record<string, any> = {
  movie: require('../../assets/images/event_card/movie.png'),
  game: require('../../assets/images/event_card/game.png'),
  table_game: require('../../assets/images/event_card/table_game.png'),
  other: require('../../assets/images/event_card/other.png'),
};

const CATEGORY_MAPPING: { [key: string]: string } = {
  'Фильмы': 'movie',
  'Игры': 'game',
  'Настольные игры': 'table_game',
  'Другое': 'other',
};

const contactIcons: Record<string, any> = {
  discord: require('@/assets/images/groups/contacts/discord.png'),
  vk: require('@/assets/images/groups/contacts/vk.png'),
  telegram: require('@/assets/images/groups/contacts/telegram.png'),
  twitch: require('@/assets/images/groups/contacts/twitch.png'),
  youtube: require('@/assets/images/groups/contacts/youtube.png'),
  whatsapp: require('@/assets/images/groups/contacts/whatsapp.png'),
  max: require('@/assets/images/groups/contacts/max.png'),
  default: require('@/assets/images/groups/contacts/default.png'),
};

const GroupPage = () => {
  const route = useRoute<GroupPageRouteProp>();
  const { groupId } = route.params;
  const [descriptionModalVisible, setDescriptionModalVisible] = useState(false);
  const [groupData, setGroupData] = useState<PublicGroupResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isProcessing, setIsProcessing] = useState(false);
  const [membershipStatus, setMembershipStatus] = useState<MembershipStatus>('not_member');
  const [isPrivateGroup, setIsPrivateGroup] = useState(false);
  const [privateGroupName, setPrivateGroupName] = useState<string>('');
  const { sortingState, sortingActions } = useSearchState();
  const [confirmLeaveModalVisible, setConfirmLeaveModalVisible] = useState(false);
  const [requestStatus, setRequestStatus] = useState<'none' | 'pending' | 'approved' | 'rejected'>('none');
  const navigation = useNavigation<GroupManagePageNavigationProp>();

  const loadGroupData = async () => {
    try {
      setIsLoading(true);
      setIsPrivateGroup(false);
      
      console.log('[GroupPage] Загружаем данные группы...');

      let isAdmin = false;
      try {
        await groupService.getGroupDetail(groupId);
        isAdmin = true;
      } catch (adminCheckError: any) {
        if (adminCheckError.response?.status === 403) {
          console.log('[GroupPage] Пользователь не является администратором');
        }
      }

      try {
        const data = await groupService.getPublicGroupDetail(groupId);

        setGroupData(data);

        if (isAdmin) {
          setMembershipStatus('admin');
        } else if (data.subscription) {
          setMembershipStatus('member');
        } else {
          setMembershipStatus('not_member');
        }

      } catch (publicError: any) {
        console.error('[GroupPage] Ошибка загрузки:', publicError);
        
        if (publicError.response?.status === 500 && 
            publicError.response?.data?.error?.includes('приватной группе запрещен')) {
          setIsPrivateGroup(true);
          setPrivateGroupName('');
        } else {
          throw publicError;
        }
      }

    } catch (error: any) {
      console.error('[GroupPage] Критическая ошибка:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось загрузить группу');
    } finally {
      setIsLoading(false);
    }
  };

  useFocusEffect(
    useCallback(() => {
      loadGroupData();
    }, [groupId])
  );

  const handleSessionUpdate = useCallback(() => {
    console.log('[GroupPage] 🔄 Обновление данных группы после изменения сессии');
    loadGroupData();
  }, [groupId]);

  const handlePrivateGroupRequestJoin = async () => {
    try {
      setIsProcessing(true);
      console.log('[GroupPage] Подача заявки в приватную группу');
      
      const result = await groupService.joinGroup(parseInt(groupId));
      
      console.log('[GroupPage] Результат вступления:', result);
      
      if (result.joined) {
        setRequestStatus('approved');
      } else {
        setRequestStatus('pending');
      }
      
      Alert.alert('Успешно', result.message);
      await loadGroupData();
      
    } catch (error: any) {
      console.error('[GroupPage] Ошибка вступления в группу:', error);

      const errorMessage = error.response?.data?.error || error.message || '';
      
      if (errorMessage.includes('заявка уже отправлена')) {
        setRequestStatus('pending');
        Alert.alert(
          'Заявка уже отправлена',
          'Ваша заявка на вступление уже находится на рассмотрении у администратора группы.'
        );
        await loadGroupData();
      } else {
        Alert.alert('Ошибка', errorMessage || 'Не удалось подать заявку');
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleContactPress = (link: string) => {
    if (link) {
      Linking.openURL(link).catch(err => {
        console.error('Не удалось открыть ссылку:', err);
        Alert.alert('Ошибка', 'Не удалось открыть ссылку');
      });
    }
  };

  const handleUserPress = (username: string) => {
    console.log('🔍 [GroupPage] handleUserPress вызван');
    console.log('🔍 [GroupPage] username:', username);
    console.log('🔍 [GroupPage] тип:', typeof username);
    
    navigation.navigate('ProfilePage', { userId: username });
  };

  const getContactIcon = (contactName: string, contactLink?: string) => {
    if (!contactLink) {
      return contactIcons.default;
    }
    
    const lowerLink = contactLink.toLowerCase();

    if (lowerLink.includes('discord.gg') || lowerLink.includes('discord.com')) {
      return contactIcons.discord;
    }
    if (lowerLink.includes('vk.com') || lowerLink.includes('vk.ru')) {
      return contactIcons.vk;
    }
    if (lowerLink.includes('t.me') || lowerLink.includes('telegram')) {
      return contactIcons.telegram;
    }
    if (lowerLink.includes('twitch.tv')) {
      return contactIcons.twitch;
    }
    if (lowerLink.includes('youtube.com') || lowerLink.includes('youtu.be')) {
      return contactIcons.youtube;
    }
    if (lowerLink.includes('wa.me') || lowerLink.includes('whatsapp')) {
      return contactIcons.whatsapp;
    }
    if (lowerLink.includes('max.ru')) {
      return contactIcons.max;
    }
    
    return contactIcons.default;
  };

  const handleActionButton = async () => {
    if (membershipStatus === 'admin') {
      console.log('[GroupPage] Переход к управлению группой');
      navigation.navigate('GroupManagePage', { groupId });
    } else if (membershipStatus === 'member') {
      setConfirmLeaveModalVisible(true);
    } else if (membershipStatus === 'not_member') {
      try {
        setIsProcessing(true);
        console.log('[GroupPage] Подача заявки на вступление');
        
        const result = await groupService.joinGroup(parseInt(groupId));
        
        console.log('[GroupPage] Результат вступления:', result);

        Alert.alert('Успешно', result.message);

        if (result.joined) {
          setMembershipStatus('member');
        } else {
          setMembershipStatus('pending');
        }
        
        await loadGroupData();
      } catch (error: any) {
        console.error('[GroupPage] Ошибка вступления в группу:', error);
        Alert.alert('Ошибка', error.message || 'Не удалось подать заявку');
      } finally {
        setIsProcessing(false);
      }
    }
  };

  const handleLeaveGroupConfirm = async () => {
    try {
      setConfirmLeaveModalVisible(false);
      setIsProcessing(true);
      
      console.log('[GroupPage] 🔍 Попытка покинуть группу, groupId:', groupId);
      console.log('[GroupPage] 🔍 Тип groupId:', typeof groupId);
      
      await groupMemberService.leaveGroup(parseInt(groupId));
      
      Alert.alert('Успешно', 'Вы покинули группу');
      await loadGroupData();
    } catch (error: any) {
      console.error('[GroupPage] Ошибка выхода из группы:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось покинуть группу');
    } finally {
      setIsProcessing(false);
    }
  };

  const getActionButtonText = () => {
    if (isProcessing) return 'Загрузка...';
    
    switch (membershipStatus) {
      case 'admin':
        return 'Управлять';
      case 'member':
        return 'Покинуть';
      case 'pending':
        return 'Заявка отправлена';
      case 'not_member':
        return 'Присоединиться';
      default:
        return 'Присоединиться';
    }
  };

  const getActionButtonStyle = () => {
    switch (membershipStatus) {
      case 'admin':
        return styles.actionButton;
      case 'member':
        return [styles.actionButton, styles.leaveButton];
      case 'pending':
        return [styles.actionButton, styles.pendingButton];
      case 'not_member':
        return [styles.actionButton, styles.joinButton];
      default:
        return styles.actionButton;
    }
  };

  if (isLoading) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color={Colors.blue} />
          <Text style={styles.loadingText}>Загрузка группы...</Text>
        </View>
        <BottomBar />
      </SafeAreaView>
    );
  }
  
  if (isPrivateGroup) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <PrivateGroupPreview
          groupName={privateGroupName}
          onRequestJoin={handlePrivateGroupRequestJoin}
          isProcessing={isProcessing}
          requestStatus={requestStatus}
        />
        <BottomBar />
      </SafeAreaView>
    );
  }

  if (!groupData) {
    return (
      <SafeAreaView style={styles.container}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>Группа не найдена</Text>
        </View>
        <BottomBar />
      </SafeAreaView>
    );
  }

  const mappedCategories = groupData.categories
  .map(cat => CATEGORY_MAPPING[cat])
  .filter(cat => cat !== undefined);

  const activeSessions = groupData.sessions 
    ? filterActiveSessions(groupData.sessions)
    : [];

  const formattedSessions = groupSessionsToEvents(
    { ...groupData, sessions: activeSessions }, 
    handleSessionUpdate
  );

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <View style={styles.header}>
          <Image source={{ uri: groupData.image }} style={styles.groupImage} />
          <View style={{ flexDirection: 'column', flex: 1 }}>
            <View style={styles.headerInfo}>
              <Text style={styles.groupName}>{groupData.name}</Text>
              {groupData.city && groupData.city.trim() !== '' && (
                <Text style={styles.location}>{groupData.city}</Text>
              )}
              <View style={styles.categoriesContainer}>
                {mappedCategories.map((category, index) => (
                  <Image
                    key={index}
                    source={categoryIcons[category]}
                    style={styles.categoryIcon}
                  />
                ))}
              </View>
            </View>
            <TouchableOpacity
              onPress={handleActionButton}
              style={getActionButtonStyle()}
              disabled={isProcessing || membershipStatus === 'pending'}
            >
              {isProcessing ? (
                <ActivityIndicator color={Colors.white} size="small" />
              ) : (
                <Text style={styles.actionButtonText}>
                  {getActionButtonText()}
                </Text>
              )}
            </TouchableOpacity>
          </View>
        </View>

        <CategorySection title="Описание:">
          <TouchableOpacity onPress={() => setDescriptionModalVisible(true)}>
            <Text style={styles.descriptionText} numberOfLines={3}>
              {groupData.description || 'Описание пока не добавлено'}
            </Text>
          </TouchableOpacity>
        </CategorySection>

        <CategorySection title="Сессии:">
          {formattedSessions && formattedSessions.length > 0 ? (
            <EventCarousel events={formattedSessions} />
          ) : (
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>Пока нет активных сессий</Text>
            </View>
          )}
        </CategorySection>

        <CategorySection title={`Участники: ${groupData.count_members || groupData.users?.length || 0}`}>
          {groupData.users && groupData.users.length > 0 ? (
            <View style={styles.membersContainer}>
                {groupData.users.map((user, index) => {
                  console.log('👤 [GroupPage] user:', user);
                  console.log('👤 [GroupPage] user.us:', user.us);
                  
                  return (
                    <TouchableOpacity 
                      key={`member-${index}`} 
                      style={styles.memberItem}
                      onPress={() => handleUserPress(user.us)}
                      activeOpacity={0.7}
                    >
                      <Image 
                        source={{ uri: user.image }} 
                        style={styles.memberImage} 
                      />
                  </TouchableOpacity>
                );
              })}
            </View>
          ) : (
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>Пока нет участников</Text>
            </View>
          )}
        </CategorySection>

        <CategorySection title="Контакты:">
          {groupData.contacts && groupData.contacts.length > 0 ? (
            <View style={styles.contactsContainer}>
              {groupData.contacts.map((contact, index) => {
                const icon = getContactIcon(contact.name, contact.link);
                return (
                  <TouchableOpacity
                    key={`contact-${index}`}
                    style={styles.contactItem}
                    onPress={() => handleContactPress(contact.link)}
                  >
                    <View style={styles.contactIconContainer}>
                      <Image 
                        source={icon} 
                        style={styles.contactIcon} 
                      />
                    </View>
                    <Text style={styles.contactDescription} numberOfLines={2}>
                      {contact.name}
                    </Text>
                  </TouchableOpacity>
                );
              })}
            </View>
          ) : (
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>Контакты пока не добавлены</Text>
            </View>
          )}
        </CategorySection>
      </ScrollView>
      <BottomBar />

      <Modal
        visible={descriptionModalVisible}
        animationType="fade"
        transparent
        onRequestClose={() => setDescriptionModalVisible(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>Описание группы</Text>
              <TouchableOpacity
                onPress={() => setDescriptionModalVisible(false)}
                style={styles.closeButton}
              >
                <Text style={styles.closeButtonText}>✕</Text>
              </TouchableOpacity>
            </View>
            <ScrollView style={styles.modalScrollView}>
              <Text style={styles.modalDescriptionText}>{groupData.description}</Text>
            </ScrollView>
          </View>
        </View>
      </Modal>

      <ConfirmationModal
        visible={confirmLeaveModalVisible}
        title="Покинуть группу?"
        message="Вы уверены, что хотите покинуть эту группу?"
        onConfirm={handleLeaveGroupConfirm}
        onCancel={() => setConfirmLeaveModalVisible(false)}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  scrollView: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 10,
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  errorText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
  },
  header: {
    flexDirection: 'row',
    padding: 20,
    alignItems: 'center',
  },
  groupImage: {
    width: 110,
    height: 110,
    borderRadius: 100,
    marginRight: 20,
  },
  headerInfo: {
    flex: 1,
  },
  groupName: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.black,
    marginBottom: 4,
  },
  location: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    marginBottom: 8,
  },
  categoriesContainer: {
    flexDirection: 'row',
  },
  categoryIcon: {
    width: 24,
    height: 24,
    marginRight: 8,
    resizeMode: 'contain',
  },
  actionButton: {
    backgroundColor: Colors.lightBlue,
    paddingVertical: 4,
    marginTop: 20,
    borderRadius: 25,
    alignItems: 'center',
  },
  actionButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
  descriptionText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    lineHeight: 20,
    paddingHorizontal: 18,
    paddingVertical: 12,
  },
  contactsContainer: {
    paddingHorizontal: 18,
    paddingVertical: 12,
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'flex-start',
  },
  contactItem: {
    alignItems: 'center',
    width: 80,
    marginBottom: 16,
    marginHorizontal: 8,
  },
  contactIconContainer: {
    width: 60,
    height: 60,
    borderRadius: 100,
    backgroundColor: Colors.white,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 4,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.2,
    shadowRadius: 2,
  },
  contactIcon: {
    width: 40,
    height: 40,
    resizeMode: 'contain',
  },
  contactDescription: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.black,
    textAlign: 'center',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  modalContent: {
    backgroundColor: Colors.white,
    borderRadius: 20,
    margin: 20,
    maxHeight: '70%',
    width: '90%',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 12,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
  },
  modalTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    marginLeft: 8,
    color: Colors.black,
  },
  closeButton: {
    width: 30,
    height: 30,
    justifyContent: 'center',
    alignItems: 'center',
  },
  closeButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.black,
  },
  modalScrollView: {
    maxHeight: 300,
  },
  modalDescriptionText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    lineHeight: 20,
    padding: 20,
  },
  emptyContainer: {
    paddingHorizontal: 18,
    paddingVertical: 24,
    alignItems: 'center',
  },
  emptyText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
    textAlign: 'center',
  },
  joinButton: {
    backgroundColor: Colors.green,
  },
  leaveButton: {
    backgroundColor: Colors.red,
  },
  pendingButton: {
    backgroundColor: Colors.grey,
  },
  membersContainer: {
    flexDirection: 'row',
    paddingHorizontal: 18,
    paddingVertical: 12,
    flexWrap: 'wrap',
    gap: 12,
  },
  memberItem: {
    alignItems: 'center',
  },
  memberImage: {
    width: 75,
    height: 75,
    borderRadius: 40,
  },
});

export default GroupPage;