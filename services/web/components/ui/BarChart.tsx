/**
 * BarChart — lightweight SVG bar chart, no external deps.
 * Designed for the Neon Shrine palette.
 */

interface BarChartProps {
  data: { label: string; value: number }[];
  height?: number;
  barColor?: string;
  labelInterval?: number; // show x-axis label every N bars (default: auto)
}

export function BarChart({
  data,
  height = 120,
  barColor = "var(--color-primary)",
  labelInterval,
}: BarChartProps) {
  if (data.length === 0) {
    return (
      <div
        className="flex items-center justify-center rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] text-sm text-[var(--color-text-muted)]"
        style={{ height }}
      >
        No data
      </div>
    );
  }

  const max = Math.max(...data.map((d) => d.value), 1);
  const paddingTop = 8;
  const paddingBottom = 24; // room for labels
  const chartH = height - paddingTop - paddingBottom;
  const svgH = height;

  // Auto label interval: aim for ≤ 8 visible labels
  const auto = Math.ceil(data.length / 8);
  const every = labelInterval ?? (data.length <= 8 ? 1 : auto);

  return (
    <svg
      viewBox={`0 0 ${data.length * 12} ${svgH}`}
      width="100%"
      height={svgH}
      preserveAspectRatio="none"
      aria-label="Play history bar chart"
    >
      {data.map((d, i) => {
        const barH = Math.max(2, (d.value / max) * chartH);
        const x = i * 12 + 1;
        const y = paddingTop + (chartH - barH);

        return (
          <g key={i}>
            <rect
              x={x}
              y={y}
              width={10}
              height={barH}
              rx={2}
              fill={barColor}
              opacity={d.value === 0 ? 0.15 : 0.85}
            />
            {i % every === 0 && (
              <text
                x={x + 5}
                y={svgH - 4}
                textAnchor="middle"
                fontSize={7}
                fill="var(--color-text-muted)"
              >
                {d.label}
              </text>
            )}
          </g>
        );
      })}
    </svg>
  );
}
