'use client';

import React from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';
import section2Styles from '../../styles/profile/ProfileSection2.module.css';

interface ChartEntry {
  name: string;
  value: number;
  color?: string;
}

interface Props {
  data: ChartEntry[];
  animated: boolean;
}

/* Небольшие вспомогательные лейблы (стабильные — определены вне родителя) */
const RADIAN = Math.PI / 180;

const AnimatedNumberLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, index }: any) => {
  const radius = innerRadius + (outerRadius - innerRadius) * 0.4;
  const x = cx + radius * Math.cos(-midAngle * RADIAN);
  const y = cy + radius * Math.sin(-midAngle * RADIAN);

  return (
    <text
      x={x}
      y={y}
      fill="white"
      stroke="black"
      strokeWidth={2}
      paintOrder="stroke fill"
      textAnchor="middle"
      dominantBaseline="central"
      fontSize="18"
      fontWeight="bold"
      className={section2Styles.chartLabelAnimated} /* класс контролирует анимацию */
      style={{ animationDelay: `${index * 0.2}s` }}
    >
      {index + 1}
    </text>
  );
};

const AnimatedPercentLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, percent, index }: any) => {
  const radius = innerRadius + (outerRadius - innerRadius) * 0.8;
  const x = cx + radius * Math.cos(-midAngle * RADIAN);
  const y = cy + radius * Math.sin(-midAngle * RADIAN);

  return (
    <text
      x={x}
      y={y}
      fill="white"
      stroke="black"
      strokeWidth={2}
      paintOrder="stroke fill"
      textAnchor="middle"
      dominantBaseline="central"
      fontSize="14"
      fontWeight="bold"
      className={section2Styles.chartLabelAnimated}
      style={{ animationDelay: `${0.4 + index * 0.2}s` }}
    >
      {`${(percent * 100).toFixed(0)}%`}
    </text>
  );
};

/* Кастомный тултип (копия вашей реализации, чтобы компонент был самодостаточен) */
const CustomTooltip = ({ active, payload }: any) => {
  if (active && payload && payload.length) {
    const data = payload[0];
    return (
      <div style={{
        background: 'white',
        border: '2px solid #37A2E6',
        borderRadius: '8px',
        padding: '8px 12px',
        fontSize: '14px',
        color: '#000',
        pointerEvents: 'none',
        zIndex: 1000
      }}>
        <p style={{ margin: 0, fontWeight: 'bold' }}>{data.name}</p>
        <p style={{ margin: 0, color: '#666' }}>Количество: {data.value}</p>
      </div>
    );
  }
  return null;
};

function GenrePieChart({ data, animated }: Props) {
  return (
    <ResponsiveContainer width="60%" height={320}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          outerRadius={130}
          dataKey="value"
          labelLine={false}
          label={AnimatedNumberLabel}
        >
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.color} />
          ))}
        </Pie>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          outerRadius={130}
          dataKey="value"
          labelLine={false}
          label={AnimatedPercentLabel}
        >
          {data.map((entry, index) => (
            <Cell key={`cell-p-${index}`} fill="transparent" />
          ))}
        </Pie>
        <Tooltip content={<CustomTooltip />} />
      </PieChart>
    </ResponsiveContainer>
  );
}

/* memo: компонент перерисуется только если изменятся data или animated (сравнение по === для ссылок) */
export default React.memo(GenrePieChart, (prev, next) => prev.data === next.data && prev.animated === next.animated);
