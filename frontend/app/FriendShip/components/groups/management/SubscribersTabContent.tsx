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

interface SubscriberItem {
  id: string;
  name: string;
  imageUri: string;
  role: string;
}

interface SubscribersTabContentProps {
  subscribers: SubscriberItem[];
  searchQuery: string;
  setSearchQuery: (value: string) => void;
  onRemoveMember: (userId: string) => void;
  onUserPress: (userId: string) => void;
  isProcessing?: boolean;
}

const SubscribersTabContent: React.FC<SubscribersTabContentProps> = ({
  subscribers,
  searchQuery,
  setSearchQuery,
  onRemoveMember,
  onUserPress,
  isProcessing = false,
}) => {
  const colors = useThemedColors();

  const filteredSubscribers = subscribers.filter(subscriber =>
    subscriber.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getRoleBadgeColor = (role: string) => {
    switch (role.toLowerCase()) {
      case 'admin':
        return Colors.red;
      case 'operator':
        return Colors.lightBlue;
      default:
        return Colors.lightGrey;
    }
  };

  const getRoleText = (role: string) => {
    switch (role.toLowerCase()) {
      case 'admin':
        return 'Админ';
      case 'operator':
        return 'Оператор';
      default:
        return 'Участник';
    }
  };

  return (
    <ScrollView
      style={styles.tabContent}
      showsVerticalScrollIndicator={false}
    >
      <View style={[styles.content, { backgroundColor: colors.white }]}>
        <View style={styles.header}>
          <View style={styles.counterContainer}>
            <Text style={[styles.counterNumber, { color: colors.black }]}>
              {subscribers.length}
            </Text>
            <Text style={[styles.counterText, { color: colors.grey }]}>
              участников
            </Text>
          </View>
        </View>

        {subscribers.length > 0 && (
          <View style={styles.searchContainer}>
            <TextInput
              style={[
                styles.searchInput,
                {
                  backgroundColor: colors.veryLightGrey,
                  color: colors.black
                }
              ]}
              placeholder="Найти участника..."
              placeholderTextColor={colors.grey}
              value={searchQuery}
              onChangeText={setSearchQuery}
              editable={!isProcessing}
            />
          </View>
        )}

        {filteredSubscribers.length > 0 ? (
          <View style={styles.subscribersList}>
            {filteredSubscribers.map((subscriber) => (
              <TouchableOpacity
                key={subscriber.id}
                style={[styles.subscriberItem, { backgroundColor: colors.card }]}
                onPress={() => onUserPress(subscriber.id)}
                activeOpacity={0.7}
                disabled={isProcessing}
              >
                <Image source={{ uri: subscriber.imageUri }} style={styles.userAvatar} />
                
                <View style={styles.userInfo}>
                  <Text style={[styles.userName, { color: colors.black }]}>
                    {subscriber.name}
                  </Text>
                  <View style={[styles.roleBadge, { backgroundColor: getRoleBadgeColor(subscriber.role) }]}>
                    <Text style={styles.roleText}>{getRoleText(subscriber.role)}</Text>
                  </View>
                </View>
                
                {subscriber.role.toLowerCase() !== 'admin' && (
                  <TouchableOpacity 
                    style={styles.removeButton}
                    onPress={(e) => {
                      e.stopPropagation();
                      onRemoveMember(subscriber.id);
                    }}
                    disabled={isProcessing}
                  >
                    <Image 
                      source={require('@/assets/images/groups/close.png')} 
                      style={[styles.removeIcon, isProcessing && styles.iconDisabled]}
                    />
                  </TouchableOpacity>
                )}
              </TouchableOpacity>
            ))}
          </View>
        ) : subscribers.length === 0 ? (
          <View style={styles.emptyState}>
            <Text style={[styles.emptyStateText, { color: colors.grey }]}>
              Нет участников
            </Text>
          </View>
        ) : (
          <View style={styles.emptyState}>
            <Text style={[styles.emptyStateText, { color: colors.grey }]}>
              Участники не найдены
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
  content: {
    paddingHorizontal: 16,
    paddingTop: 20,
    minHeight: '100%',
  },
  header: {
    alignItems: 'center',
    marginBottom: 24,
  },
  counterContainer: {
    alignItems: 'center',
  },
  counterNumber: {
    fontFamily: Montserrat_Alternates.bold,
    fontSize: 26,
  },
  counterText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
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
  subscribersList: {
    borderRadius: 12,
    overflow: 'visible',
  },
  subscriberItem: {
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
    marginBottom: 4,
  },
  roleBadge: {
    alignSelf: 'flex-start',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 12,
  },
  roleText: {
    fontFamily: Montserrat.bold,
    fontSize: 12,
    color: Colors.white,
  },
  removeButton: {
    width: 36,
    height: 36,
    borderRadius: 18,
    justifyContent: 'center',
    alignItems: 'center',
  },
  removeIcon: {
    width: 40,
    height: 40,
    resizeMode: 'contain',
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

export default SubscribersTabContent;