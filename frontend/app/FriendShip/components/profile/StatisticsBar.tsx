import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { Image, StyleSheet, Text, View } from 'react-native';

interface StatisticsBarProps {
  title: string;
  count?: number | string;
  icon?: string;
  fullWidth?: boolean;
}

const StatisticsBar = ({ title, count = 20, icon, fullWidth = false }: StatisticsBarProps) => {
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
    <View style={[styles.container, fullWidth ? styles.containerFullWidth : styles.containerFitContent]}>
      {icon && (
        <Image 
          source={getIconSource(icon)} 
          style={styles.icon}
        />
      )}
      <Text style={styles.text} numberOfLines={1}>{title}</Text>
      <Text style={styles.separator}>-</Text>
      <Text style={styles.countText} numberOfLines={1}>{count}</Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.lightBlue2,
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
    color: Colors.black,
    fontSize: 14,
    fontFamily: Montserrat.regular,
    flexShrink: 1,
  },
  separator: {
    color: Colors.black,
    fontSize: 14,
    fontFamily: Montserrat.regular,
    marginHorizontal: 6,
    flexShrink: 0,
  },
  countText: {
    color: Colors.black,
    fontSize: 14,
    fontFamily: Montserrat.regular,
    flexShrink: 1,
  },
});

export default StatisticsBar;