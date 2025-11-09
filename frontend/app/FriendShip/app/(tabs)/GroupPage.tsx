import groupService, { GroupDetailResponse } from '@/api/services/groupService';
import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import { Event as EventType } from '@/components/event/EventCard';
import EventCarousel from '@/components/event/EventCarousel';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useFocusEffect, useNavigation, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useCallback, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Dimensions,
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

type GroupManagePageNavigationProp = StackNavigationProp<
  RootStackParamList,
  'GroupManagePage'
>;

const { width } = Dimensions.get('window');

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
};

const GroupPage = () => {
  const route = useRoute<GroupPageRouteProp>();
  const { groupId, mode } = route.params;
  const [descriptionModalVisible, setDescriptionModalVisible] = useState(false);
  const [groupData, setGroupData] = useState<GroupDetailResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const { sortingState, sortingActions } = useSearchState();
  const navigation = useNavigation<GroupManagePageNavigationProp>();

  const loadGroupData = async () => {
    try {
      setIsLoading(true);
      const data = await groupService.getGroupDetail(groupId);

      if (data.image && data.image.includes('localhost')) {
        data.image = data.image.replace('http://localhost:8080', 'http://192.168.0.209:8080');
      }

      if (data.sessions) {
        data.sessions = data.sessions.map(session => ({
          ...session,
          image_url: session.image_url?.includes('localhost')
            ? session.image_url.replace('http://localhost:8080', 'http://192.168.0.209:8080')
            : session.image_url
        }));
      }
      
      setGroupData(data);
    } catch (error: any) {
      console.error('Ошибка загрузки группы:', error);
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

  const handleContactPress = (link: string) => {
    if (link) {
      Linking.openURL(link).catch(err => {
        console.error('Не удалось открыть ссылку:', err);
        Alert.alert('Ошибка', 'Не удалось открыть ссылку');
      });
    }
  };

  const getContactIcon = (contactName: string) => {
    const lowerName = contactName.toLowerCase();
    for (const key in contactIcons) {
      if (lowerName.includes(key)) {
        return contactIcons[key];
      }
    }
    return contactIcons.max;
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

  const formattedSessions: EventType[] = groupData.sessions?.map(session => ({
    id: session.id.toString(),
    title: session.title,
    date: new Date(session.start_time).toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    }),
    genres: session.genres || [],
    currentParticipants: session.current_users,
    maxParticipants: session.count_users_max,
    duration: `${session.duration} мин`,
    imageUri: session.image_url,
    description: '',
    typeEvent: session.session_type,
    typePlace: session.session_place === 'offline' || session.session_place === 'online' 
      ? session.session_place as 'online' | 'offline'
      : 'online',
    eventPlace: session.city || '',
    publisher: groupData.name,
    publicationDate: session.start_time,
    ageRating: '',
    category: mappedCategories[0] as 'movie' | 'game' | 'table_game' | 'other' || 'other',
    group: groupData.name,
    onPress: () => {
      console.log('Переход на сессию:', session.id);
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
              onPress={() => {
                if (mode === 'manage') {
                  navigation.navigate('GroupManagePage', { groupId });
                } else {
                  console.log('Присоединиться');
                }
              }}
              style={styles.actionButton}
            >
              <Text style={styles.actionButtonText}>
                {mode === 'manage' ? 'Управлять' : 'Присоединиться'}
              </Text>
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
              {groupData.contacts.map((contact, index) => (
                <TouchableOpacity
                  key={index}
                  style={styles.contactItem}
                  onPress={() => handleContactPress(contact.link)}
                >
                  <View style={styles.contactIconContainer}>
                    <Image 
                      source={getContactIcon(contact.name)} 
                      style={styles.contactIcon} 
                    />
                  </View>
                  <Text style={styles.contactDescription} numberOfLines={1}>
                    {contact.name}
                  </Text>
                </TouchableOpacity>
              ))}
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
    padding: 40,
    alignItems: 'center',
  },
  emptyText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    textAlign: 'center',
    marginBottom: 8,
  },
});

export default GroupPage;