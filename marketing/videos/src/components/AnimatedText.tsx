import React from "react";
import { useCurrentFrame } from "remotion";
import { textRevealOpacity, textRevealY } from "../lib/animations";
import { FONTS } from "../lib/theme";

interface AnimatedTextProps {
  text: string;
  mode?: "words" | "letters";
  staggerFrames?: number;
  revealDuration?: number;
  style?: React.CSSProperties;
  className?: string;
}

export const AnimatedText: React.FC<AnimatedTextProps> = ({
  text,
  mode = "words",
  staggerFrames = 4,
  revealDuration = 12,
  style,
  className,
}) => {
  const frame = useCurrentFrame();
  const units = mode === "words" ? text.split(" ") : text.split("");

  return (
    <span className={className} style={{ fontFamily: FONTS.sans, ...style }}>
      {units.map((unit, i) => (
        <span
          key={i}
          style={{
            display: "inline-block",
            overflow: "hidden",
            marginRight: mode === "words" ? "0.25em" : 0,
          }}
        >
          <span
            style={{
              display: "inline-block",
              opacity: textRevealOpacity(frame, i, staggerFrames, revealDuration),
              transform: `translateY(${textRevealY(frame, i, staggerFrames, revealDuration)}px)`,
            }}
          >
            {unit}
          </span>
        </span>
      ))}
    </span>
  );
};
