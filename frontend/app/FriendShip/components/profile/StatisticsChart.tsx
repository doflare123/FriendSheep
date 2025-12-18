import CategorySection from '@/components/CategorySection';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useRef, useState } from 'react';
import { Dimensions, ScrollView, StyleSheet, View } from 'react-native';
import { PieChart } from 'react-native-chart-kit';

export interface StatisticsDataItem {
  name: string;
  percentage: number;
  color: string;
  legendFontColor: string;
}

interface StatisticsChartProps {
  genresData: StatisticsDataItem[];
  categoriesData: StatisticsDataItem[];
}

const StatisticsChart: React.FC<StatisticsChartProps> = ({
  genresData,
  categoriesData,
}) => {
  const colors = useThemedColors();
  const [currentPage, setCurrentPage] = useState(0);
  const scrollViewRef = useRef<ScrollView>(null);
  const screenWidth = Dimensions.get('window').width;

  const chartConfig = {
    color: (opacity = 1) => `rgba(0, 0, 0, ${opacity})`,
  };

  const handleScroll = (event: any) => {
    const offsetX = event.nativeEvent.contentOffset.x;
    const page = Math.round(offsetX / screenWidth);
    setCurrentPage(page);
  };

  const renderChart = (data: StatisticsDataItem[]) => (
    <View style={[styles.chartPage, { width: screenWidth }]}>
      <View style={styles.chartContainer}>
        <PieChart
          data={data.map(item => ({
            name: item.name,
            population: item.percentage,
            color: item.color,
            legendFontColor: item.legendFontColor,
            legendFontSize: 9,
            legendFontFamily: Montserrat.regular,
          }))}
          width={screenWidth - 60}
          height={200}
          chartConfig={chartConfig}
          accessor="population"
          backgroundColor="transparent"
          paddingLeft="4"
          absolute={false}
          hasLegend={true}
          style={{
            marginBottom: 8,
          }}
        />
      </View>
    </View>
  );

  return (
    <CategorySection title="Статистика:" marginBottom={16}>
      <View style={styles.statisticsContainer}>
        <ScrollView
          ref={scrollViewRef}
          horizontal
          pagingEnabled
          showsHorizontalScrollIndicator={false}
          onScroll={handleScroll}
          scrollEventThrottle={16}
        >
          {renderChart(genresData)}
          {renderChart(categoriesData)}
        </ScrollView>

        <View style={styles.paginationContainer}>
          {[0, 1].map((index) => (
            <View
              key={index}
              style={[
                styles.paginationDot,
                { backgroundColor: colors.lightGrey },
                currentPage === index && [
                  styles.paginationDotActive,
                  { backgroundColor: colors.lightBlue }
                ],
              ]}
            />
          ))}
        </View>
      </View>
    </CategorySection>
  );
};

const styles = StyleSheet.create({
  statisticsContainer: {
    alignItems: 'center',
    marginTop: -16,
  },
  chartPage: {
    alignItems: 'center',
  },
  chartContainer: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  paginationContainer: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    marginTop: 8,
    gap: 8,
  },
  paginationDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  paginationDotActive: {
    width: 24,
  },
});

export default StatisticsChart;