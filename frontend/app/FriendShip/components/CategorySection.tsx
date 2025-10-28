import { Colors } from '@/constants/Colors';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import React, { ReactNode } from 'react';
import { Image, ImageSourcePropType, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { Event } from './event/EventCard';
import EventCarousel from './event/EventCarousel';

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
  showBackButton = false,
  onBackPress,
  customNavigationButtons,
}) => {
  const lineImage = showLineVariant === 'line2' 
    ? require('../assets/images/line2.png')
    : require('../assets/images/line.png');

  if (customNavigationButtons) {
    return (
      <>
        <View style={[styles.resultsHeader, marginBottom ? { marginBottom } : null]}>
          <View style={styles.navigationContainer}>
            <TouchableOpacity onPress={customNavigationButtons.leftButton.onPress}>
              <Image
                source={customNavigationButtons.leftButton.icon}
                style={styles.navigationIcon}
              />
            </TouchableOpacity>
            
            <View style={styles.titleWithLineContainer}>
              <Text style={[styles.resultsText, centerTitle && { textAlign: 'center' }]}>
                {title}
              </Text>
            </View>
            
            <TouchableOpacity onPress={customNavigationButtons.rightButton.onPress}>
              <Image
                source={customNavigationButtons.rightButton.icon}
                style={styles.navigationIcon}
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
          justifyContent: showArrow || customActionButton ? 'space-between' : centerTitle ? 'center' : 'flex-start',
          alignItems: 'center'
        }}>
          {showBackButton && onBackPress && (
            <TouchableOpacity onPress={onBackPress} style={styles.backButton}>
              <Image
                source={require('../assets/images/arrowLeft.png')}
                style={{ resizeMode: 'contain', width: 30, height: 30 }}
              />
            </TouchableOpacity>
          )}
          <Text style={[styles.resultsText, centerTitle && { textAlign: 'center' }]}>
            {title}
          </Text>
          {showArrow && onArrowPress && (
            <TouchableOpacity onPress={onArrowPress}>
              <Image
                source={require('../assets/images/arrow.png')}
                style={{ resizeMode: 'contain', width: 30, height: 30 }}
              />
            </TouchableOpacity>
          )}
          {customActionButton && (
            <TouchableOpacity onPress={customActionButton.onPress}>
              <Image
                source={customActionButton.icon}
                style={{ resizeMode: 'contain', width: 30, height: 30 }}
                tintColor={customActionButton.tintColor}
              />
            </TouchableOpacity>
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
    color: Colors.blue2,
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