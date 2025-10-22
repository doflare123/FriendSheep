import CategorySection from '@/components/CategorySection';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { Dimensions, StyleSheet, View } from 'react-native';
import { PieChart } from 'react-native-chart-kit';

export interface StatisticsDataItem {
  name: string;
  percentage: number;
  color: string;
  legendFontColor: string;
}

interface StatisticsChartProps {
  statisticsData: StatisticsDataItem[];
  mostPopularDay?: string;
  sessionsCreated?: number;
  sessionsSeries?: number;
}

const StatisticsChart: React.FC<StatisticsChartProps> = ({
  statisticsData,
  mostPopularDay = 'Воскресенье',
  sessionsCreated = 4,
  sessionsSeries = 4,
}) => {
  const chartConfig = {
    color: (opacity = 1) => `rgba(0, 0, 0, ${opacity})`,
  };

  return (
    <CategorySection title="Статистика:" marginBottom={16}>
      <View style={styles.statisticsContainer}>
        <View style={styles.chartContainer}>
          <PieChart
            data={statisticsData.map(item => ({
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

        </View>
      </View>
    </CategorySection>
  );
};

const styles = StyleSheet.create({
  statisticsContainer: {
    alignItems: 'center',
    marginTop: -16
  },
  chartContainer: {
    justifyContent: 'space-between'
  },
  statisticsBottomBar: {
    gap: 4,
  },
});

export default StatisticsChart;