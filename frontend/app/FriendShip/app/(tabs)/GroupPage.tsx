import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import EventCarousel from '@/components/event/EventCarousel';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { Contact, getGroupData, Subscriber } from '@/data/groupsData';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useState } from 'react';
import {
  Dimensions,
  FlatList,
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
const SUBSCRIBER_ITEM_WIDTH = 80;
const SUBSCRIBER_SPACING = 16;

const categoryIcons: Record<string, any> = {
  movie: require('../../assets/images/event_card/movie.png'),
  game: require('../../assets/images/event_card/game.png'),
  table_game: require('../../assets/images/event_card/table_game.png'),
  other: require('../../assets/images/event_card/other.png'),
};

const GroupPage = () => {
  const route = useRoute<GroupPageRouteProp>();
  const { groupId, mode } = route.params;
  const [descriptionModalVisible, setDescriptionModalVisible] = useState(false);
  const { sortingState, sortingActions } = useSearchState();
  const navigation = useNavigation<GroupManagePageNavigationProp>();
  
  const groupData = getGroupData(groupId);
  
  if (!groupData) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>Группа не найдена</Text>
        </View>
      </SafeAreaView>
    );
  }

  const handleContactPress = (contact: Contact) => {
    if (contact.link) {
      Linking.openURL(contact.link);
    }
  };

  const renderSubscriber = ({ item, index }: { item: Subscriber; index: number }) => (
    <View
      style={[
        styles.subscriberItem,
        {
          marginLeft: index === 0 ? SUBSCRIBER_SPACING : 0,
          marginRight: SUBSCRIBER_SPACING,
        },
      ]}
    >
      <Image source={{ uri: item.imageUri }} style={styles.subscriberImage} />
      <Text style={styles.subscriberName} numberOfLines={1}>
        {item.name}
      </Text>
    </View>
  );

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        <View style={styles.header}>
          <Image source={{ uri: groupData.imageUri }} style={styles.groupImage} />
          <View style={{flexDirection: 'column'}}>
            <View style={styles.headerInfo}>
              <Text style={styles.groupName}>{groupData.name}</Text>
              <Text style={styles.location}>
                {groupData.city}, {groupData.country}
              </Text>
              <View style={styles.categoriesContainer}>
                {groupData.categories.map((category, index) => (
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
              {groupData.description}
            </Text>
          </TouchableOpacity>
        </CategorySection>

        <CategorySection title={`Подписчики: ${groupData.subscribersCount.toLocaleString()}`}>
          <FlatList
            horizontal
            data={groupData.subscribers}
            renderItem={renderSubscriber}
            keyExtractor={(item) => item.id}
            showsHorizontalScrollIndicator={false}
            snapToInterval={SUBSCRIBER_ITEM_WIDTH + SUBSCRIBER_SPACING}
            decelerationRate="fast"
            style={styles.subscribersList}
          />
        </CategorySection>

        <CategorySection title="Сессии:">
          <EventCarousel events={groupData.sessions} />
        </CategorySection>

        <CategorySection title="Контакты:">
          <View style={styles.contactsContainer}>
            {groupData.contacts.map((contact) => (
              <TouchableOpacity
                key={contact.id}
                style={styles.contactItem}
                onPress={() => handleContactPress(contact)}
              >
                <View style={styles.contactIconContainer}>
                  <Image source={contact.icon} style={styles.contactIcon} />
                </View>
                <Text style={styles.contactDescription}>{contact.description}</Text>
              </TouchableOpacity>
            ))}
          </View>
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
  subscribersList: {
    paddingVertical: 16,
  },
  subscriberItem: {
    alignItems: 'center',
    width: SUBSCRIBER_ITEM_WIDTH,
  },
  subscriberImage: {
    width: 80,
    height: 80,
    borderRadius: 100,
  },
  subscriberName: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    textAlign: 'center',
  },
  contactsContainer: {
    paddingHorizontal: 18,
    paddingVertical: 12,
    flexDirection: 'row',
    justifyContent: 'space-around',
  },
  contactItem: {
    alignItems: 'center',
    flex: 1,
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
    fontSize: 14,
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
});

export default GroupPage;