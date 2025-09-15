import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React from 'react';
import {
  Image,
  ImageBackground,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import ConfirmationModal from '@/components/ConfirmationModal';
import ContactsModal from '@/components/ContactsModal';
import GroupManageTabPanel from '@/components/GroupManageTabPanel';
import InfoTabContent from '@/components/InfoTabContent';
import RequestsTabContent from '@/components/RequestsTabContent';
import TopBar from '@/components/TopBar';

import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import { useGroupManage } from '@/hooks/useGroupManage';
import { useSearchState } from '@/hooks/useSearchState';
import { RootStackParamList } from '@/navigation/types';

type GroupManagePageRouteProp = RouteProp<RootStackParamList, 'GroupManagePage'>;
type GroupManagePageNavigationProp = StackNavigationProp<RootStackParamList, 'GroupManagePage'>;

const GroupManagePage = () => {
  const route = useRoute<GroupManagePageRouteProp>();
  const navigation = useNavigation<GroupManagePageNavigationProp>();
  const { sortingState, sortingActions } = useSearchState();
  const { groupId } = route.params;

  const {
    activeTab,
    setActiveTab,
    
    groupData,
    groupName,
    setGroupName,
    shortDescription,
    setShortDescription,
    fullDescription,
    setFullDescription,
    country,
    setCountry,
    city,
    setCity,
    isPrivate,
    setIsPrivate,
    selectedCategories,
    selectedContacts,
    groupImage,
    setGroupImage,

    toggleCategory,
    handleContactsPress,
    handleContactsSave,
    contactsModalVisible,
    setContactsModalVisible,

    pendingRequests,
    searchQuery,
    setSearchQuery,
    handleAcceptRequest,
    handleRejectRequest,
    handleAcceptAll,
    handleRejectAll,

    confirmationModal,
    confirmAction,
    cancelConfirmation,

    handleSaveChanges,
    getSectionTitle,
  } = useGroupManage(groupId);

  const handleBackPress = () => {
    navigation.goBack();
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'info':
        return (
          <InfoTabContent
            groupName={groupName}
            setGroupName={setGroupName}
            shortDescription={shortDescription}
            setShortDescription={setShortDescription}
            fullDescription={fullDescription}
            setFullDescription={setFullDescription}
            country={country}
            setCountry={setCountry}
            city={city}
            setCity={setCity}
            isPrivate={isPrivate}
            setIsPrivate={setIsPrivate}
            selectedCategories={selectedCategories}
            toggleCategory={toggleCategory}
            groupImage={groupImage}
            setGroupImage={setGroupImage}
            onContactsPress={handleContactsPress}
            onSaveChanges={handleSaveChanges}
          />
        );
      
      case 'requests':
        return (
          <RequestsTabContent
            pendingRequests={pendingRequests}
            searchQuery={searchQuery}
            setSearchQuery={setSearchQuery}
            onAcceptRequest={handleAcceptRequest}
            onRejectRequest={handleRejectRequest}
            onAcceptAll={handleAcceptAll}
            onRejectAll={handleRejectAll}
          />
        );
      
      case 'events':
        return (
          <View style={styles.tabContent}>
            <Text style={styles.comingSoonText}>Управление событиями - скоро</Text>
          </View>
        );
      
      default:
        return null;
    }
  };

  if (!groupData) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>Группа не найдена</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      
      <ImageBackground
        source={require('../../assets/images/groups/top_rectangle.png')}
        style={styles.topBackground}
        resizeMode='stretch'
      >
        <View style={styles.sectionHeader}>
          <GroupManageTabPanel
            activeTab={activeTab}
            onTabChange={setActiveTab}
          />
          
          <TouchableOpacity style={styles.backButton} onPress={handleBackPress}>
            <Image
              source={require('../../assets/images/event_card/back.png')}
              style={styles.backIcon}
              tintColor={Colors.darkGrey}
            />
          </TouchableOpacity>
        </View>
        
        <Text style={styles.sectionTitle}>{getSectionTitle()}</Text>
      </ImageBackground>
        
      {renderTabContent()}

      <ContactsModal
        visible={contactsModalVisible}
        onClose={() => setContactsModalVisible(false)}
        onSave={handleContactsSave}
      />
      
      <ConfirmationModal
        visible={confirmationModal.visible}
        title={confirmationModal.title}
        message={confirmationModal.message}
        onConfirm={confirmAction}
        onCancel={cancelConfirmation}
      />
      
      <BottomBar />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.white,
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  errorText: {
    fontFamily: inter.regular,
    fontSize: 16,
    color: Colors.black,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    width: '100%',
    paddingHorizontal: 16,
    paddingVertical: 20,
    paddingBottom: 20,
    marginTop: 0,
    overflow: 'visible',
  },
  backButton: {
    position: 'absolute',
    right: 16,
    top: '50%',
    marginTop: -10,
    justifyContent: 'center',
    alignItems: 'center',
  },
  backIcon: {
    width: 40,
    height: 40,
    resizeMode: 'contain',
    tintColor: Colors.black,
  },
  topBackground: {
    width: '100%',
  },
  sectionTitle: {
    fontFamily: inter.black,
    fontSize: 20,
    color: Colors.black,
    textAlign: 'center',
  },
  tabContent: {
    flex: 1,
  },
  comingSoonText: {
    fontFamily: inter.regular,
    fontSize: 16,
    color: Colors.grey,
    textAlign: 'center',
    marginTop: 50,
  },
});

export default GroupManagePage;