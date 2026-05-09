import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate } from "remotion";
import { springEnter } from "../lib/animations";
import { COLORS, FONTS, RADIUS } from "../lib/theme";

const CLAMP = { extrapolateLeft: "clamp", extrapolateRight: "clamp" } as const;

const TRAFFIC_LIGHTS = ["#ff5f57", "#ffbd2e", "#28ca41"];

interface CodeWindowProps {
  lines: string[];
  title?: string;
  /** Frames between each line appearing. Default: 8 */
  framesPerLine?: number;
  style?: React.CSSProperties;
}

export const CodeWindow: React.FC<CodeWindowProps> = ({
  lines,
  title = "main.ts",
  framesPerLine = 8,
  style,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const windowScale = springEnter(frame, fps, { damping: 14 });
  const windowOpacity = interpolate(frame, [0, 10], [0, 1], CLAMP);
  const visibleCount = Math.min(lines.length, Math.floor(frame / framesPerLine) + 1);

  return (
    <div
      style={{
        transform: `scale(${windowScale})`,
        opacity: windowOpacity,
        transformOrigin: "top center",
        backgroundColor: COLORS.card,
        borderRadius: RADIUS.lg,
        border: `1px solid ${COLORS.border}`,
        overflow: "hidden",
        fontFamily: FONTS.mono,
        ...style,
      }}
    >
      {/* Title bar */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          padding: "10px 14px",
          backgroundColor: COLORS.secondary,
          borderBottom: `1px solid ${COLORS.border}`,
        }}
      >
        {TRAFFIC_LIGHTS.map((color, i) => (
          <div
            key={i}
            style={{ width: 12, height: 12, borderRadius: "50%", backgroundColor: color }}
          />
        ))}
        <span
          style={{
            marginLeft: 8,
            color: COLORS.mutedForeground,
            fontSize: 13,
            flex: 1,
            textAlign: "center",
          }}
        >
          {title}
        </span>
      </div>

      {/* Code lines */}
      <div style={{ padding: "16px 20px" }}>
        {lines.slice(0, visibleCount).map((line, i) => {
          const isLast = i === visibleCount - 1;
          const lineOpacity = isLast
            ? interpolate(
                frame % framesPerLine,
                [0, framesPerLine * 0.5],
                [0.3, 1],
                CLAMP
              )
            : 1;

          return (
            <div
              key={i}
              style={{
                display: "flex",
                gap: 20,
                lineHeight: 1.6,
                fontSize: 14,
                opacity: lineOpacity,
              }}
            >
              <span
                style={{
                  color: COLORS.mutedForeground,
                  userSelect: "none",
                  minWidth: 24,
                  textAlign: "right",
                  flexShrink: 0,
                }}
              >
                {i + 1}
              </span>
              <span style={{ color: COLORS.foreground }}>{line}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
