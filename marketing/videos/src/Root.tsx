import React from "react";
import { Composition } from "remotion";
import {
  DevShareDemo,
  devShareDemoSchema,
} from "./compositions/DevShareDemo";
import {
  FeatureWalkthrough,
  featureWalkthroughSchema,
} from "./compositions/FeatureWalkthrough";
import { SocialClip, socialClipSchema } from "./compositions/SocialClip";
import { COLORS } from "./lib/theme";


export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id="DevShareDemo"
        component={DevShareDemo}
        width={1920}
        height={1080}
        fps={30}
        durationInFrames={1470}
        schema={devShareDemoSchema}
        defaultProps={{
          workspaceName: "my-workspace",
          language: "typescript",
          accentColor: COLORS.accentBlue,
          showWorkflowVideo: false,
          showAdminVideo: false,
        }}
      />

      <Composition
        id="FeatureWalkthrough"
        component={FeatureWalkthrough}
        width={1920}
        height={1080}
        fps={30}
        durationInFrames={600}
        schema={featureWalkthroughSchema}
        defaultProps={{
          featureName: "Environment Creation",
          subtitle: "From template to running environment in seconds.",
          codeLines: [
            'import { createEnvironment } from "@devshare/sdk"',
            "",
            "const env = await createEnvironment({",
            '  template: "sample-infrastructure",',
            '  name: "my-dev-env",',
            "});",
          ],
          terminalLines: [
            "npx devshare env create --template sample-infrastructure",
            "Environment ready in 4.2s",
          ],
        }}
      />

      <Composition
        id="SocialClip"
        component={SocialClip}
        width={1080}
        height={1920}
        fps={30}
        durationInFrames={450}
        schema={socialClipSchema}
        defaultProps={{
          headline: "Infrastructure as Code\nfor your whole team.",
          featureLine: "Spin up dev environments in seconds.",
          ctaText: "Try Dev Share",
          ctaUrl: "devshare.io",
          accentColor: COLORS.accentBlue,
        }}
      />
    </>
  );
};
