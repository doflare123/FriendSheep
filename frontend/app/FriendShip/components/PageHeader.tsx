import { Colors } from '@/constants/Colors';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import React from 'react';
import { Image, ImageSourcePropType, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

interface PageHeaderProps {
  title: string;
  showWave?: boolean;
  actionButton?: {
    icon: ImageSourcePropType;
    onPress: () => void;
  };
}

const PageHeader: React.FC<PageHeaderProps> = ({ title, showWave = false, actionButton }) => {
  return (
    <>
      <View style={styles.headerContainer}>
        <Text style={styles.headerTitle}>{title}</Text>
        {actionButton && (
          <TouchableOpacity 
            style={styles.actionButton}
            onPress={actionButton.onPress}
            activeOpacity={0.7}
          >
            <Image 
              source={actionButton.icon} 
              style={styles.actionIcon}
            />
          </TouchableOpacity>
        )}
      </View>
      {showWave && (
        <Image
          source={require('@/assets/images/wave.png')}
          style={{ resizeMode: 'none', marginBottom: -20 }}
        />
      )}
    </>
  );
};

const styles = StyleSheet.create({
  headerContainer: {
    backgroundColor: Colors.lightBlue2,
    paddingVertical: 20,
    paddingHorizontal: 16,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 12,
    position: 'relative',
  },
  headerTitle: {
    fontFamily: Montserrat_Alternates.medium,
    fontSize: 24,
    color: Colors.white,
    textTransform: 'none',
  },
  actionButton: {
    position: 'absolute',
    right: 16,
    padding: 8,
  },
  actionIcon: {
    width: 30,
    height: 30,
    tintColor: Colors.white,
  },
});

export default PageHeader;