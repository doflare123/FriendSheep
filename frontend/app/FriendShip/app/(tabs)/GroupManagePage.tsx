import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import React, { useState } from 'react';
import {
  Image,
  StyleSheet,
  Text,
  TouchableOpacity,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import BottomBar from '@/components/BottomBar';
import ConfirmationModal from '@/components/ConfirmationModal';
import EventsTabContent from '@/components/groups/management/EventsTabContent';
import GroupManageTabPanel from '@/components/groups/management/GroupManageTabPanel';
import InfoTabContent from '@/components/groups/management/InfoTabContent';
import RequestsTabContent from '@/components/groups/management/RequestsTabContent';
import ContactsModal from '@/components/groups/modal/ContactsModal';
import TopBar from '@/components/TopBar';

import CreateEditEventModal from '@/components/event/modal/CreateEventModal';
import { GroupAdminGuard } from '@/components/groups/management/GroupAdminGuard';
import SubscribersTabContent from '@/components/groups/management/SubscribersTabContent';
import { Montserrat } from '@/constants/Montserrat';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import { useGroupManage } from '@/hooks/groups/useGroupManage';
import { useSearchState } from '@/hooks/useSearchState';
import { useThemedColors } from '@/hooks/useThemedColors';
import { RootStackParamList } from '@/navigation/types';

type GroupManagePageRouteProp = RouteProp<RootStackParamList, 'GroupManagePage'>;
type GroupManagePageNavigationProp = StackNavigationProp<RootStackParamList, 'GroupManagePage'>;

const GroupManagePage = () => {
  const colors = useThemedColors();
  const route = useRoute<GroupManagePageRouteProp>();
  const navigation = useNavigation<GroupManagePageNavigationProp>();
  const { sortingState, sortingActions } = useSearchState();
  const { groupId } = route.params;
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const {
    activeTab,
    setActiveTab,
    
    groupData,
    isLoading,
    isSaving,
    groupName,
    setGroupName,
    shortDescription,
    setShortDescription,
    fullDescription,
    setFullDescription,
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

    availableGenres,
    isCreatingEvent,
    isUpdatingEvent,
    createEventModalVisible,
    setCreateEventModalVisible,
    editEventModalVisible,
    setEditEventModalVisible,
    selectedEventId,
    deleteGroup,
    selectedEventData,
    handleCreateEvent,
    handleEditEvent,
    handleCreateEventSave,
    handleEditEventSave,
    handleDeleteEvent,
    formattedEvents,   
    formattedSubscribers,
    subscriberSearchQuery,
    setSubscriberSearchQuery,
    isProcessingSubscriber,
    removeMemberModal,
    handleRemoveMemberPress,
    handleConfirmRemoveMember,
    handleCancelRemoveMember,
    handleUserPress: handleUserPressFromHook,
  } = useGroupManage(groupId);

  const handleBackPress = () => {
    navigation.goBack();
  };

  const handleUserPress = (userId: string) => {
    console.log('[GroupManagePage] Переход к профилю пользователя:', userId);
    navigation.navigate('ProfilePage', { userId });
  };

  const handleDeletePress = () => {
    setShowDeleteModal(true);
  };

  const handleConfirmDelete = async () => {
    try {
      const success = await deleteGroup();
      if (success) {
        setShowDeleteModal(false);
        navigation.reset({
          index: 0,
          routes: [{ name: 'GroupsPage' }],
        });
      }
    } catch (error) {
      console.error('[GroupManagePage] ❌ Ошибка при удалении группы:', error);
      setShowDeleteModal(false);
    }
  };

  const handleCancelDelete = () => {
    setShowDeleteModal(false);
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
            isSaving={isSaving}
            selectedContacts={selectedContacts}
          />
        );
      
      case 'subscribers':
        return (
          <SubscribersTabContent
            subscribers={formattedSubscribers}
            searchQuery={subscriberSearchQuery}
            setSearchQuery={setSubscriberSearchQuery}
            onRemoveMember={handleRemoveMemberPress}
            onUserPress={handleUserPress}
            isProcessing={isProcessingSubscriber}
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
          <EventsTabContent
            events={formattedEvents}
            onCreateEvent={handleCreateEvent}
            onEditEvent={handleEditEvent}
          />
        );
      
      default:
        return null;
    }
  };

  if (!groupData) {
    return (
      <SafeAreaView style={[styles.container, { backgroundColor: colors.white }]}>
        <View style={styles.errorContainer}>
          <Text style={[styles.errorText, { color: colors.black }]}>
            Группа не найдена
          </Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <GroupAdminGuard groupId={groupId}>
      <SafeAreaView style={[styles.container, { backgroundColor: colors.white }]}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
          <View style={styles.sectionHeader}>
            <GroupManageTabPanel
              activeTab={activeTab}
              onTabChange={setActiveTab}
              isPrivateGroup={isPrivate}
            />
            
            <TouchableOpacity style={styles.backButton} onPress={handleBackPress}>
              <Image
                source={require('../../assets/images/event_card/back.png')}
                style={[styles.backIcon, { tintColor: colors.black }]}
              />
            </TouchableOpacity>
          </View>
          
          <View style={styles.titleContainer}>
            <Text style={[styles.sectionTitle, { color: colors.black }]}>
              {getSectionTitle()}
            </Text>
            {activeTab === 'info' && (
              <TouchableOpacity
                style={styles.deleteIconButton}
                onPress={handleDeletePress}
                disabled={isSaving}
              >
                <Image
                  source={require('@/assets/images/groups/delete.png')}
                  style={styles.deleteIcon}
                />
              </TouchableOpacity>
            )}
          </View>
          
        {renderTabContent()}

        <ContactsModal
          visible={contactsModalVisible}
          onClose={() => setContactsModalVisible(false)}
          onSave={handleContactsSave}
          initialContacts={groupData?.contacts || []}
        />
        
        <ConfirmationModal
          visible={confirmationModal.visible}
          title={confirmationModal.title}
          message={confirmationModal.message}
          onConfirm={confirmAction}
          onCancel={cancelConfirmation}
        />

        <ConfirmationModal
          visible={showDeleteModal}
          title="Удалить группу?"
          message="Это действие нельзя отменить. Все участники, события и заявки будут безвозвратно удалены."
          onConfirm={handleConfirmDelete}
          onCancel={handleCancelDelete}
        />

      <ConfirmationModal
        visible={removeMemberModal.visible}
        title="Удалить участника?"
        message={`Пользователь ${removeMemberModal.userName} будет исключён из группы`}
        onConfirm={handleConfirmRemoveMember}
        onCancel={handleCancelRemoveMember}
      />

        <CreateEditEventModal
          visible={createEventModalVisible}
          onClose={() => setCreateEventModalVisible(false)}
          onCreate={handleCreateEventSave}
          groupName={groupData?.name || 'Группа'}
          availableGenres={availableGenres}
          isLoading={isCreatingEvent}
          editMode={false}
        />

        <CreateEditEventModal
          visible={editEventModalVisible}
          onClose={() => setEditEventModalVisible(false)}
          onUpdate={handleEditEventSave}
          onDelete={handleDeleteEvent}
          groupName={groupData?.name || 'Группа'}
          editMode={true}
          initialData={selectedEventData}
          availableGenres={availableGenres}
          isLoading={isUpdatingEvent}
        />
        
        <BottomBar />
      </SafeAreaView>
    </GroupAdminGuard>
  );
};

const styles = StyleSheet.create({
  container: {
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
  },
  topBackground: {
    width: '100%',
  },
  tabContent: {
    flex: 1,
  },
  titleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 16,
    position: 'relative',
    marginBottom: 20,
  },
  sectionTitle: {
    fontFamily: Montserrat_Alternates.bold,
    fontSize: 20,
  },
  deleteIconButton: {
    position: 'absolute',
    right: 16,
    padding: 8,
  },
  deleteIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
  },
});

export default GroupManagePage;