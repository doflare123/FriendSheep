import BottomBar from '@/components/BottomBar';
import PageHeader from '@/components/PageHeader';
import TopBar from '@/components/TopBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useSearchState } from '@/hooks/useSearchState';
import { useThemedColors } from '@/hooks/useThemedColors';
import { useNavigation } from '@react-navigation/native';
import React from 'react';
import {
    Image,
    Linking,
    ScrollView,
    StyleSheet,
    Text,
    TouchableOpacity,
    View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

const AboutPage = () => {
  const { sortingState, sortingActions } = useSearchState();
  const colors = useThemedColors();
  const navigation = useNavigation();

  const handleFeedbackPress = () => {
    Linking.openURL('https://docs.google.com/forms/d/e/1FAIpQLScq8yseDrHN2dQ7eTfon6KqiohGzPAE95FRoyh8KkaFWuTB9Q/viewform');
  };

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: colors.white }]}>
      <TopBar sortingState={sortingState} sortingActions={sortingActions} />
      
      <ScrollView 
        style={styles.scrollView} 
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        <PageHeader 
          title="О приложении" 
          showWave 
          showBackButton
          onBackPress={() => navigation.goBack()}
        />

        <View style={[styles.section, { backgroundColor: colors.card }]}>
          <View style={styles.sectionHeader}>
            <Image
              source={require('@/assets/images/logo.png')}
              style={styles.appIcon}
            />
            <View style={styles.appInfo}>
              <Text style={[styles.appName, { color: colors.black }]}>
                FriendShip
              </Text>
              <Text style={[styles.appVersion, { color: Colors.grey }]}>
                Версия 1.0.0
              </Text>
            </View>
          </View>
          
          <Text style={[styles.description, { color: colors.black }]}>
            <Text style={styles.bold}>FriendShip</Text> — это социальная платформа для организации совместных мероприятий с друзьями. Создавайте группы по интересам, планируйте события и находите единомышленников для совместного просмотра фильмов, игр, настольных игр и других активностей.{'\n\n'}
            
            Приложение помогает легко координировать встречи, отслеживать участников и управлять расписанием событий в удобном формате.
          </Text>
        </View>

        <View style={[styles.section, { backgroundColor: colors.card }]}>
          <Text style={[styles.sectionTitle, { color: colors.blue2 }]}>
            Основные возможности
          </Text>
          
          <View style={styles.featureItem}>
            <View style={styles.featureText}>
              <Text style={[styles.featureTitle, { color: colors.black }]}>
                👥 Группы по интересам
              </Text>
              <Text style={[styles.featureDescription, { color: colors.black }]}>
                Создавайте публичные или приватные группы для фильмов, игр, настольных игр и других активностей
              </Text>
            </View>
          </View>

          <View style={styles.featureItem}>
            <View style={styles.featureText}>
              <Text style={[styles.featureTitle, { color: colors.black }]}>
                📅 События и встречи
              </Text>
              <Text style={[styles.featureDescription, { color: colors.black }]}>
                Планируйте онлайн и оффлайн события с указанием места, времени и максимального количества участников
              </Text>
            </View>
          </View>

          <View style={styles.featureItem}>
            <View style={styles.featureText}>
              <Text style={[styles.featureTitle, { color: colors.black }]}>
                🔔 Уведомления
              </Text>
              <Text style={[styles.featureDescription, { color: colors.black }]}>
                Получайте push-уведомления о новых событиях и приглашениях в группы
              </Text>
            </View>
          </View>

          <View style={styles.featureItem}>
            <View style={styles.featureText}>
              <Text style={[styles.featureTitle, { color: colors.black }]}>
                🎮 Интеграция с базами данных
              </Text>
              <Text style={[styles.featureDescription, { color: colors.black }]}>
                Автоматическое заполнение информации о фильмах (Кинопоиск) и играх (RAWG)
              </Text>
            </View>
          </View>

          <View style={styles.featureItem}>
            <View style={styles.featureText}>
              <Text style={[styles.featureTitle, { color: colors.black }]}>
                📊 Статистика
              </Text>
              <Text style={[styles.featureDescription, { color: colors.black }]}>
                Отслеживайте свою активность, любимые жанры и время, проведённое за различными мероприятиями
              </Text>
            </View>
          </View>
        </View>

        <View style={[styles.section, { backgroundColor: colors.card }]}>
          <Text style={[styles.sectionTitle, { color: colors.blue2 }]}>
            Как пользоваться
          </Text>
          
          <Text style={[styles.instructionTitle, { color: colors.black }]}>
            1. Создание группы
          </Text>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>1.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Перейдите на вкладку "Группы"
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>2.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Нажмите кнопку "+" справа от "Группы под управлением"
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>3.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Заполните информацию: название, краткое описание, полное описание, город
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>4.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Выберите тип группы (публичная или приватная)
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>5.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Укажите категории группы (медиа, видеоигры, настольные игры, другое)
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>6.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Добавьте контакты для связи
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>7.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Загрузите изображение группы
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>8.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Сохраните изменения
            </Text>
          </View>

          <Text style={[styles.instructionTitle, { color: colors.black }]}>
            2. Создание события
          </Text>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>1.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              На главной странице нажмите кнопку "+" слева внизу экрана
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>2.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Выберите группу, в которой создаёте событие
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>3.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Заполните название и описание
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>4.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Укажите категорию (медиа, видеоигры, настольные игры, другое)
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>5.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Укажите тип события (оффлайн или онлайн)
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>6.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Выберите жанры для события
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>7.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Заполните издателя, год издания и возрастной рейтинг
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>8.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Выберите дату, время и место проведения
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>9.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Укажите максимальное количество участников
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>10.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Загрузите изображение и создайте событие
            </Text>
          </View>

          <Text style={[styles.instructionTitle, { color: colors.black }]}>
            3. Присоединение к группе
          </Text>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>1.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Найдите интересующую группу через поиск
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>2.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Откройте страницу группы
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>3.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Нажмите "Присоединиться"
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>4.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Для публичных групп — вступление мгновенное
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>5.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Для приватных групп — дождитесь одобрения администратора
            </Text>
          </View>

          <Text style={[styles.instructionTitle, { color: colors.black }]}>
            4. Участие в событии
          </Text>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>1.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Найдите интересное событие на главной странице
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>2.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Откройте карточку события
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>3.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Нажмите "Присоединиться"
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>4.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Следите за уведомлениями о начале события
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>5.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              После завершения события оно появится в вашей статистике
            </Text>
          </View>

          <Text style={[styles.instructionTitle, { color: colors.black }]}>
            5. Управление профилем
          </Text>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>1.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Откройте свой профиль через нижнюю панель
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>2.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Нажмите на кнопку шестерёнку возле аватара
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>3.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Выберите "Редактировать профиль" для изменения данных
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>4.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Настройте отображаемую статистику через "Сменить плитки"
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>5.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Просматривайте свои завершённые и предстоящие события
            </Text>
          </View>
          <View style={styles.listItem}>
            <Text style={[styles.bullet, { color: Colors.lightBlue }]}>6.</Text>
            <Text style={[styles.listText, { color: colors.black }]}>
              Отслеживайте любимые жанры и активность
            </Text>
          </View>

        <Text style={[styles.instructionTitle, { color: colors.black }]}>
        6. Использование поиска
        </Text>
        <View style={styles.listItem}>
        <Text style={[styles.bullet, { color: Colors.lightBlue }]}>1.</Text>
        <Text style={[styles.listText, { color: colors.black }]}>
            Нажмите на иконку слева в поисковой строке для выбора типа поиска
        </Text>
        </View>
        <View style={styles.listItem}>
        <Text style={[styles.bullet, { color: Colors.lightBlue }]}>2.</Text>
        <Text style={[styles.listText, { color: colors.black }]}>
            Выберите тип: события, пользователи или группы
        </Text>
        </View>
        <View style={styles.listItem}>
        <Text style={[styles.bullet, { color: Colors.lightBlue }]}>3.</Text>
        <Text style={[styles.listText, { color: colors.black }]}>
            Введите запрос в строку поиска
        </Text>
        </View>
        <View style={styles.listItem}>
        <Text style={[styles.bullet, { color: Colors.lightBlue }]}>4.</Text>
        <Text style={[styles.listText, { color: colors.black }]}>
            Нажмите на иконку фильтра справа для дополнительной настройки
        </Text>
        </View>
        <View style={styles.listItem}>
        <Text style={[styles.bullet, { color: Colors.lightBlue }]}>5.</Text>
        <Text style={[styles.listText, { color: colors.black }]}>
            Для событий: фильтруйте по городу, категориям, дате и количеству участников
        </Text>
        </View>
        <View style={styles.listItem}>
        <Text style={[styles.bullet, { color: Colors.lightBlue }]}>6.</Text>
        <Text style={[styles.listText, { color: colors.black }]}>
            Для групп: сортируйте по участникам, категориям и дате регистрации
        </Text>
        </View>
        <View style={styles.listItem}>
        <Text style={[styles.bullet, { color: Colors.lightBlue }]}>7.</Text>
        <Text style={[styles.listText, { color: colors.black }]}>
            Нажмите Enter или иконку поиска для отображения результатов
        </Text>
        </View>          
        </View>

        <View style={[styles.section, { backgroundColor: colors.card }]}>
        <Text style={[styles.sectionTitle, { color: Colors.blue2 }]}>
            Поддержка и обратная связь
        </Text>
        
        <Text style={[styles.contactText, { color: colors.black }]}>
            По вопросам работы приложения, предложениям и сообщениям об ошибках:
        </Text>

        <TouchableOpacity 
            style={styles.feedbackButton}
            onPress={handleFeedbackPress}
            activeOpacity={0.7}
        >
            <Text style={styles.feedbackButtonText}>Форма обратной связи</Text>
        </TouchableOpacity>

        <TouchableOpacity 
            style={[styles.feedbackButton, { marginTop: 8 }]}
            onPress={() => Linking.openURL('https://friendsheep.ru')}
            activeOpacity={0.7}
        >
            <Text style={styles.feedbackButtonText}>Наш сайт</Text>
        </TouchableOpacity>
        </View>
      </ScrollView>

      <BottomBar />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 100,
  },
  section: {
    marginHorizontal: 16,
    marginTop: 16,
    padding: 20,
    borderRadius: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 6,
    elevation: 2,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  appIcon: {
    width: 60,
    height: 60,
    borderRadius: 12,
    marginRight: 12,
  },
  appInfo: {
    flex: 1,
  },
  appName: {
    fontFamily: Montserrat.bold,
    fontSize: 24,
    marginBottom: 4,
  },
  appVersion: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  description: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    lineHeight: 24,
  },
  bold: {
    fontFamily: Montserrat.bold,
  },
  sectionTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    marginBottom: 16,
  },
  featureItem: {
    marginBottom: 16,
  },
  featureText: {
    flex: 1,
  },
  featureTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    marginBottom: 4,
  },
  featureDescription: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    lineHeight: 20,
  },
  instructionTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    marginTop: 16,
    marginBottom: 12,
  },
  listItem: {
    flexDirection: 'row',
    marginBottom: 8,
    alignItems: 'flex-start',
  },
  bullet: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    marginLeft: 8,
    width: 24,
  },
  listText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    lineHeight: 20,
    flex: 1,
  },
  contactText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    lineHeight: 24,
    marginBottom: 16,
  },
  feedbackButton: {
    backgroundColor: Colors.lightBlue,
    paddingVertical: 8,
    paddingHorizontal: 24,
    borderRadius: 30,
    alignItems: 'center',
    marginTop: 8,
  },
  feedbackButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
});

export default AboutPage;