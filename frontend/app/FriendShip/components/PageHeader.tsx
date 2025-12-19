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
  showBackButton?: boolean;
  onBackPress?: () => void;
}

const PageHeader: React.FC<PageHeaderProps> = ({ 
  title, 
  showWave = false, 
  actionButton,
  showBackButton = false,
  onBackPress
}) => {
  const { isDark } = useTheme();
  const colors = getColors(isDark);

  const getFontSize = () => {
    const length = title.length;
    if (length <= 15) return 24;
    if (length <= 25) return 20;
    if (length <= 35) return 18;
    return 16;
  };

  const getTextContainerStyle = () => {
    const baseStyle = { flex: 1, alignItems: 'center' as const };
    
    if (showBackButton && actionButton) {
      return { ...baseStyle, paddingHorizontal: 46 };
    } else if (showBackButton || actionButton) {
      return { 
        ...baseStyle, 
        paddingLeft: showBackButton ? 46 : 0,
        paddingRight: actionButton ? 46 : 0
      };
    }
    
    return baseStyle;
  };

  return (
    <>
      <View style={[styles.headerContainer, { backgroundColor: colors.lightBlue2 }]}>
        {showBackButton && onBackPress && (
          <TouchableOpacity 
            style={styles.backButton}
            onPress={onBackPress}
            activeOpacity={0.7}
          >
            <Image 
              source={require('@/assets/images/arrowLeft.png')} 
              style={styles.backIcon}
            />
          </TouchableOpacity>
        )}
        
        <View style={getTextContainerStyle()}>
          <Text 
            style={[styles.headerTitle, { fontSize: getFontSize() }]}
            numberOfLines={2}
            adjustsFontSizeToFit
            minimumFontScale={0.7}
          >
            {title}
          </Text>
        </View>
        
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
    flexDirection: 'row',
  },
  headerTitle: {
    fontFamily: Montserrat_Alternates.medium,
    textTransform: 'none',
    color: Colors.white,
    textAlign: 'center',
  },
  backButton: {
    position: 'absolute',
    left: 16,
    padding: 8,
    zIndex: 10,
  },
  backIcon: {
    width: 30,
    height: 30,
    tintColor: Colors.white
  },
  actionButton: {
    position: 'absolute',
    right: 16,
    padding: 8,
    zIndex: 10,
  },
  actionIcon: {
    width: 30,
    height: 30,
    tintColor: Colors.white
  },
});

export default PageHeader;