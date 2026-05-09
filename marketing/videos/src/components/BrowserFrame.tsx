import React from "react";
import { useCurrentFrame, useVideoConfig } from "remotion";
import { springEnter, fadeIn } from "../lib/animations";
import { COLORS, RADIUS, FONTS } from "../lib/theme";

const TRAFFIC_LIGHTS = ["#ff5f57", "#ffbd2e", "#28ca41"];

interface BrowserFrameProps {
  url?: string;
  children: React.ReactNode;
  /** Animate entry with spring scale + fade. Default: true */
  animate?: boolean;
  style?: React.CSSProperties;
}

export const BrowserFrame: React.FC<BrowserFrameProps> = ({
  url = "localhost:3000",
  children,
  animate = true,
  style,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const scale = animate ? springEnter(frame, fps, { damping: 14, stiffness: 100 }) : 1;
  const opacity = animate ? fadeIn(frame, 10) : 1;

  return (
    <div
      style={{
        transform: `scale(${scale})`,
        opacity,
        transformOrigin: "top center",
        borderRadius: RADIUS.xl,
        overflow: "hidden",
        border: `1px solid ${COLORS.border}`,
        boxShadow: "0 25px 60px rgba(0,0,0,0.5)",
        ...style,
      }}
    >
      {/* Browser chrome */}
      <div
        style={{
          backgroundColor: "oklch(0.22 0 0)",
          padding: "10px 14px",
          display: "flex",
          alignItems: "center",
          gap: 10,
          borderBottom: `1px solid ${COLORS.border}`,
        }}
      >
        {TRAFFIC_LIGHTS.map((c, i) => (
          <div
            key={i}
            style={{ width: 12, height: 12, borderRadius: "50%", backgroundColor: c }}
          />
        ))}
        <div
          style={{
            flex: 1,
            backgroundColor: "oklch(0.18 0 0)",
            borderRadius: RADIUS.sm,
            padding: "4px 12px",
            fontSize: 12,
            fontFamily: FONTS.sans,
            color: COLORS.mutedForeground,
            textAlign: "center",
            border: `1px solid ${COLORS.border}`,
          }}
        >
          {url}
        </div>
      </div>

      {/* Content */}
      <div style={{ position: "relative", overflow: "hidden" }}>{children}</div>
    </div>
  );
};
