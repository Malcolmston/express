import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Hero } from '../../../src/components/Hero';
import { EXPRESS } from '../../../src/data';

describe('Hero', () => {
  beforeEach(() => {
    // VersionBadge fetches on mount; keep it pending so the hero renders cleanly.
    global.fetch = vi.fn().mockReturnValue(new Promise(() => {}));
  });

  it('renders the library name, package and tagline', () => {
    render(<Hero lib={EXPRESS} />);
    expect(screen.getByRole('heading', { level: 1, name: /Express/ })).toBeInTheDocument();
    expect(screen.getByText(EXPRESS.pkg)).toBeInTheDocument();
    expect(screen.getByText(EXPRESS.tagline)).toBeInTheDocument();
  });

  it('links to GitHub (new tab) and the generated API docs', () => {
    render(<Hero lib={EXPRESS} />);
    const github = screen.getByRole('link', { name: /GitHub/ });
    expect(github).toHaveAttribute('href', EXPRESS.repo);
    expect(github).toHaveAttribute('target', '_blank');
    expect(github).toHaveAttribute('rel', expect.stringContaining('noopener'));

    const api = screen.getByRole('link', { name: /API docs/ });
    expect(api).toHaveAttribute('href', './api/');
  });
});
