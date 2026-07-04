import { VersionBadge, ghrepo } from 'go-ui';
import type { Lib } from '../data';
import { Install } from './Install';
import { QuickStart } from './QuickStart';

export interface DocsViewProps {
  lib: Lib;
}

// DocsView is the "docs" tab: a prominent link to the generated Go API
// reference (served alongside this site at /api/), plus the install + usage
// snippets so a reader can get running without leaving the page.
export function DocsView({ lib }: DocsViewProps) {
  return (
    <section className="view active" id="view-docs">
      <div className="sec-h"><span className="bar" /><h2 style={{ margin: 0 }}>API documentation</h2></div>
      <p className="muted">The complete package-by-package Go API reference is generated from source and served alongside this site. It documents every exported type, function and method across the {lib.name} module and its 100+ helper packages.</p>

      <div className="actions" style={{ marginBottom: '1.4rem' }}>
        <a className="pill b" href="./api/"><i className="fa-solid fa-book" />&nbsp;Open the API reference</a>
        <a className="pill b" href={lib.repo} target="_blank" rel="noopener"><i className="fa-brands fa-github" />&nbsp;Source on GitHub</a>
        <VersionBadge repo={ghrepo(lib)} href={`${lib.repo}/releases`} />
      </div>

      <Install lib={lib} />
      <QuickStart lib={lib} />

      <div className="note">Looking for the full reference? <a href="./api/">Browse the generated Go API docs</a>.</div>
    </section>
  );
}
