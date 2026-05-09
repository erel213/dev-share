import React from "react";
import { z } from "zod";
import {
  AbsoluteFill,
  Series,
  useCurrentFrame,
  useVideoConfig,
  Img,
  OffthreadVideo,
  staticFile,
} from "remotion";
import { springEnter, fadeIn, slideInX } from "../lib/animations";
import { COLORS, FONTS, BRAND, RADIUS } from "../lib/theme";
import { AnimatedText } from "../components/AnimatedText";
import { BrowserFrame } from "../components/BrowserFrame";

export const devShareDemoSchema = z.object({
  workspaceName: z.string().default("my-workspace"),
  language: z.enum(["go", "typescript"]).default("typescript"),
  accentColor: z.string().default(COLORS.accentBlue),
  showWorkflowVideo: z.boolean().default(false),
  showAdminVideo: z.boolean().default(false),
});

export type DevShareDemoProps = z.infer<typeof devShareDemoSchema>;

// ── Shared primitives ─────────────────────────────────────────────────────────

const ChapterLabel: React.FC<{ text: string; frame: number }> = ({
  text,
  frame,
}) => (
  <div
    style={{
      display: "inline-block",
      opacity: fadeIn(frame, 12),
      backgroundColor: COLORS.muted,
      borderRadius: RADIUS.md,
      padding: "5px 14px",
      fontSize: 12,
      fontWeight: 600,
      color: COLORS.mutedForeground,
      letterSpacing: "0.12em",
      textTransform: "uppercase" as const,
      fontFamily: FONTS.sans,
    }}
  >
    {text}
  </div>
);

// Bullet row: dot + text, fades in at `startFrame`
const Bullet: React.FC<{ text: string; frame: number; startFrame: number }> = ({
  text,
  frame,
  startFrame,
}) => (
  <div
    style={{
      opacity: fadeIn(frame - startFrame, 18),
      display: "flex",
      alignItems: "center",
      gap: 14,
    }}
  >
    <div
      style={{
        width: 8,
        height: 8,
        borderRadius: "50%",
        backgroundColor: COLORS.primary,
        flexShrink: 0,
      }}
    />
    <span
      style={{
        fontSize: 22,
        color: COLORS.mutedForeground,
        fontFamily: FONTS.sans,
        lineHeight: 1.4,
      }}
    >
      {text}
    </span>
  </div>
);

// Split layout: left = message, right = app teaser that slides in from right
const SplitScene: React.FC<{
  frame: number;
  fps: number;
  label?: string;
  heading: string;
  bullets: string[];
  url: string;
  children: React.ReactNode; // screenshot or video
}> = ({ frame, fps, label, heading, bullets, url, children }) => {
  const browserX = slideInX(frame - 20, 28, 60);
  const browserOpacity = fadeIn(frame - 20, 22);

  return (
    <AbsoluteFill
      style={{
        backgroundColor: COLORS.background,
        flexDirection: "row",
        alignItems: "center",
        padding: "0 80px",
        gap: 60,
      }}
    >
      {/* ── Message column ── */}
      <div
        style={{
          flex: "0 0 54%",
          display: "flex",
          flexDirection: "column",
          gap: 22,
        }}
      >
        {label && <ChapterLabel text={label} frame={frame} />}
        <div style={{ opacity: fadeIn(frame, 20) }}>
          <AnimatedText
            text={heading}
            style={{
              fontSize: 48,
              fontWeight: 700,
              color: COLORS.foreground,
              fontFamily: FONTS.sans,
              lineHeight: 1.2,
            }}
          />
        </div>
        {bullets.map((text, i) => (
          <Bullet
            key={i}
            text={text}
            frame={frame}
            startFrame={20 + i * 20}
          />
        ))}
      </div>

      {/* ── App teaser column ── */}
      <div
        style={{
          flex: 1,
          transform: `translateX(${browserX}px)`,
          opacity: browserOpacity,
        }}
      >
        <BrowserFrame url={url} animate={false}>
          {children}
        </BrowserFrame>
      </div>
    </AbsoluteFill>
  );
};

// ── Title scene (90 frames = 3s) ─────────────────────────────────────────────

const TitleScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const logoScale = springEnter(frame, fps, { damping: 10, stiffness: 80 });
  const logoOpacity = fadeIn(frame, 15);
  const taglineOpacity = fadeIn(frame - 20, 20);

  return (
    <AbsoluteFill
      style={{
        backgroundColor: COLORS.background,
        alignItems: "center",
        justifyContent: "center",
        flexDirection: "column",
        gap: 24,
      }}
    >
      <div
        style={{
          width: 80,
          height: 80,
          backgroundColor: COLORS.primary,
          borderRadius: RADIUS.lg,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontSize: 32,
          fontWeight: 700,
          color: COLORS.primaryForeground,
          fontFamily: FONTS.sans,
          transform: `scale(${logoScale})`,
          opacity: logoOpacity,
        }}
      >
        {BRAND.monogram}
      </div>
      <div style={{ opacity: logoOpacity, transform: `scale(${logoScale})` }}>
        <AnimatedText
          text={BRAND.name}
          style={{
            fontSize: 64,
            fontWeight: 700,
            color: COLORS.foreground,
            fontFamily: FONTS.sans,
          }}
        />
      </div>
      <div style={{ opacity: taglineOpacity }}>
        <AnimatedText
          text={BRAND.tagline}
          style={{
            fontSize: 24,
            color: COLORS.mutedForeground,
            fontFamily: FONTS.sans,
          }}
          staggerFrames={3}
        />
      </div>
    </AbsoluteFill>
  );
};

// ── Problem scene (210 frames = 7s) ──────────────────────────────────────────

const PROBLEM_LINES: { text: string; offset: number }[] = [
  { text: "Your team is blocked — waiting on environments.", offset: 0 },
  { text: "Manual setup. Wrong configs. Inconsistent state.", offset: 45 },
  { text: "Days of onboarding just to get a working setup.", offset: 90 },
  { text: "There's a better way.", offset: 145 },
];

const ProblemScene: React.FC = () => {
  const frame = useCurrentFrame();

  return (
    <AbsoluteFill
      style={{
        backgroundColor: COLORS.background,
        alignItems: "center",
        justifyContent: "center",
        flexDirection: "column",
        gap: 20,
        padding: 120,
      }}
    >
      <div style={{ marginBottom: 8 }}>
        <ChapterLabel text="The problem" frame={frame} />
      </div>
      {PROBLEM_LINES.map(({ text, offset }, i) => {
        const isLast = i === PROBLEM_LINES.length - 1;
        return (
          <div
            key={i}
            style={{ opacity: fadeIn(frame - offset, 18), textAlign: "center" }}
          >
            <AnimatedText
              text={text}
              style={{
                fontSize: isLast ? 44 : 26,
                fontWeight: isLast ? 700 : 400,
                color: isLast ? COLORS.foreground : COLORS.mutedForeground,
                fontFamily: FONTS.sans,
                textAlign: "center",
              }}
            />
          </div>
        );
      })}
    </AbsoluteFill>
  );
};

// ── Dashboard scene (150 frames = 5s) ────────────────────────────────────────

const DashboardScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  return (
    <SplitScene
      frame={frame}
      fps={fps}
      label="The solution"
      heading="Meet Dev Share"
      bullets={[
        "Shared environments, one workspace",
        "Admin control over who gets what",
        "Any developer, running in seconds",
      ]}
      url="localhost:3000"
    >
      <Img
        src={staticFile("screenshots/hero-dashboard.png")}
        style={{ width: "100%", display: "block" }}
      />
    </SplitScene>
  );
};

// ── Environments scene (180 frames = 6s) ─────────────────────────────────────

const EnvironmentsScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  return (
    <SplitScene
      frame={frame}
      fps={fps}
      heading={"All your team's\nenvironments. One view."}
      bullets={[
        "Status, owner, and template at a glance",
        "Know exactly who is running what",
        "No more asking around on Slack",
      ]}
      url="localhost:3000/environments"
    >
      <Img
        src={staticFile("screenshots/environments-hover.png")}
        style={{ width: "100%", display: "block" }}
      />
    </SplitScene>
  );
};

// ── User management scene (150 frames = 5s) ──────────────────────────────────

const UserManagementScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  return (
    <SplitScene
      frame={frame}
      fps={fps}
      heading={"New teammate?\nReady in under a minute."}
      bullets={[
        "Invite with a single click",
        "Assign role instantly — Admin or Member",
        "No IT tickets. No waiting days.",
      ]}
      url="localhost:3000/users"
    >
      <Img
        src={staticFile("screenshots/users-management.png")}
        style={{ width: "100%", display: "block" }}
      />
    </SplitScene>
  );
};

// ── Group permissions scene (180 frames = 6s) ─────────────────────────────────

const GroupPermissionsScene: React.FC<{ showVideo: boolean }> = ({
  showVideo,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  return (
    <SplitScene
      frame={frame}
      fps={fps}
      heading={"Set who gets what.\nOnce."}
      bullets={[
        "Group devs by team or project",
        "Each group gets exactly the templates they need",
        "Change access once — it applies everywhere",
      ]}
      url="localhost:3000/groups"
    >
      {showVideo ? (
        <OffthreadVideo
          src={staticFile("recordings/admin-provision-workflow.webm")}
          style={{ width: "100%", display: "block" }}
        />
      ) : (
        <Img
          src={staticFile("screenshots/groups-templates-dialog.png")}
          style={{ width: "100%", display: "block" }}
        />
      )}
    </SplitScene>
  );
};

// ── Workflow scene (360 frames = 12s) ────────────────────────────────────────

const WORKFLOW_STEPS = [
  "1  Pick a template",
  "2  Name your environment",
  "3  It's running",
];

const WorkflowScene: React.FC<{ showVideo: boolean }> = ({ showVideo }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const browserX = slideInX(frame - 20, 28, 60);
  const browserOpacity = fadeIn(frame - 20, 22);

  return (
    <AbsoluteFill
      style={{
        backgroundColor: COLORS.background,
        flexDirection: "row",
        alignItems: "center",
        padding: "0 80px",
        gap: 60,
      }}
    >
      {/* ── Steps column ── */}
      <div
        style={{
          flex: "0 0 54%",
          display: "flex",
          flexDirection: "column",
          gap: 24,
        }}
      >
        <div style={{ opacity: fadeIn(frame, 20) }}>
          <AnimatedText
            text={"Template → running\nenvironment. One click."}
            style={{
              fontSize: 48,
              fontWeight: 700,
              color: COLORS.foreground,
              fontFamily: FONTS.sans,
              lineHeight: 1.2,
            }}
          />
        </div>
        <div style={{ opacity: fadeIn(frame - 15, 18) }}>
          <AnimatedText
            text="No docs. No waiting on infra. No context switching."
            style={{
              fontSize: 20,
              color: COLORS.mutedForeground,
              fontFamily: FONTS.sans,
            }}
            staggerFrames={4}
          />
        </div>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 14,
            marginTop: 8,
          }}
        >
          {WORKFLOW_STEPS.map((step, i) => (
            <div
              key={i}
              style={{
                opacity: fadeIn(frame - (50 + i * 25), 18),
                backgroundColor: COLORS.muted,
                borderRadius: RADIUS.md,
                padding: "10px 20px",
                fontSize: 18,
                fontWeight: 600,
                color: COLORS.foreground,
                fontFamily: FONTS.sans,
                display: "inline-block",
                alignSelf: "flex-start",
              }}
            >
              {step}
            </div>
          ))}
        </div>
      </div>

      {/* ── Browser teaser ── */}
      <div
        style={{
          flex: 1,
          transform: `translateX(${browserX}px)`,
          opacity: browserOpacity,
        }}
      >
        <BrowserFrame url="localhost:3000/environments" animate={false}>
          {showVideo ? (
            <OffthreadVideo
              src={staticFile("recordings/create-environment-workflow.webm")}
              style={{ width: "100%", display: "block" }}
            />
          ) : (
            <Img
              src={staticFile("screenshots/environments-hover.png")}
              style={{ width: "100%", display: "block" }}
            />
          )}
        </BrowserFrame>
      </div>
    </AbsoluteFill>
  );
};

// ── Outro scene (150 frames = 5s) ─────────────────────────────────────────────

const OutroScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const scale = springEnter(frame, fps, { damping: 10 });
  const opacity = fadeIn(frame, 20);

  return (
    <AbsoluteFill
      style={{
        backgroundColor: COLORS.background,
        alignItems: "center",
        justifyContent: "center",
        flexDirection: "column",
        gap: 16,
      }}
    >
      <div
        style={{
          transform: `scale(${scale})`,
          opacity,
          textAlign: "center",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: 14,
        }}
      >
        <AnimatedText
          text={BRAND.name}
          style={{
            fontSize: 72,
            fontWeight: 700,
            color: COLORS.foreground,
            fontFamily: FONTS.sans,
          }}
        />
        <div style={{ opacity: fadeIn(frame - 15, 20) }}>
          <AnimatedText
            text="Your team. Their environments. Zero conflicts."
            style={{
              fontSize: 22,
              color: COLORS.mutedForeground,
              fontFamily: FONTS.sans,
            }}
            staggerFrames={3}
          />
        </div>
        <div
          style={{
            opacity: fadeIn(frame - 35, 20),
            marginTop: 8,
            backgroundColor: COLORS.primary,
            borderRadius: RADIUS.lg,
            padding: "10px 28px",
          }}
        >
          <AnimatedText
            text="Start free — devshare.io"
            style={{
              fontSize: 18,
              fontWeight: 600,
              color: COLORS.primaryForeground,
              fontFamily: FONTS.sans,
            }}
            staggerFrames={2}
          />
        </div>
      </div>
    </AbsoluteFill>
  );
};

// ── Composition root ──────────────────────────────────────────────────────────
// Total: 90 + 210 + 150 + 180 + 150 + 180 + 360 + 150 = 1470 frames

export const DevShareDemo: React.FC<DevShareDemoProps> = ({
  showWorkflowVideo,
  showAdminVideo,
}) => {
  return (
    <Series>
      <Series.Sequence durationInFrames={90} name="Title">
        <TitleScene />
      </Series.Sequence>
      <Series.Sequence durationInFrames={210} name="Problem">
        <ProblemScene />
      </Series.Sequence>
      <Series.Sequence durationInFrames={150} name="Dashboard">
        <DashboardScene />
      </Series.Sequence>
      <Series.Sequence durationInFrames={180} name="Environments">
        <EnvironmentsScene />
      </Series.Sequence>
      <Series.Sequence durationInFrames={150} name="UserManagement">
        <UserManagementScene />
      </Series.Sequence>
      <Series.Sequence durationInFrames={180} name="GroupPermissions">
        <GroupPermissionsScene showVideo={showAdminVideo} />
      </Series.Sequence>
      <Series.Sequence durationInFrames={360} name="Workflow">
        <WorkflowScene showVideo={showWorkflowVideo} />
      </Series.Sequence>
      <Series.Sequence durationInFrames={150} name="Outro">
        <OutroScene />
      </Series.Sequence>
    </Series>
  );
};
