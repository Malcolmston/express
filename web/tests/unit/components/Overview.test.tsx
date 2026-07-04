import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Overview } from '../../../src/components/Overview';
import { EXPRESS } from '../../../src/data';

describe('Overview', () => {
  beforeEach(() => {
    // Hero mounts VersionBadge which fetches; keep it pending.
    global.fetch = vi.fn().mockReturnValue(new Promise(() => {}));
  });

  it('renders the hero, blurb and every documentation section', () => {
    const { container } = render(<Overview lib={EXPRESS} />);
    expect(container.querySelector('#view-overview')).not.toBeNull();
    expect(screen.getByRole('heading', { level: 1, name: /Express/ })).toBeInTheDocument();
    expect(screen.getByText(EXPRESS.blurb)).toBeInTheDocument();
    for (const name of ['Install', 'Quick start', 'Going further', 'Features']) {
      expect(screen.getByRole('heading', { name })).toBeInTheDocument();
    }
    expect(container.querySelectorAll('ul.feat li').length).toBe(EXPRESS.features.length);
  });
});
