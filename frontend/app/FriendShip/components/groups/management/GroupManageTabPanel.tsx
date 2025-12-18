import { Colors } from '@/constants/Colors';
import { useThemedColors } from '@/hooks/useThemedColors';
import { LinearGradient } from 'expo-linear-gradient';
import React from 'react';
import { Image, StyleSheet, TouchableOpacity, View } from 'react-native';

export type TabType = 'info' | 'subscribers' | 'requests' | 'events';

interface GroupManageTabPanelProps {
  activeTab: TabType;
  onTabChange: (tab: TabType) => void;
  isPrivateGroup: boolean;
}

const GroupManageTabPanel: React.FC<GroupManageTabPanelProps> = ({
  activeTab,
  onTabChange,
  isPrivateGroup,
}) => {
  const colors = useThemedColors();

  const renderTabButton = (tab: TabType, icon: any) => (
    <TouchableOpacity
      key={tab}
      style={styles.tabButton}
      onPress={() => onTabChange(tab)}
      activeOpacity={0.7}
    >
      <LinearGradient
        colors={[Colors.darkGrey, Colors.lightGreyBlue]}
        start={{ x: 0, y: 0 }}
        end={{ x: 0, y: 1 }}
        style={styles.gradientBorder}
      >
        <View style={[
          isPrivateGroup ? styles.innerButtonPrivate : styles.innerButtonPublic,
          { backgroundColor: colors.veryLightGrey },
          activeTab === tab
        ]}>
          <Image 
            source={icon} 
            style={[
              styles.tabIcon, {tintColor: Colors.lightGreyBlue},
              activeTab === tab && {tintColor: Colors.darkGrey},
            ]} 
          />
        </View>
      </LinearGradient>
    </TouchableOpacity>
  );

  return (
    <View style={[
        isPrivateGroup ? styles.tabsContainerPrivate : styles.tabsContainerPublic
      ]}>
      {renderTabButton('info', require('@/assets/images/groups/panel/edit.png'))}
      {renderTabButton('subscribers', require('@/assets/images/groups/panel/group.png'))}
      {isPrivateGroup && renderTabButton('requests', require('@/assets/images/groups/panel/person_add.png'))}
      {renderTabButton('events', require('@/assets/images/groups/panel/event.png'))}
    </View>
  );
};

const styles = StyleSheet.create({
  tabsContainerPublic: {
    flexDirection: 'row',
    marginHorizontal: 45,
    marginTop: -40,
  },
  tabsContainerPrivate:{
    flexDirection: 'row',
    marginHorizontal: 35,
    marginTop: -40,
  },
  tabButton: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: 20,
    marginHorizontal: 4,
    minHeight: 75,
    overflow: 'hidden',
  },
  gradientBorder: {
    flex: 1,
    borderRadius: 20,
    padding: 3,
  },
  innerButtonPrivate: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'flex-end',
    borderRadius: 18,
    paddingVertical: 12,
    paddingHorizontal: 4,
    paddingBottom: 4,
  },
  innerButtonPublic: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'flex-end',
    borderRadius: 18,
    paddingVertical: 12,
    paddingHorizontal: 16,
    paddingBottom: 4,
  },
  tabIcon: {
    width: 40,
    height: 40,
    resizeMode: 'contain',
  },
});

export default GroupManageTabPanel;