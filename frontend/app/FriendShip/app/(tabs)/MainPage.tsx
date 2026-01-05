import permissionsService from '@/api/services/permissionsService';
import sessionService from '@/api/services/session';
import { AdminGroup } from '@/api/types/auth';
import BottomBar from '@/components/BottomBar';
import CategorySection from '@/components/CategorySection';
import CreateEventButton from '@/components/event/CreateEventButton';
import { Event } from '@/components/event/EventCard';
import CreateEditEventModal from '@/components/event/modal/CreateEventModal';
import EventModal from '@/components/event/modal/EventModal';
import VerticalEventList from '@/components/event/VerticalEventList';
import GroupSelectorModal from '@/components/groups/modal/GroupSelectorModal';
import PageHeader from '@/components/PageHeader';
import PermissionsModal from '@/components/permissions/PermissionsModal';
import SearchResultsSection from '@/components/search/SearchResultsSection';
import { useToast } from '@/components/ToastContext';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useHeaderAnimation } from '@/contexts/HeaderContext';
import { categoryToSessionType } from '@/hooks/groups/groupManageHelpers';
import { useEvents } from '@/hooks/useEvents';
import { useSearchState } from '@/hooks/useSearchState';
import { useThemedColors } from '@/hooks/useThemedColors';
import { RootStackParamList } from '@/navigation/types';
import { RouteProp, useFocusEffect, useNavigation, useRoute } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ActivityIndicator, Animated, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

type MainPageRouteProp = RouteProp<RootStackParamList, 'MainPage'>;
type MainPageNavigationProp = NativeStackNavigationProp<RootStackParamList, 'MainPage'>;

const MainPage = () => {
  const colors = useThemedColors();
  const route = useRoute<MainPageRouteProp>();
  const navigation = useNavigation<MainPageNavigationProp>();

  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [groupSelectorVisible, setGroupSelectorVisible] = useState(false);
  const [createEventModalVisible, setCreateEventModalVisible] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState<AdminGroup | null>(null);
  const [isCreatingEvent, setIsCreatingEvent] = useState(false);
  const [permissionsModalVisible, setPermissionsModalVisible] = useState(false);
  const { showToast } = useToast();
  const { hasPlayedAnimation, markAnimationPlayed } = useHeaderAnimation();

  const headerHeight = useRef(new Animated.Value(hasPlayedAnimation ? 0 : 1)).current;
  const headerOpacity = useRef(new Animated.Value(hasPlayedAnimation ? 0 : 1)).current;

  const { sortingState, sortingActions } = useSearchState();
  const { 
    movieEvents, 
    gameEvents, 
    tableGameEvents, 
    otherEvents, 
    popularEvents, 
    newEvents, 
    searchResults,
    isLoading,
    error,
    refreshEvents,
  } = useEvents(sortingState);

  const isFirstRender = useRef(true);

  useFocusEffect(
    useCallback(() => {
      if (isFirstRender.current) {
        isFirstRender.current = false;
        return;
      }

      console.log('[MainPage] 🔄 Страница получила фокус, обновляем события');
      refreshEvents();
    }, [])
  );

  useEffect(() => {
    const checkAndRequestPermissions = async () => {
      try {
        const hasRequested = await permissionsService.hasRequestedPermissions();
        
        if (!hasRequested) {
          console.log('[MainPage] 🔐 Первый запуск - показываем модалку разрешений');
          setTimeout(() => {
            setPermissionsModalVisible(true);
          }, 1000);
        } else {
          console.log('[MainPage] ✅ Разрешения уже запрашивались ранее');
        }
      } catch (error) {
        console.error('[MainPage] Ошибка проверки разрешений:', error);
      }
    };

    checkAndRequestPermissions();
  }, []);

  useEffect(() => {
    if (hasPlayedAnimation) {
      console.log('[MainPage] ⏭️ Анимация уже была проиграна, пропускаем');
      return;
    }
    
    if (!sortingState.searchQuery.trim() && !isLoading && !error) {
      const timer = setTimeout(() => {
        console.log('[MainPage] 🎬 Запускаем анимацию скрытия заголовка');
        Animated.parallel([
          Animated.timing(headerHeight, {
            toValue: 0,
            duration: 500,
            useNativeDriver: false,
          }),
          Animated.timing(headerOpacity, {
            toValue: 0,
            duration: 400,
            useNativeDriver: true,
          }),
        ]).start(() => {
          console.log('[MainPage] ✅ Анимация завершена');
          markAnimationPlayed();
        });
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [sortingState.searchQuery, isLoading, error, hasPlayedAnimation, markAnimationPlayed]);

  const handleRequestPermissions = async () => {
    try {
      console.log('[MainPage] 🚀 Пользователь согласился на разрешения');
      setPermissionsModalVisible(false);
      
      const status = await permissionsService.requestInitialPermissions();
      
      if (!status.media) {
        showToast({
          type: 'warning',
          title: 'Внимание',
          message: 'Доступ к фото не предоставлен. Вы можете разрешить его позже в настройках',
        });
      }
      
      if (!status.notifications) {
        showToast({
          type: 'warning',
          title: 'Внимание',
          message: 'Уведомления отключены. Вы можете включить их позже в настройках',
        });
      }
      
      if (status.media && status.notifications) {
        showToast({
          type: 'success',
          title: 'Отлично!',
          message: 'Все разрешения предоставлены',
        });
      }
    } catch (error) {
      console.error('[MainPage] Ошибка запроса разрешений:', error);
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Не удалось запросить разрешения',
      });
    }
  };

  const handleSkipPermissions = async () => {
    console.log('[MainPage] ⏭️ Пользователь пропустил разрешения');
    setPermissionsModalVisible(false);
    await permissionsService.setPermissionsRequested();
    
    showToast({
      type: 'success',
      title: 'Разрешения пропущены',
      message: 'Вы можете предоставить их позже в настройках приложения',
    });
  };

  useEffect(() => {
    if (route.params?.searchQuery) {
      sortingActions.setSearchQuery(route.params.searchQuery);
    }
  }, [route.params?.searchQuery]);

  const hasAnyEvents = useMemo(() => {
    return (
      popularEvents.length > 0 ||
      newEvents.length > 0 ||
      movieEvents.length > 0 ||
      gameEvents.length > 0 ||
      tableGameEvents.length > 0 ||
      otherEvents.length > 0
    );
  }, [popularEvents, newEvents, movieEvents, gameEvents, tableGameEvents, otherEvents]);

  const addOnPressToEvents = (events: Event[]) =>
    events.map(event => ({
      ...event,
      onPress: () => {
        setSelectedEvent(event);
        setModalVisible(true);
      },
      onSessionUpdate: refreshEvents,
    }));

  const navigateToCategory = (
    category: 'movie' | 'game' | 'table_game' | 'other' | 'popular' | 'new',
    title: string,
  ) => {
    try {
      navigation.navigate('CategoryPage', {
        category,
        title
      });
    } catch (error) {
      console.error('❌ Navigation error:', error);
    }
  };

  const getEventWordForm = (count: number): string => {
    if (count % 10 === 1 && count % 100 !== 11) {
      return 'событие';
    } else if ([2, 3, 4].includes(count % 10) && ![12, 13, 14].includes(count % 100)) {
      return 'события';
    } else {
      return 'событий';
    }
  };
  
  const handleCreateEventPress = () => {
    setGroupSelectorVisible(true);
  };

  const handleSelectGroup = (group: AdminGroup) => {
    console.log('[MainPage] Выбрана группа для создания события:', group.name);
    setSelectedGroup(group);
    setGroupSelectorVisible(false);
    setCreateEventModalVisible(true);
  };

  const handleCreateEvent = async (eventData: any) => {
    if (!selectedGroup) {
      showToast({
        type: 'error',
        title: 'Ошибка',
        message: 'Группа не выбрана',
      });
      return;
    }

    try {
      setIsCreatingEvent(true);
      console.log('[MainPage] 📤 Создание события для группы:', selectedGroup.id);
      
      const sessionData = {
        title: eventData.title,
        session_type: categoryToSessionType[eventData.category] || 'Другое',
        session_place: eventData.typePlace === 'online' ? 1 : 2,
        group_id: selectedGroup.id,
        start_time: eventData.start_time,
        count_users: eventData.maxParticipants,
        image: eventData.image,
        duration: eventData.duration,
        genres: eventData.genres || [],
        location: eventData.eventPlace || '',
        year: eventData.year,
        country: eventData.country || eventData.publisher || '',
        age_limit: eventData.ageRating || '',
        notes: eventData.description || '',
        fields: eventData.category,
      };

      console.log('[MainPage] 📤 Отправляемые данные:', sessionData);
      
      await sessionService.createSession(sessionData);
      
      showToast({
        type: 'success',
        title: 'Успешно',
        message: 'Событие создано!',
      });
      
      setCreateEventModalVisible(false);
      setSelectedGroup(null);
      
      } catch (error: any) {
        console.error('[MainPage] ❌ Ошибка создания события:', error);
        
        let errorMessage = 'Не удалось создать событие';
        
        if (error.response?.data?.error) {
          errorMessage = error.response.data.error;
        } else if (error.message) {
          errorMessage = error.message;
        }
        
        showToast({
          type: 'error',
          title: 'Ошибка',
          message: errorMessage,
        });
      } finally {
        setIsCreatingEvent(false);
      }
  };

  const handleCloseCreateEventModal = () => {
    setCreateEventModalVisible(false);
    setSelectedGroup(null);
  };

  if (isLoading) {
    return (
      <SafeAreaView style={[styles.container, { backgroundColor: colors.white }]}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color={Colors.lightBlue} />
          <Text style={[styles.loadingText, { color: colors.black }]}>
            Загрузка событий...
          </Text>
        </View>
        <BottomBar />
      </SafeAreaView>
    );
  }

  if (error) {
    return (
      <SafeAreaView style={[styles.container, { backgroundColor: colors.white }]}>
        <TopBar sortingState={sortingState} sortingActions={sortingActions} />
        <View style={styles.centerContainer}>
          <Text style={styles.errorTitle}>
            Ошибка загрузки
          </Text>
          <Text style={[styles.errorText, { color: colors.black }]}>
            {error}
          </Text>
        </View>
        <BottomBar />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: colors.white }]}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />

      {sortingState.searchQuery.trim() ? (
        <View style={{ flex: 1 }}>
          <SearchResultsSection
            title="Поиск по событиям"
            searchQuery={sortingState.searchQuery}
            hasResults={searchResults.length > 0}
          >
            {searchResults.length > 0 && (
              <View style={styles.resultsHeader}>
                <Text style={styles.resultsCount}>
                  Найдено: {searchResults.length} {getEventWordForm(searchResults.length)}
                </Text>
              </View>
            )}

            <View style={styles.searchResultsContainer}>
              <VerticalEventList events={addOnPressToEvents(searchResults)} />
            </View>
          </SearchResultsSection>
        </View>
      ) : (
        <ScrollView contentContainerStyle={styles.scrollContent}>
          <Animated.View 
            style={{ 
              height: headerHeight.interpolate({
                inputRange: [0, 1],
                outputRange: [0, 120],
              }),
              opacity: headerOpacity,
              overflow: 'hidden',
            }}
          >
            <PageHeader title="С возвращением!" showWave />
          </Animated.View>


          {!hasAnyEvents ? (
            <View style={styles.noEventsContainer}>
              <Text style={[styles.noEventsTitle, { color: Colors.blue2 }]}>
                Пока нет доступных событий
              </Text>
              <Text style={[styles.noEventsSubtext, { color: Colors.grey }]}>
                Создайте первое событие в вашей группе,{'\n'}
                чтобы оно появилось здесь
              </Text>
            </View>
          ) : (
            <>
              {popularEvents.length > 0 && (
                <CategorySection
                  title="Популярные:"
                  events={addOnPressToEvents(popularEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'popular',
                      'Популярные события'
                    )
                  }
                />
              )}

              {newEvents.length > 0 && (
                <CategorySection
                  title="Новые:"
                  events={addOnPressToEvents(newEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'new',
                      'Новые события'
                    )
                  }
                />
              )}

              <CategorySection
                title="Категории"
                events={[]}
                centerTitle
                showLineVariant="line2"
                marginBottom={16}
              />

              {movieEvents.length > 0 && (
                <CategorySection
                  title="Медиа:"
                  events={addOnPressToEvents(movieEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'movie',
                      'Медиа'
                    )
                  }
                />
              )}

              {gameEvents.length > 0 && (
                <CategorySection
                  title="Видеоигры:"
                  events={addOnPressToEvents(gameEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'game',
                      'Видеоигры'
                    )
                  }
                />
              )}

              {tableGameEvents.length > 0 && (
                <CategorySection
                  title="Настольные игры:"
                  events={addOnPressToEvents(tableGameEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'table_game',
                      'Настольные игры'
                    )
                  }
                />
              )}

              {otherEvents.length > 0 && (
                <CategorySection
                  title="Другое:"
                  events={addOnPressToEvents(otherEvents)}
                  showArrow
                  onArrowPress={() =>
                    navigateToCategory(
                      'other',
                      'Другое'
                    )
                  }
                />
              )}
            </>
          )}
        </ScrollView>
      )}

      <BottomBar />

      <CreateEventButton onPress={handleCreateEventPress} />

      <GroupSelectorModal
        visible={groupSelectorVisible}
        onClose={() => setGroupSelectorVisible(false)}
        onSelectGroup={handleSelectGroup}
      />

      {selectedGroup && (
        <CreateEditEventModal
          visible={createEventModalVisible}
          onClose={handleCloseCreateEventModal}
          onCreate={handleCreateEvent}
          groupName={selectedGroup.name}
          editMode={false}
          isLoading={isCreatingEvent}
        />
      )}

      {selectedEvent && (
        <EventModal
          visible={modalVisible}
          onClose={() => setModalVisible(false)}
          event={selectedEvent}
          onSessionUpdate={refreshEvents}
        />
      )}

      <PermissionsModal
        visible={permissionsModalVisible}
        onRequestPermissions={handleRequestPermissions}
        onSkip={handleSkipPermissions}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 80,
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  loadingText: {
    marginTop: 12,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  errorTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.red,
    marginBottom: 8,
  },
  errorText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    textAlign: 'center',
  },
  resultsHeader: {
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  resultsCount: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
  },
  searchResultsContainer: {
    flex: 1,
    paddingHorizontal: 16,
  },
  noEventsContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
    paddingVertical: 60,
  },
  noEventsTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    textAlign: 'center',
    marginBottom: 12,
  },
  noEventsSubtext: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    textAlign: 'center',
    lineHeight: 24,
  },
});

export default MainPage;