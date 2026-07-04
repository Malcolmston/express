import { CodeBlock } from 'go-ui';
import type { Lib } from '../data';

export interface InstallProps {
  lib: Lib;
}

// Install renders the `go get` command for the library as a glass code card.
export function Install({ lib }: InstallProps) {
  return (
    <>
      <div className="sec-h" id="install"><span className="bar" /><h3 style={{ margin: 0 }}>Install</h3></div>
      <CodeBlock lang="shell" html={`<span class="tok-c">$</span> go get ${lib.pkg}`} />
    </>
  );
}
