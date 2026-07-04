import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DocsView } from '../../../src/components/DocsView';
import { EXPRESS } from '../../../src/data';

describe('DocsView', () => {
  beforeEach(() => {
    // VersionBadge fetches on mount; keep it pending.
    global.fetch = vi.fn().mockReturnValue(new Promise(() => {}));
  });

  it('renders the API heading, a link to the generated docs and the usage snippets', () => {
    const { container } = render(<DocsView lib={EXPRESS} />);
    expect(container.querySelector('#view-docs')).not.toBeNull();
    expect(screen.getByRole('heading', { level: 2, name: /API documentation/ })).toBeInTheDocument();
    const apiLink = screen.getByRole('link', { name: /Open the API reference/ });
    expect(apiLink).toHaveAttribute('href', './api/');
    // Install + QuickStart are embedded so a reader can get running in place.
    expect(screen.getByRole('heading', { name: 'Install' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Quick start' })).toBeInTheDocument();
  });
});
