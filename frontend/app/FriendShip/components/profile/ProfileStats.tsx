import StatisticsBar from '@/components/profile/StatisticsBar';
import { TileType } from '@/components/profile/TileSelectionModal';
import React from 'react';
import { StyleSheet, View } from 'react-native';

interface ProfileStatsProps {
  selectedTiles: TileType[];
  stats: {
    media: number;
    games: number;
    table_games: number;
    other: number;
    hours: number;
    sessions: number;
  };
}

interface TileConfig {
  type: TileType;
  title: string;
  icon: string;
}

const tileConfigs: Record<TileType, TileConfig> = {
  media: { type: 'media', title: 'Медиа', icon: 'movies' },
  games: { type: 'games', title: 'Игры', icon: 'games' },
  table_games: { type: 'table_games', title: 'Настолки', icon: 'table_games' },
  other: { type: 'other', title: 'Другое', icon: 'other' },
  hours: { type: 'hours', title: 'Часов', icon: 'hours' },
  sessions: { type: 'sessions', title: 'Сессий', icon: 'sessions' },
};

const ProfileStats: React.FC<ProfileStatsProps> = ({ selectedTiles, stats }) => {
  const displayedTiles = selectedTiles.slice(0, 4);
  
  const renderTile = (tileType: TileType) => {
    const config = tileConfigs[tileType];
    const count = stats[tileType];
    
    return (
      <StatisticsBar
        key={tileType}
        title={config.title}
        count={count}
        icon={config.icon}
        fullWidth
      />
    );
  };

  const firstRow = displayedTiles.slice(0, 2);
  const secondRow = displayedTiles.slice(2, 4);

  return (
    <>
      {firstRow.length > 0 && (
        <View style={styles.statisticsRow}>
          {firstRow.map(renderTile)}
        </View>
      )}
      
      {secondRow.length > 0 && (
        <View style={[styles.statisticsRow, styles.lastRow]}>
          {secondRow.map(renderTile)}
        </View>
      )}
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