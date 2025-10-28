import barsStyle from '@/app/styles/barsStyle';
import MainSearchBar from '@/components/search/MainSearchBar';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { SortingActions, SortingState } from '@/hooks/useSearchState';
import React, { useRef, useState } from 'react';
import { Image, Modal, Pressable, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

interface TopBarProps {
  sortingState: SortingState;
  sortingActions: SortingActions;
}

const TopBar : React.FC<TopBarProps> = ({ sortingState, sortingActions }) => {
  const [notificationsVisible, setNotificationsVisible] = useState(false);
  const [notifModalPos, setNotifModalPos] = useState({ top: 0, left: 0 });
  const notifIconRef = useRef<View>(null);

  const openNotifications = () => {
    notifIconRef.current?.measureInWindow((x, y, width, height) => {
      const modalWidth = 320;
      const centerX = x + width / 2;
      const left = centerX - modalWidth + 5;
      const top = y + (height * 0.5);

      setNotifModalPos({ top, left });
      setNotificationsVisible(true);
    });
  };

  return (
    <View style={styles.container}>
      <MainSearchBar 
        sortingState={sortingState} 
        sortingActions={sortingActions} 
      />
      <TouchableOpacity
        ref={notifIconRef}
        onPress={openNotifications}
        activeOpacity={0.7}
      >
        <Image style={barsStyle.notifications} source={require("@/assets/images/top_bar/notifications.png")} />
      </TouchableOpacity>

      <Modal
        transparent
        animationType="fade"
        visible={notificationsVisible}
        onRequestClose={() => setNotificationsVisible(false)}
      >
        <Pressable style={styles.modalOverlay} onPress={() => setNotificationsVisible(false)}>
          <Pressable
            style={[styles.notificationsModal, { top: notifModalPos.top, left: notifModalPos.left }]}
            onPress={() => {}}
          >
            <Text style={styles.title}>У вас N непрочитанных уведомлений:</Text>

            <View style={styles.notificationItem}>
              <View style={styles.iconWrapper}>
                <Image
                  style={styles.notificationIcon}
                  source={{ uri: 'https://cdn-icons-png.flaticon.com/512/2921/2921222.png' }}
                />
              </View>
              <Text style={styles.notificationText}>
                Событие “Какое-то событие” начнётся через 30 мин
              </Text>
              <Text style={styles.notificationTime}>1 час назад</Text>
            </View>

            <View style={styles.notificationItem}>
              <View style={styles.iconWrapper}>
                <Image
                  style={styles.notificationIcon}
                  source={{ uri: 'https://cdn-icons-png.flaticon.com/512/2921/2921222.png' }}
                />
              </View>
              <View style={{ flex: 1 }}>
                <Text style={styles.notificationText}>
                  Вас приглашают в группу “МЕГАКРУТЫЕ ПАЦАНЯТА”
                </Text>
                <View style={styles.buttonsRow}>
                  <TouchableOpacity style={styles.acceptButton}>
                    <Text style={styles.buttonText}>Принять</Text>
                  </TouchableOpacity>
                  <TouchableOpacity style={styles.rejectButton}>
                    <Text style={styles.buttonText}>Отклонить</Text>
                  </TouchableOpacity>
                </View>
              </View>
              <Text style={styles.notificationTime}>1 час назад</Text>
            </View>
          </Pressable>
        </Pressable>
      </Modal>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.white,
    paddingHorizontal: 16,
    paddingVertical: 8,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    zIndex: 10,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: Colors.lightBlack,
  },
  notificationsModal: {
    position: 'absolute',
    width: 320,
    backgroundColor: Colors.white,
    borderRadius: 12,
    padding: 16,
    borderTopEndRadius: 0,
    shadowColor: Colors.black,
    shadowRadius: 10,
    elevation: 10,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 12,
    fontWeight: '600',
    marginBottom: 12,
  },
  notificationItem: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    marginBottom: 16,
  },
  iconWrapper: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: Colors.red,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  notificationIcon: {
    width: 32,
    height: 32,
  },
  notificationText: {
    fontFamily: Montserrat.regular,
    flex: 1,
    fontSize: 11,
  },
  notificationTime: {
    fontFamily: Montserrat.regular,
    fontSize: 10,
    color: Colors.lightGrey,
    marginLeft: 8,
    alignSelf: 'flex-end'
  },
  buttonsRow: {
    flexDirection: 'row',
    marginTop: 8,
  },
  acceptButton: {
    backgroundColor: Colors.lightBlue,
    paddingHorizontal: 12,
    paddingVertical: 2,
    borderRadius: 12,
    marginRight: 8,
  },
  rejectButton: {
    backgroundColor: Colors.red,
    paddingVertical: 2,
    paddingHorizontal: 8,
    borderRadius: 12,
  },
  buttonText: {
    fontFamily: Montserrat.regular,
    color: Colors.white,
    fontSize: 10,
  },
});

export default TopBar;
