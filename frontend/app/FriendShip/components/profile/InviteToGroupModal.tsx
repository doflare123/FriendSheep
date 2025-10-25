import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import {
    Dimensions,
    FlatList,
    Image,
    Modal,
    StyleSheet,
    Text,
    TouchableOpacity,
    TouchableWithoutFeedback,
    View
} from 'react-native';

const screenHeight = Dimensions.get("window").height;

interface Group {
  id: string;
  name: string;
  imageUri: any;
  participantsCount: number;
}

interface InviteToGroupModalProps {
  visible: boolean;
  onClose: () => void;
  onGroupSelect: (groupId: string) => void;
}

const myGroups: Group[] = [
  {
    id: '1',
    name: 'Моя супер группа',
    imageUri: require('@/assets/images/profile/profile_avatar.jpg'),
    participantsCount: 25,
  },
  {
    id: '2',
    name: 'Любители кино',
    imageUri: require('@/assets/images/profile/profile_avatar.jpg'),
    participantsCount: 48,
  },
  {
    id: '3',
    name: 'Геймеры',
    imageUri: require('@/assets/images/profile/profile_avatar.jpg'),
    participantsCount: 15,
  },
  {
    id: '4',
    name: 'Настольные игры',
    imageUri: require('@/assets/images/profile/profile_avatar.jpg'),
    participantsCount: 12,
  },
];

const InviteToGroupModal: React.FC<InviteToGroupModalProps> = ({
  visible,
  onClose,
  onGroupSelect,
}) => {
  const handleGroupPress = (groupId: string) => {
    onGroupSelect(groupId);
  };

  const renderGroupItem = ({ item }: { item: Group }) => (
    <TouchableOpacity
      style={styles.groupItem}
      onPress={() => handleGroupPress(item.id)}
      activeOpacity={0.7}
    >
      <Image source={item.imageUri} style={styles.groupImage} />
      <View style={styles.groupInfo}>
        <Text style={styles.groupName} numberOfLines={1}>
          {item.name}
        </Text>
        <Text style={styles.participantsText}>
          Участники: {item.participantsCount}
        </Text>
      </View>
      <Image 
        source={require('@/assets/images/arrow.png')} 
        style={styles.arrowIcon}
      />
    </TouchableOpacity>
  );

  return (
    <Modal visible={visible} animationType="fade" transparent>
      <TouchableWithoutFeedback onPress={onClose}>
        <View style={styles.overlay}>
          <TouchableWithoutFeedback>
            <View style={styles.modal}>
              <View style={styles.header}>
                <Text style={styles.title}>Выберите группу</Text>
                <TouchableOpacity style={styles.closeButton} onPress={onClose}>
                  <Image
                    tintColor={Colors.black}
                    style={{ width: 35, height: 35, resizeMode: 'cover' }}
                    source={require('@/assets/images/event_card/back.png')}
                  />
                </TouchableOpacity>
              </View>

              <View style={styles.content}>
                {myGroups.length > 0 ? (
                  <FlatList
                    data={myGroups}
                    keyExtractor={(item) => item.id}
                    renderItem={renderGroupItem}
                    showsVerticalScrollIndicator={false}
                  />
                ) : (
                  <View style={styles.emptyContainer}>
                    <Text style={styles.emptyText}>
                      У вас пока нет групп
                    </Text>
                  </View>
                )}
              </View>
            </View>
          </TouchableWithoutFeedback>
        </View>
      </TouchableWithoutFeedback>
    </Modal>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  modal: {
    marginHorizontal: 6,
    borderRadius: 25,
    overflow: "hidden",
    backgroundColor: Colors.white,
    maxHeight: screenHeight * 0.7,
    width: '90%',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 14,
    paddingHorizontal: 16,
    backgroundColor: Colors.white,
    position: 'relative',
    borderBottomWidth: 1,
    borderBottomColor: Colors.veryLightGrey,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 20,
    color: Colors.black,
    textAlign: 'center',
  },
  closeButton: {
    position: 'absolute',
    right: 10,
    zIndex: 1,
  },
  content: {
    backgroundColor: Colors.white,
    paddingVertical: 8,
    maxHeight: screenHeight * 0.5,
  },
  groupItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: Colors.veryLightGrey,
  },
  groupImage: {
    width: 50,
    height: 50,
    borderRadius: 25,
    marginRight: 12,
  },
  groupInfo: {
    flex: 1,
  },
  groupName: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.black,
    marginBottom: 4,
  },
  participantsText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.grey,
  },
  arrowIcon: {
    width: 20,
    height: 20,
    tintColor: Colors.grey,
  },
  emptyContainer: {
    paddingVertical: 40,
    alignItems: 'center',
  },
  emptyText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.grey,
  },
});

export default InviteToGroupModal;