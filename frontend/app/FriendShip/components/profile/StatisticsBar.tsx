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
    <View style={[styles.container, fullWidth && styles.containerFullWidth]}>
      {icon && (
        <Image 
          source={getIconSource(icon)} 
          style={styles.icon}
        />
      )}
      <Text style={styles.text}>{title}</Text>
      <Text style={styles.separator}>-</Text>
      <Text style={styles.text}>{count}</Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: Colors.lightBlue2,
    flexDirection: 'row',    
    borderRadius: 40,
    paddingVertical: 10,
    paddingHorizontal: 16,
    marginRight: 8,
    alignItems: 'center',
  },
  containerFullWidth: {
    marginRight: 0,
  },
  icon: {
    width: 30,
    height: 30,
    resizeMode: 'contain',
    marginRight: 6
  },
  text: {
    color: Colors.black,
    fontSize: 16,
    fontFamily: Montserrat.regular,
  },
  separator: {
    color: Colors.black,
    fontSize: 16,
    fontFamily: Montserrat.regular,
    marginHorizontal: 6,
  },
});

export default StatisticsBar;