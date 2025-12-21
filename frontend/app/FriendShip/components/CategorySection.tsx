import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { ReactNode } from 'react';
import { Image, ImageSourcePropType, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { Event } from './event/EventCard';
import EventCarousel from './event/EventCarousel';
import { useTheme } from './ThemeContext';

interface CategorySectionProps {
  title: string;
  events?: Event[];
  children?: ReactNode;
  showArrow?: boolean;
  onArrowPress?: () => void;
  centerTitle?: boolean;
  showLineVariant?: 'line' | 'line2';
  marginBottom?: number;
  customActionButton?: {
    icon: ImageSourcePropType;
    onPress: () => void;
    tintColor?: string;
  };
  secondaryActionButton?: {
    icon: ImageSourcePropType;
    onPress: () => void;
    tintColor?: string;
  };
  showBackButton?: boolean;
  onBackPress?: () => void;
  customNavigationButtons?: {
    leftButton: {
      icon: ImageSourcePropType;
      onPress: () => void;
    };
    rightButton: {
      icon: ImageSourcePropType;
      onPress: () => void;
    };
  };
}

const CategorySection: React.FC<CategorySectionProps> = ({
  title,
  events,
  children,
  showArrow = false,
  onArrowPress,
  centerTitle = false,
  showLineVariant = 'line',
  marginBottom,
  customActionButton,
  secondaryActionButton,
  showBackButton = false,
  onBackPress,
  customNavigationButtons,
}) => {
  const colors = useThemedColors();
  const { isDark } = useTheme();

const lineImage = showLineVariant === 'line2' 
  ? (isDark 
      ? require('../assets/images/line2_2.png')
      : require('../assets/images/line2.png'))
  : (isDark
      ? require('../assets/images/line_2.png')
      : require('../assets/images/line.png')); 

  if (customNavigationButtons) {
    return (
      <>
        <View style={[styles.resultsHeader, marginBottom ? { marginBottom } : null]}>
          <View style={styles.navigationContainer}>
            <TouchableOpacity onPress={customNavigationButtons.leftButton.onPress}>
              <Image
                source={customNavigationButtons.leftButton.icon}
                style={[styles.navigationIcon, { tintColor: colors.blue2 }]}
              />
            </TouchableOpacity>
            
            <View style={styles.titleWithLineContainer}>
              <Text style={[styles.resultsText, centerTitle && { textAlign: 'center' }, { color: colors.blue2 }]}>
                {title}
              </Text>
            </View>
            
            <TouchableOpacity onPress={customNavigationButtons.rightButton.onPress}>
              <Image
                source={customNavigationButtons.rightButton.icon}
                style={[styles.navigationIcon, { tintColor: colors.blue2 }]}
              />
            </TouchableOpacity>
          </View>
          <Image
              source={lineImage}
              style={styles.lineImage}
          />
        </View>
        {events && events.length > 0 && <EventCarousel events={events} />}
        {children}
      </>
    );
  }

  return (
    <>
      <View style={[styles.resultsHeader, marginBottom ? { marginBottom } : null]}>
        <View style={{
          flexDirection: 'row',
          justifyContent: showArrow || customActionButton || secondaryActionButton ? 'space-between' : centerTitle ? 'center' : 'flex-start',
          alignItems: 'center'
        }}>
          <View style={{ flexDirection: 'row', alignItems: 'center', flex: centerTitle ? 0 : 1 }}>
            {showBackButton && onBackPress && (
              <TouchableOpacity onPress={onBackPress} style={styles.backButton}>
                <Image
                  source={require('../assets/images/arrowLeft.png')}
                  style={[styles.navigationIcon, { tintColor: colors.blue2 }]}
                />
              </TouchableOpacity>
            )}
            <Text style={[styles.resultsText, centerTitle && { textAlign: 'center' }, { color: colors.blue2 }]}>
              {title}
            </Text>
          </View>

          {(showArrow || customActionButton || secondaryActionButton) && (
            <View style={{ flexDirection: 'row', gap: 8 }}>
              {customActionButton && (
                <TouchableOpacity onPress={customActionButton.onPress}>
                  <Image
                    source={customActionButton.icon}
                    style={{ resizeMode: 'contain', width: 30, height: 30 }}
                    tintColor={customActionButton.tintColor || colors.blue2}
                  />
                </TouchableOpacity>
              )}
              
              {secondaryActionButton && (
                <TouchableOpacity onPress={secondaryActionButton.onPress}>
                  <Image
                    source={secondaryActionButton.icon}
                    style={{ resizeMode: 'contain', width: 30, height: 30 }}
                    tintColor={secondaryActionButton.tintColor || colors.blue2}
                  />
                </TouchableOpacity>
              )}
              
              {showArrow && onArrowPress && (
                <TouchableOpacity onPress={onArrowPress}>
                  <Image
                    source={require('../assets/images/more.png')}
                    style={[styles.navigationIcon, { tintColor: colors.blue2 }]}
                  />
                </TouchableOpacity>
              )}
            </View>
          )}
        </View>
        <Image
          source={lineImage}
          style={{ resizeMode: 'none' }}
        />
      </View>
      {events && events.length > 0 && <EventCarousel events={events} />}
      {children}
    </>
  );
};

const styles = StyleSheet.create({
  resultsHeader: {
    paddingHorizontal: 16,
  },
  resultsText: {
    fontFamily: Montserrat_Alternates.medium,
    fontSize: 20,
    marginBottom: 4,
  },
  backButton: {
    marginRight: 12,
  },
  navigationContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  navigationIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
  },
  titleWithLineContainer: {
    flex: 1,
    marginHorizontal: 12,
  },
  lineImage: {
    width: '100%',
    alignSelf: 'center',
    resizeMode: 'none',
    marginTop: 4,
  },
});

export default CategorySection;