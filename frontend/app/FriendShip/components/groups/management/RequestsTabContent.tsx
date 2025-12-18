import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import { useThemedColors } from '@/hooks/useThemedColors';
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
  isProcessing?: boolean;
}

const RequestsTabContent: React.FC<RequestsTabContentProps> = ({
  pendingRequests,
  searchQuery,
  setSearchQuery,
  onAcceptRequest,
  onRejectRequest,
  onAcceptAll,
  onRejectAll,
  isProcessing = false,
}) => {
  const colors = useThemedColors();

  const filteredRequests = pendingRequests.filter(request =>
    request.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    request.username.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <ScrollView
      style={styles.tabContent}
      showsVerticalScrollIndicator={false}
    >
      <View style={[styles.requestsContent, { backgroundColor: colors.white }]}>
        <View style={styles.requestsHeader}>
          <View style={styles.counterContainer}>
            <Text style={[styles.counterNumber, { color: colors.black }]}>
              {pendingRequests.length}
            </Text>
            <Text style={[styles.counterText, { color: colors.grey }]}>
              нерассмотренных заявок
            </Text>
          </View>
          
          {pendingRequests.length > 0 && (
            <View style={styles.actionButtons}>
              <TouchableOpacity 
                style={[
                  styles.acceptAllButton,
                  { backgroundColor: colors.lightBlue },
                  isProcessing && styles.buttonDisabled
                ]} 
                onPress={onAcceptAll}
                disabled={isProcessing}
              >
                <Text style={styles.acceptAllText}>Принять все</Text>
              </TouchableOpacity>
              <TouchableOpacity 
                style={[styles.rejectAllButton, isProcessing && styles.buttonDisabled]} 
                onPress={onRejectAll}
                disabled={isProcessing}
              >
                <Text style={styles.rejectAllText}>Отклонить все</Text>
              </TouchableOpacity>
            </View>
          )}
        </View>

        {pendingRequests.length > 0 && (
          <View style={styles.searchContainer}>
            <TextInput
              style={[
                styles.searchInput,
                {
                  backgroundColor: colors.veryLightGrey,
                  color: colors.black
                }
              ]}
              placeholder="Найти пользователя..."
              placeholderTextColor={colors.grey}
              value={searchQuery}
              onChangeText={setSearchQuery}
              editable={!isProcessing}
            />
          </View>
        )}

        {filteredRequests.length > 0 ? (
          <View style={styles.requestsList}>
            {filteredRequests.map((request) => (
              <View key={request.id} style={[styles.requestItem, { backgroundColor: colors.card }]}>
                <Image source={{ uri: request.imageUri }} style={styles.userAvatar} />
                
                <View style={styles.userInfo}>
                  <Text style={[styles.userName, { color: colors.black }]}>
                    {request.name}
                  </Text>
                  <Text style={[styles.userUsername, { color: colors.grey }]}>
                    @{request.username}
                  </Text>
                </View>
                
                <View style={styles.requestActions}>
                  <TouchableOpacity 
                    style={styles.acceptButton}
                    onPress={() => onAcceptRequest(request.id)}
                    disabled={isProcessing}
                  >
                    <Image 
                      source={require('@/assets/images/groups/check.png')} 
                      style={[styles.actionIcon, isProcessing && styles.iconDisabled]}
                    />
                  </TouchableOpacity>
                  
                  <TouchableOpacity 
                    style={styles.rejectButton}
                    onPress={() => onRejectRequest(request.id)}
                    disabled={isProcessing}
                  >
                    <Image 
                      source={require('@/assets/images/groups/close.png')} 
                      style={[styles.actionIcon, isProcessing && styles.iconDisabled]}
                    />
                  </TouchableOpacity>
                </View>
              </View>
            ))}
          </View>
        ) : pendingRequests.length === 0 ? (
          <View style={styles.emptyState}>
            <Text style={[styles.emptyStateText, { color: colors.grey }]}>
              Нет заявок на рассмотрение
            </Text>
          </View>
        ) : (
          <View style={styles.emptyState}>
            <Text style={[styles.emptyStateText, { color: colors.grey }]}>
              Пользователи не найдены
            </Text>
          </View>
        )}
      </View>
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  tabContent: {
    flex: 1,
  },
  requestsContent: {
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
    fontFamily: Montserrat_Alternates.bold,
    fontSize: 26,
  },
  counterText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  actionButtons: {
    flexDirection: 'row',
    gap: 16,
  },
  acceptAllButton: {
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
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.white,
  },
  rejectAllText: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.white,
  },
  searchContainer: {
    marginBottom: 20,
  },
  searchInput: {
    borderRadius: 40,
    paddingHorizontal: 16,
    paddingVertical: 8,
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
  requestsList: {
    borderRadius: 12,
    overflow: 'visible',
  },
  requestItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 8,
    marginBottom: 12,
    borderRadius: 30,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
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
    fontFamily: Montserrat.bold,
    fontSize: 16,
  },
  userUsername: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
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
  buttonDisabled: {
    opacity: 0.5,
  },
  iconDisabled: {
    opacity: 0.5,
  },
  emptyState: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 40,
  },
  emptyStateText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
  },
});

export default RequestsTabContent;