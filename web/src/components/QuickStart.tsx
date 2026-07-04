import { CodeBlock, hi } from 'go-ui';
import type { Lib } from '../data';

export interface QuickStartProps {
  lib: Lib;
}

// QuickStart renders the minimal "hello, server" example plus a "going further"
// snippet showing routers, middleware and streaming.
export function QuickStart({ lib }: QuickStartProps) {
  return (
    <>
      <div className="sec-h" id="quick"><span className="bar" /><h3 style={{ margin: 0 }}>Quick start</h3></div>
      <CodeBlock lang="main.go" html={hi(lib.go_code)} />

      <div className="sec-h" id="more"><span className="bar" /><h3 style={{ margin: 0 }}>Going further</h3></div>
      <CodeBlock lang="go" html={lib.integrate} />
    </>
  );
}
