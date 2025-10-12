import StatisticsBar from '@/components/profile/StatisticsBar';
import React from 'react';
import { StyleSheet, View } from 'react-native';

interface ProfileStatsProps {
  media: number;
  games: number;
  hours: number;
  sessions: number;
}

const ProfileStats: React.FC<ProfileStatsProps> = ({
  media,
  games,
  hours,
  sessions,
}) => {
  return (
    <>
      <View style={styles.statisticsRow}>
        <StatisticsBar title='Медиа' count={media} icon="movies" fullWidth />
        <StatisticsBar title='Игры' count={games} icon="games" fullWidth />
      </View>
      
      <View style={[styles.statisticsRow, styles.lastRow]}>
        <StatisticsBar title='Часов' count={hours} icon="hours" fullWidth />
        <StatisticsBar title='Сессий' count={sessions} icon="sessions" fullWidth />
      </View>
    </>
  );
};

const styles = StyleSheet.create({
  statisticsRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 8,
    paddingHorizontal: 16,
    marginVertical: 2,
  },
  lastRow: {
    marginBottom: 14,
  },
});

export default ProfileStats;