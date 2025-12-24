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

  return (
    <>
      <View style={[styles.headerContainer, { backgroundColor: colors.lightBlue2 }]}>
        <View style={styles.sideButton}>
          {showBackButton && onBackPress ? (
            <TouchableOpacity 
              onPress={onBackPress}
              activeOpacity={0.7}
              style={styles.buttonTouchable}
            >
              <Image 
                source={require('@/assets/images/arrowLeft.png')} 
                style={styles.icon}
              />
            </TouchableOpacity>
          ) : (
            <View style={styles.buttonTouchable} />
          )}
        </View>
        
        <View style={styles.textContainer}>
          <Text 
            style={[styles.headerTitle, { fontSize: getFontSize() }]}
            numberOfLines={2}
            adjustsFontSizeToFit
            minimumFontScale={0.7}
          >
            {title}
          </Text>
        </View>
        
        <View style={styles.sideButton}>
          {actionButton ? (
            <TouchableOpacity 
              onPress={actionButton.onPress}
              activeOpacity={0.7}
              style={styles.buttonTouchable}
            >
              <Image 
                source={actionButton.icon} 
                style={styles.icon}
              />
            </TouchableOpacity>
          ) : (
            <View style={styles.buttonTouchable} />
          )}
        </View>
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
    justifyContent: 'space-between',
    marginBottom: 12,
    flexDirection: 'row',
  },
  sideButton: {
    width: 46,
    alignItems: 'center',
    justifyContent: 'center',
  },
  buttonTouchable: {
    width: 46,
    height: 46,
    alignItems: 'center',
    justifyContent: 'center',
  },
  textContainer: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 8,
  },
  headerTitle: {
    fontFamily: Montserrat_Alternates.medium,
    textTransform: 'none',
    color: Colors.white,
    textAlign: 'center',
  },
  icon: {
    width: 30,
    height: 30,
    tintColor: Colors.white
  },
});

export default PageHeader;