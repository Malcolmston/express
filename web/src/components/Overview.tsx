import type { Lib } from '../data';
import { Hero } from './Hero';
import { Install } from './Install';
import { QuickStart } from './QuickStart';
import { NodeVsGo } from './NodeVsGo';
import { Features } from './Features';

export interface OverviewProps {
  lib: Lib;
}

// Overview is the landing tab: the hero, an "on this page" jump nav, and the
// install / quick-start / comparison / features sections.
export function Overview({ lib }: OverviewProps) {
  return (
    <section className="view active" id="view-overview">
      <Hero lib={lib} />

      <p className="muted">{lib.blurb}</p>
      <div className="onthispage">
        <a href="#install">Install</a>
        <a href="#quick">Quick start</a>
        <a href="#cmp">Node &rarr; Go</a>
        <a href="#more">Going further</a>
        <a href="#feat">Features</a>
      </div>

      <Install lib={lib} />
      <QuickStart lib={lib} />
      <NodeVsGo lib={lib} />
      <Features lib={lib} />

      <div className="note">Full API reference &amp; runnable examples: <a href="./api/">the generated Go API docs</a>.</div>
    </section>
  );
}
