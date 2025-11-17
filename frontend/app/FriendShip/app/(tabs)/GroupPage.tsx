import groupService, { PublicGroupResponse } from '@/api/services/groupService';
import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event as EventType } from '@/components/event/EventCard';
import EventCarousel from '@/components/event/EventCarousel';
import PrivateGroupPreview from '@/components/groups/PrivateGroupPreview';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
// eslint-disable-next-line import/no-unresolved
import { LOCAL_IP } from '@env';
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
  const navigation = useNavigation<GroupManagePageNavigationProp>();

  const loadGroupData = async () => {
    try {
      setIsLoading(true);
      setIsPrivateGroup(false);
      
      console.log('[GroupPage] Загружаем данные группы...');

      let isAdmin = false;
      try {
        console.log('[GroupPage] Проверяем права администратора...');
        await groupService.getGroupDetail(groupId);
        isAdmin = true;
        console.log('[GroupPage] ✅ Пользователь является администратором группы');
      } catch (adminCheckError: any) {
        if (adminCheckError.response?.status === 403) {
          console.log('[GroupPage] ❌ Пользователь не является администратором (403)');
        } else {
          console.warn('[GroupPage] Ошибка проверки прав админа:', adminCheckError.message);
        }
      }

      try {
        const data = await groupService.getPublicGroupDetail(groupId);

        if (data.image && data.image.includes('localhost')) {
          data.image = data.image.replace('http://localhost:8080', 'http://' + LOCAL_IP + ':8080');
        }

        if (data.users) {
          data.users = data.users.map(user => ({
            ...user,
            image: user.image?.includes('localhost')
              ? user.image.replace('http://localhost:8080', 'http://' + LOCAL_IP + ':8080')
              : user.image
          }));
        }

        if (data.sessions) {
          data.sessions = data.sessions.map(session => ({
            ...session,
            session: {
              ...session.session,
              image_url: session.session.image_url?.includes('localhost')
                ? session.session.image_url.replace('http://localhost:8080', 'http://' + LOCAL_IP + ':8080')
                : session.session.image_url
            }
          }));
        }
        
        setGroupData(data);

        console.log('[GroupPage] Определяем статус пользователя...');
        console.log('[GroupPage] Является админом:', isAdmin);
        console.log('[GroupPage] Подписка (subscription):', data.subscription);

        if (isAdmin) {
          console.log('[GroupPage] 🎯 Статус: АДМИНИСТРАТОР');
          setMembershipStatus('admin');
        } else if (data.subscription) {
          console.log('[GroupPage] 👥 Статус: УЧАСТНИК');
          setMembershipStatus('member');
        } else {
          console.log('[GroupPage] 🚪 Статус: НЕ УЧАСТНИК');
          setMembershipStatus('not_member');
        }

      } catch (publicError: any) {
        console.error('[GroupPage] Ошибка загрузки публичных данных:', publicError);
        
        if (publicError.response?.status === 500 && 
            publicError.response?.data?.error?.includes('приватной группе запрещен')) {
          console.log('[GroupPage] 🔒 Это приватная группа, доступ запрещён');
          setIsPrivateGroup(true);
          setPrivateGroupName('');
        } else {
          throw publicError;
        }
      }

    } catch (error: any) {
      console.error('[GroupPage] Критическая ошибка загрузки группы:', error);
      Alert.alert('Ошибка', error.message || 'Не удалось загрузить информацию о группе');
    } finally {
      setIsLoading(false);
    }
  };

  useFocusEffect(
    useCallback(() => {
      loadGroupData();
    }, [groupId])
  );

  const handlePrivateGroupRequestJoin = async () => {
    try {
      setIsProcessing(true);
      console.log('[GroupPage] Подача заявки в приватную группу');
      
      const result = await groupService.joinGroup(parseInt(groupId));
      
      console.log('[GroupPage] Результат вступления:', result);
      Alert.alert('Успешно', result.message);
      
      await loadGroupData();
      
    } catch (error: any) {
      console.error('[GroupPage] Ошибка вступления в группу:', error);

      const errorMessage = error.response?.data?.error || error.message || '';
      
      if (errorMessage.includes('заявка уже отправлена')) {
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
      Alert.alert(
        'Покинуть группу?',
        'Вы уверены, что хотите покинуть эту группу?',
        [
          { text: 'Отмена', style: 'cancel' },
          {
            text: 'Покинуть',
            style: 'destructive',
            onPress: async () => {
              try {
                setIsProcessing(true);
                await groupService.leaveGroup(parseInt(groupId));
                Alert.alert('Успешно', 'Вы покинули группу');
                await loadGroupData();
              } catch (error: any) {
                console.error('[GroupPage] Ошибка выхода из группы:', error);
                Alert.alert('Ошибка', error.message || 'Не удалось покинуть группу');
              } finally {
                setIsProcessing(false);
              }
            },
          },
        ]
      );
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
        return 'Подать заявку';
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

  const formattedSessions: EventType[] = groupData.sessions?.map(item => ({
    id: item.session.id.toString(),
    title: item.session.title,
    date: new Date(item.session.start_time).toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    }),
    genres: item.metadata?.genres || [],
    currentParticipants: item.session.current_users,
    maxParticipants: item.session.count_users_max,
    duration: `${item.session.duration} мин`,
    imageUri: item.session.image_url,
    description: '',
    typeEvent: item.session.session_type,
    typePlace: item.session.session_place === 'offline' || item.session.session_place === 'online' 
      ? item.session.session_place as 'online' | 'offline'
      : 'online',
    eventPlace: item.metadata?.location || '',
    publisher: groupData.name,
    publicationDate: item.session.start_time,
    ageRating: '',
    category: mappedCategories[0] as 'movie' | 'game' | 'table_game' | 'other' || 'other',
    group: groupData.name,
    onPress: () => {
      console.log('Переход на сессию:', item.session.id);
    }
  })) || [];

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <View style={styles.header}>
          <Image source={{ uri: groupData.image }} style={styles.groupImage} />
          <View style={{ flexDirection: 'column', flex: 1 }}>
            <View style={styles.headerInfo}>
              <Text style={styles.groupName}>{groupData.name}</Text>
              <Text style={styles.location}>{groupData.city}</Text>
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
          {groupData.sessions && groupData.sessions.length > 0 ? (
            <EventCarousel events={formattedSessions} />
          ) : (
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>Пока нет активных сессий</Text>
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
    backgroundColor: Colors.lightBlue3,
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
  applicationsContainer: {
    paddingHorizontal: 18,
    paddingVertical: 12,
  },
  applicationItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightGrey,
  },
  applicationImage: {
    width: 50,
    height: 50,
    borderRadius: 25,
    marginRight: 12,
  },
  applicationInfo: {
    flex: 1,
  },
  applicationName: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.black,
  },
  applicationUsername: {
    fontFamily: Montserrat.regular,
    fontSize: 12,
    color: Colors.grey,
  },
  applicationActions: {
    flexDirection: 'row',
  },
  approveButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: Colors.lightBlue3,
    justifyContent: 'center',
    alignItems: 'center',
    marginLeft: 8,
  },
  approveButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.white,
  },
  rejectButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: Colors.red,
    justifyContent: 'center',
    alignItems: 'center',
    marginLeft: 8,
  },
  rejectButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.white,
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
});

export default GroupPage;