import { Colors } from '@/constants/Colors';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import { Image, Modal, Pressable, StyleSheet, TouchableOpacity, View } from 'react-native';

interface SearchTypeModalProps {
  visible: boolean;
  onClose: () => void;
  position: { top: number; left: number };
  activeType: 'event' | 'profile' | 'group';
  onTypeChange: (type: 'event' | 'profile' | 'group') => void;
}

const SEARCH_TYPE_ICONS = {
  event: {
    arrow: require('@/assets/images/top_bar/search_bar/event-arrow.png'),
    bar: require('@/assets/images/top_bar/search_bar/event-bar.png'),
  },
  profile: {
    arrow: require('@/assets/images/top_bar/search_bar/profile-arrow.png'),
    bar: require('@/assets/images/top_bar/search_bar/profile.png'),
  },
  group: {
    arrow: require('@/assets/images/top_bar/search_bar/group-arrow.png'),
    bar: require('@/assets/images/top_bar/search_bar/group.png'),
  },
};

const SearchTypeModal: React.FC<SearchTypeModalProps> = ({
  visible,
  onClose,
  position,
  activeType,
  onTypeChange,
}) => {
  const colors = useThemedColors();

  const getIconSource = (type: 'event' | 'profile' | 'group') => {
    return type === activeType 
      ? SEARCH_TYPE_ICONS[type].arrow 
      : SEARCH_TYPE_ICONS[type].bar;
  };

  return (
    <Modal
      transparent
      animationType="fade"
      visible={visible}
      onRequestClose={onClose}
    >
      <Pressable style={styles.modalOverlay} onPress={onClose}>
        <View
          style={[
            styles.categoryModal,
            { 
              top: position.top, 
              left: position.left,
              backgroundColor: colors.white
            }
          ]}
        >
          <View>
            <Image
              source={SEARCH_TYPE_ICONS[activeType].arrow}
              style={[styles.iconOnly, {tintColor: Colors.lightGreyBlue}]}
              resizeMode="contain"
            />
          </View>

          {(['event', 'profile', 'group'] as const)
            .filter(type => type !== activeType)
            .map((type) => (
              <TouchableOpacity key={type} onPress={() => onTypeChange(type)}>
                <Image 
                  source={getIconSource(type)} 
                  style={[styles.iconOnly, {tintColor: Colors.lightGreyBlue}]} 
                  resizeMode="contain" 
                />
              </TouchableOpacity>
            ))}
        </View>
      </Pressable>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    justifyContent: 'flex-start',
    alignItems: 'flex-start',
  },
  categoryModal: {
    position: 'absolute',
    borderRadius: 10,
    padding: 10,
    elevation: 5,
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 2,
  },
  iconOnly: {
    width: 25,
    height: 25,
    marginVertical: 4,
  },
});

export default SearchTypeModal;