import React from "react";
import { useCurrentFrame, interpolate } from "remotion";
import { COLORS, FONTS, RADIUS } from "../lib/theme";

const CLAMP = { extrapolateLeft: "clamp", extrapolateRight: "clamp" } as const;

const TRAFFIC_LIGHTS = ["#ff5f57", "#ffbd2e", "#28ca41"];

interface TerminalOutputProps {
  lines: string[];
  /** Frames per character for typewriter effect. Default: 3 */
  framesPerChar?: number;
  /** Frames to pause after completing each line. Default: 10 */
  pausePerLine?: number;
  promptColor?: string;
  style?: React.CSSProperties;
}

export const TerminalOutput: React.FC<TerminalOutputProps> = ({
  lines,
  framesPerChar = 3,
  pausePerLine = 10,
  promptColor = COLORS.accentBlue,
  style,
}) => {
  const frame = useCurrentFrame();

  // Determine which line is active and how many chars are visible on it
  let frameAccumulator = 0;
  let activeLineIndex = lines.length - 1;
  let activeLineChars = lines[lines.length - 1]?.length ?? 0;

  for (let i = 0; i < lines.length; i++) {
    const lineFrames = lines[i].length * framesPerChar + pausePerLine;
    if (frame < frameAccumulator + lineFrames) {
      activeLineIndex = i;
      activeLineChars = Math.min(
        lines[i].length,
        Math.floor((frame - frameAccumulator) / framesPerChar)
      );
      break;
    }
    frameAccumulator += lineFrames;
  }

  const cursorVisible = Math.floor(frame / 15) % 2 === 0;
  const containerOpacity = interpolate(frame, [0, 8], [0, 1], CLAMP);

  return (
    <div
      style={{
        backgroundColor: COLORS.background,
        borderRadius: RADIUS.lg,
        border: `1px solid ${COLORS.border}`,
        fontFamily: FONTS.mono,
        fontSize: 14,
        padding: "16px 20px",
        opacity: containerOpacity,
        ...style,
      }}
    >
      {/* Traffic lights */}
      <div style={{ display: "flex", gap: 6, marginBottom: 16 }}>
        {TRAFFIC_LIGHTS.map((c, i) => (
          <div
            key={i}
            style={{ width: 12, height: 12, borderRadius: "50%", backgroundColor: c }}
          />
        ))}
      </div>

      {lines.slice(0, activeLineIndex + 1).map((line, i) => {
        const isActive = i === activeLineIndex;
        const visibleText = isActive ? line.slice(0, activeLineChars) : line;
        const showCursor =
          isActive && cursorVisible && activeLineIndex < lines.length;

        return (
          <div
            key={i}
            style={{ display: "flex", alignItems: "center", lineHeight: 1.8, gap: 8 }}
          >
            <span style={{ color: promptColor, flexShrink: 0 }}>$</span>
            <span style={{ color: COLORS.foreground }}>{visibleText}</span>
            {showCursor && (
              <span
                style={{
                  display: "inline-block",
                  width: 8,
                  height: 16,
                  backgroundColor: COLORS.foreground,
                }}
              />
            )}
          </div>
        );
      })}
    </div>
  );
};
