import Benchmark from "../../check/Benchmark";
import Card from "../../Card";
import Container from "../Container";
import Control from "../../check/Control";
import ErrorPanel from "../../Error";
import Image from "../../Image";
import Panel from "../Panel";
import React from "react";
import Table from "../../Table";
import Text from "../../Text";
import TitleWrapper from "../../TitleWrapper";
import {
  ContainerDefinition,
  PanelDefinition,
} from "../../../../hooks/useDashboard";
import { Dashboard } from "../Dashboard";
import { RenderChart as Chart } from "../../charts/Chart";
import { RenderHierarchy as Hierarchy } from "../../hierarchies/Hierarchy";
import { RenderInput as Input } from "../../inputs/Input";

const ChildWithTitle = ({ child, level, renderChild }) => {
  return (
    <TitleWrapper definition={child} level={level} title={child.title}>
      {child.title ? null : renderChild()}
    </TitleWrapper>
  );
};

interface ChildrenProps {
  children: ContainerDefinition[] | PanelDefinition[] | undefined;
  showPanelExpand?: boolean;
}

const Children = ({ children = [], showPanelExpand = true }: ChildrenProps) => (
  <>
    {children.map((child) => {
      switch (child.node_type) {
        case "benchmark":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="container"
              renderChild={() => (
                <Panel definition={child} showExpand={showPanelExpand}>
                  <Benchmark {...child} />
                </Panel>
              )}
            />
          );
        case "card":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel definition={child} showExpand={showPanelExpand}>
                  <Card {...child} />
                </Panel>
              )}
            />
          );
        case "chart":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={!!child.data}
                  showExpand={showPanelExpand}
                >
                  <Chart {...child} />
                </Panel>
              )}
            />
          );
        case "container":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="container"
              renderChild={() => <Container definition={child} />}
            />
          );
        case "control":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel definition={child} showExpand={showPanelExpand}>
                  <Control {...child} />
                </Panel>
              )}
            />
          );
        case "dashboard":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="container"
              renderChild={() => <Dashboard definition={child} />}
            />
          );
        case "error":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel definition={child} showExpand={showPanelExpand}>
                  <ErrorPanel error={child.error} />
                </Panel>
              )}
            />
          );
        case "hierarchy":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={!!child.data}
                  showExpand={showPanelExpand}
                >
                  <Hierarchy {...child} />
                </Panel>
              )}
            />
          );
        case "image":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={child.sql ? !!child.data : !!child.properties.src}
                  showExpand={showPanelExpand}
                >
                  <Image {...child} />
                </Panel>
              )}
            />
          );
        case "input":
          return (
            <Panel
              key={child.name}
              definition={child}
              showExpand={showPanelExpand}
            >
              <Input {...child} />
            </Panel>
          );
        case "table":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={!!child.data}
                  showExpand={showPanelExpand}
                >
                  <Table {...child} />
                </Panel>
              )}
            />
          );
        case "text":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel definition={child} showExpand={showPanelExpand}>
                  <Text {...child} />
                </Panel>
              )}
            />
          );
        default:
          return null;
      }
    })}
  </>
);

export default Children;
