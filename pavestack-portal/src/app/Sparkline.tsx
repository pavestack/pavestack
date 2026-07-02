import React from "react";

/** Minimal hand-rolled sparkline — no charting library needed for illustrative golden-signal snapshots. */
export function Sparkline({
  values,
  width = 160,
  height = 36,
  color = "var(--accent)",
  markers,
}: {
  values: number[];
  width?: number;
  height?: number;
  color?: string;
  /** Indices to highlight (e.g. deploy events) with a vertical tick. */
  markers?: number[];
}) {
  if (values.length === 0) return null;
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;
  const stepX = width / Math.max(values.length - 1, 1);

  const points = values
    .map((v, i) => {
      const x = i * stepX;
      const y = height - ((v - min) / range) * (height - 4) - 2;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      className="overflow-visible"
      role="img"
      aria-label="Sample trend sparkline"
    >
      <polyline
        points={points}
        fill="none"
        stroke={color}
        strokeWidth={1.75}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {markers?.map((i) => {
        const x = i * stepX;
        return (
          <line
            key={i}
            x1={x}
            y1={0}
            x2={x}
            y2={height}
            stroke="var(--danger)"
            strokeWidth={1}
            strokeDasharray="2,2"
            opacity={0.6}
          />
        );
      })}
    </svg>
  );
}
