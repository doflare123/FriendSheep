import StatisticsBar from '@/components/profile/StatisticsBar';
import React from 'react';
import { StyleSheet, View } from 'react-native';

interface StatisticsBottomBarsProps {
  mostPopularDay: string;
  sessionsCreated: number;
  sessionsSeries: number;
}

const StatisticsBottomBars: React.FC<StatisticsBottomBarsProps> = ({
  mostPopularDay,
  sessionsCreated,
  sessionsSeries,
}) => {
  return (
    <View style={styles.container}>
      <StatisticsBar 
        title='Популярный день' 
        count={mostPopularDay} 
        icon="most-popular-day" 
      />
      <StatisticsBar 
        title='Создано сессий' 
        count={sessionsCreated} 
        icon="sessions-created" 
      />
      <StatisticsBar 
        title='Серия сессий' 
        count={sessionsSeries} 
        icon="series-sessions" 
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    gap: 4,
    paddingHorizontal: 16,
    marginTop: 8,
  },
});

export default StatisticsBottomBars;