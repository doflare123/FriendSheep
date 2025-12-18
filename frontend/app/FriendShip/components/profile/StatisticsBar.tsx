import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React from 'react';
import { Image, StyleSheet, Text, View } from 'react-native';

interface StatisticsBarProps {
  title: string;
  count?: number | string;
  icon?: string;
  fullWidth?: boolean;
}

const StatisticsBar = ({ title, count = 20, icon, fullWidth = false }: StatisticsBarProps) => {
  const colors = useThemedColors();

  const getIconSource = (iconType?: string) => {
    switch (iconType) {
      case 'movies':
        return require('@/assets/images/profile/movies.png');
      case 'games':
        return require('@/assets/images/profile/games.png');
      case 'table_games':
        return require('@/assets/images/profile/table_games.png');
      case 'other':
        return require('@/assets/images/profile/others.png');
      case 'hours':
        return require('@/assets/images/profile/hours.png');
      case 'sessions':
        return require('@/assets/images/profile/sessions.png');
      case 'most-popular-day':
        return require('@/assets/images/profile/most-popular-day.png');
      case 'sessions-created':
        return require('@/assets/images/profile/sessions-created.png');
      case 'series-sessions':
        return require('@/assets/images/profile/series-sessions.png');
      default:
        return require('@/assets/images/profile/movies.png');
    }
  };

  return (
    <View style={[
      styles.container,
      { backgroundColor: colors.lightBlue2 },
      fullWidth ? styles.containerFullWidth : styles.containerFitContent
    ]}>
      {icon && (
        <Image 
          source={getIconSource(icon)} 
          style={[styles.icon, {tintColor: Colors.black}]}
        />
      )}
      <Text style={[styles.text, { color: Colors.black }]} numberOfLines={1}>
        {title}
      </Text>
      <Text style={[styles.separator, { color: Colors.black }]}>-</Text>
      <Text style={[styles.countText, { color: Colors.black }]} numberOfLines={1}>
        {count}
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',    
    borderRadius: 40,
    paddingVertical: 8,
    paddingHorizontal: 12,
    marginRight: 8,
    marginVertical: 2,
    alignItems: 'center',
  },
  containerFullWidth: {
    marginRight: 0,
    flex: 1,
  },
  containerFitContent: {
    alignSelf: 'flex-start',
  },
  icon: {
    width: 26,
    height: 26,
    resizeMode: 'contain',
    marginRight: 6,
    flexShrink: 0,
  },
  text: {
    fontSize: 14,
    fontFamily: Montserrat.regular,
    flexShrink: 1,
  },
  separator: {
    fontSize: 14,
    fontFamily: Montserrat.regular,
    marginHorizontal: 6,
    flexShrink: 0,
  },
  countText: {
    fontSize: 14,
    fontFamily: Montserrat.regular,
    flexShrink: 1,
  },
});

export default StatisticsBar;