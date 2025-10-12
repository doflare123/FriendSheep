import CategorySection from '@/components/CategorySection';
import StatisticsBar from '@/components/profile/StatisticsBar';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { Dimensions, StyleSheet, View } from 'react-native';
import { PieChart } from 'react-native-chart-kit';

export type StatisticsType = 'media' | 'games' | 'boardgames' | 'other';

export interface StatisticsDataItem {
  name: string;
  percentage: number;
  color: string;
  legendFontColor: string;
}

export interface StatisticsData {
  media: StatisticsDataItem[];
  games: StatisticsDataItem[];
  boardgames: StatisticsDataItem[];
  other: StatisticsDataItem[];
}

interface StatisticsChartProps {
  statisticsData: StatisticsData;
  currentType: StatisticsType;
  onTypeChange: (type: StatisticsType) => void;
  mostPopularDay?: string;
  sessionsCreated?: number;
  sessionsSeries?: number;
}

const StatisticsChart: React.FC<StatisticsChartProps> = ({
  statisticsData,
  currentType,
  onTypeChange,
  mostPopularDay = 'Воскресенье',
  sessionsCreated = 4,
  sessionsSeries = 4,
}) => {
  const getStatisticsForType = (type: StatisticsType) => {
    return statisticsData[type] || statisticsData.media;
  };

  const getStatisticsTitle = (type: StatisticsType) => {
    const titles: Record<StatisticsType, string> = {
      media: 'Медиа',
      games: 'Игры',
      boardgames: 'Настольные игры',
      other: 'Другое',
    };
    return titles[type];
  };

  const handlePrevType = () => {
    const types: StatisticsType[] = ['media', 'games', 'boardgames', 'other'];
    const currentIndex = types.indexOf(currentType);
    const prevIndex = currentIndex > 0 ? currentIndex - 1 : types.length - 1;
    onTypeChange(types[prevIndex]);
  };

  const handleNextType = () => {
    const types: StatisticsType[] = ['media', 'games', 'boardgames', 'other'];
    const currentIndex = types.indexOf(currentType);
    const nextIndex = currentIndex < types.length - 1 ? currentIndex + 1 : 0;
    onTypeChange(types[nextIndex]);
  };

  const chartConfig = {
    color: (opacity = 1) => `rgba(0, 0, 0, ${opacity})`,
  };

  return (
    <CategorySection title="Статистика:" marginBottom={16}>
      <CategorySection 
        title={getStatisticsTitle(currentType)}
        centerTitle
        showLineVariant="line2"
        marginBottom={8}
        customNavigationButtons={{
          leftButton: {
            icon: require('@/assets/images/arrowLeft.png'),
            onPress: handlePrevType,
          },
          rightButton: {
            icon: require('@/assets/images/arrow.png'),
            onPress: handleNextType,
          },
        }}
      />

      <View style={styles.statisticsContainer}>
        <View style={styles.chartContainer}>
          <PieChart
            data={getStatisticsForType(currentType).map(item => ({
              name: item.name,
              population: item.percentage,
              color: item.color,
              legendFontColor: item.legendFontColor,
              legendFontSize: 12,
              legendFontFamily: Montserrat.regular,
            }))}
            width={Dimensions.get('window').width - 60}
            height={200}
            chartConfig={chartConfig}
            accessor="population"
            backgroundColor="transparent"
            paddingLeft="0"
            absolute={false}
            hasLegend={true}
          />
        </View>
      
        <View style={styles.statisticsBottomBar}>
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
      </View>
    </CategorySection>
  );
};

const styles = StyleSheet.create({
  statisticsContainer: {
    paddingHorizontal: 16,
  },
  chartContainer: {
    justifyContent: 'space-between'
  },
  statisticsBottomBar: {
    gap: 4,
  },
});

export default StatisticsChart;