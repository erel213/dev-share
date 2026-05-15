import React from "react";
import { Composition } from "remotion";
import {
  DevShareDemo,
  devShareDemoSchema,
} from "./compositions/DevShareDemo";
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
        durationInFrames={1650}
        schema={devShareDemoSchema}
        defaultProps={{
          workspaceName: "my-workspace",
          language: "typescript",
          accentColor: COLORS.accentBlue,
          showWorkflowVideo: false,
          showAdminVideo: false,
        }}
      />
    </>
  );
};
