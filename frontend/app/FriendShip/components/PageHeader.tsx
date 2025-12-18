import { useTheme } from '@/components/ThemeContext';
import { Colors, getColors } from '@/constants/Colors';
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
  const { isDark } = useTheme();
  const colors = getColors(isDark);

  return (
    <>
      <View style={[styles.headerContainer, { backgroundColor: colors.lightBlue2 }]}>
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
    textTransform: 'none',
    color: Colors.white
  },
  actionButton: {
    position: 'absolute',
    right: 16,
    padding: 8,
  },
  actionIcon: {
    width: 30,
    height: 30,
    tintColor: Colors.white
  },
});

export default PageHeader;