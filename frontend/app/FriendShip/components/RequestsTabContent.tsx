import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React from 'react';
import {
    Image,
    ScrollView,
    StyleSheet,
    Text,
    TextInput,
    TouchableOpacity,
    View,
} from 'react-native';

interface RequestItem {
  id: string;
  name: string;
  username: string;
  imageUri: string;
}

interface RequestsTabContentProps {
  pendingRequests: RequestItem[];
  searchQuery: string;
  setSearchQuery: (value: string) => void;
  onAcceptRequest: (userId: string) => void;
  onRejectRequest: (userId: string) => void;
  onAcceptAll: () => void;
  onRejectAll: () => void;
}

const RequestsTabContent: React.FC<RequestsTabContentProps> = ({
  pendingRequests,
  searchQuery,
  setSearchQuery,
  onAcceptRequest,
  onRejectRequest,
  onAcceptAll,
  onRejectAll,
}) => {
  const filteredRequests = pendingRequests.filter(request =>
    request.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    request.username.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <ScrollView
      style={styles.tabContent}
      showsVerticalScrollIndicator={false}
    >
      <View style={styles.requestsContent}>
        <View style={styles.requestsHeader}>
          <View style={styles.counterContainer}>
            <Text style={styles.counterNumber}>{pendingRequests.length}</Text>
            <Text style={styles.counterText}>нерассмотренных заявок</Text>
          </View>
          
          <View style={styles.actionButtons}>
            <TouchableOpacity style={styles.acceptAllButton} onPress={onAcceptAll}>
              <Text style={styles.acceptAllText}>Принять все</Text>
            </TouchableOpacity>
            <TouchableOpacity style={styles.rejectAllButton} onPress={onRejectAll}>
              <Text style={styles.rejectAllText}>Отклонить все</Text>
            </TouchableOpacity>
          </View>
        </View>

        <View style={styles.searchContainer}>
          <TextInput
            style={styles.searchInput}
            placeholder="Найти пользователя..."
            placeholderTextColor={Colors.grey}
            value={searchQuery}
            onChangeText={setSearchQuery}
          />
        </View>

        <View style={styles.requestsList}>
          {filteredRequests.map((request) => (
            <View key={request.id} style={styles.requestItem}>
              <Image source={{ uri: request.imageUri }} style={styles.userAvatar} />
              
              <View style={styles.userInfo}>
                <Text style={styles.userName}>{request.name}</Text>
                <Text style={styles.userUsername}>{request.username}</Text>
              </View>
              
              <View style={styles.requestActions}>
                <TouchableOpacity 
                  style={styles.acceptButton}
                  onPress={() => onAcceptRequest(request.id)}
                >
                  <Image 
                    source={require('../assets/images/groups/check.png')} 
                    style={styles.actionIcon}
                  />
                </TouchableOpacity>
                
                <TouchableOpacity 
                  style={styles.rejectButton}
                  onPress={() => onRejectRequest(request.id)}
                >
                  <Image 
                    source={require('../assets/images/groups/close.png')} 
                    style={styles.actionIcon}
                  />
                </TouchableOpacity>
              </View>
            </View>
          ))}
        </View>
      </View>
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  tabContent: {
    flex: 1,
  },
  requestsContent: {
    backgroundColor: Colors.white,
    paddingHorizontal: 16,
    paddingTop: 20,
    minHeight: '100%',
  },
  requestsHeader: {
    alignItems: 'center',
    marginBottom: 24,
  },
  counterContainer: {
    alignItems: 'center',
    marginBottom: 20,
  },
  counterNumber: {
    fontFamily: inter.black,
    fontSize: 26,
    color: Colors.black,
  },
  counterText: {
    fontFamily: inter.regular,
    fontSize: 16,
    color: Colors.grey,
  },
  actionButtons: {
    flexDirection: 'row',
    gap: 16,
  },
  acceptAllButton: {
    backgroundColor: Colors.lightBlue,
    paddingHorizontal: 16,
    paddingVertical: 4,
    borderRadius: 20,
  },
  rejectAllButton: {
    backgroundColor: Colors.red,
    paddingHorizontal: 16,
    paddingVertical: 4,
    borderRadius: 20,
  },
  acceptAllText: {
    fontFamily: inter.bold,
    fontSize: 14,
    color: Colors.white,
  },
  rejectAllText: {
    fontFamily: inter.bold,
    fontSize: 14,
    color: Colors.white,
  },
  searchContainer: {
    marginBottom: 20,
  },
  searchInput: {
    backgroundColor: Colors.veryLightGrey,
    borderRadius: 40,
    paddingHorizontal: 16,
    paddingVertical: 8,
    fontFamily: inter.regular,
    fontSize: 16,
    color: Colors.black,
  },
  requestsList: {
    borderWidth: 2,
    borderColor: Colors.lightBlue,
    borderRadius: 12,
    backgroundColor: Colors.white,
    overflow: 'hidden',
  },
  requestItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: Colors.lightLightGrey,
  },
  userAvatar: {
    width: 55,
    height: 55,
    borderRadius: 30,
    marginRight: 12,
  },
  userInfo: {
    flex: 1,
  },
  userName: {
    fontFamily: inter.bold,
    fontSize: 16,
    color: Colors.black,
  },
  userUsername: {
    fontFamily: inter.regular,
    fontSize: 14,
    color: Colors.grey,
  },
  requestActions: {
    flexDirection: 'row',
    gap: 8,
  },
  acceptButton: {
    width: 36,
    height: 36,
    borderRadius: 18,
    justifyContent: 'center',
    alignItems: 'center',
  },
  rejectButton: {
    width: 36,
    height: 36,
    borderRadius: 18,
    justifyContent: 'center',
    alignItems: 'center',
  },
  actionIcon: {
    width: 40,
    height: 40,
    resizeMode: 'contain',
  },
});

export default RequestsTabContent;