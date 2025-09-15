import { Colors } from '@/constants/Colors';
import { LinearGradient } from 'expo-linear-gradient';
import React from 'react';
import { Image, StyleSheet, TouchableOpacity, View } from 'react-native';

export type TabType = 'info' | 'requests' | 'events';

interface GroupManageTabPanelProps {
  activeTab: TabType;
  onTabChange: (tab: TabType) => void;
}

const GroupManageTabPanel: React.FC<GroupManageTabPanelProps> = ({
  activeTab,
  onTabChange,
}) => {
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
        <View style={[styles.innerButton, activeTab === tab]}>
          <Image 
            source={icon} 
            style={[
              styles.tabIcon,
              activeTab === tab && styles.tabIconActive
            ]} 
          />
        </View>
      </LinearGradient>
    </TouchableOpacity>
  );

  return (
    <View style={styles.tabsContainer}>
      {renderTabButton('info', require('../assets/images/groups/panel/group.png'))}
      {renderTabButton('requests', require('../assets/images/groups/panel/person_add.png'))}
      {renderTabButton('events', require('../assets/images/groups/panel/event.png'))}
    </View>
  );
};

const styles = StyleSheet.create({
  tabsContainer: {
    flexDirection: 'row',
    marginHorizontal: 45,
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
  innerButton: {
    flex: 1,
    backgroundColor: Colors.veryLightGrey,
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
    tintColor: Colors.lightGreyBlue,
  },
  tabIconActive: {
    tintColor: Colors.darkGrey,
  },
});

export default GroupManageTabPanel;