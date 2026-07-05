import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DocsView } from '../../../src/components/DocsView';
import { EXPRESS } from '../../../src/data';
import type { DocIndex } from 'go-ui';

// A minimal DocIndex the stubbed fetch returns for DocsApp's doc.json request.
const DOC_INDEX: DocIndex = {
  module: 'github.com/malcolmston/express',
  packages: [
    {
      importPath: 'github.com/malcolmston/express',
      name: 'express',
      synopsis: 'Package express is a web framework.',
      doc: 'Package express is a web framework.',
      consts: [],
      vars: [],
      types: [
        {
          name: 'Router',
          signature: 'type Router struct{}',
          doc: 'Router routes requests.',
          consts: [],
          vars: [],
          funcs: [],
          methods: [],
        },
      ],
      funcs: [{ name: 'New', signature: 'func New() *App', doc: 'New creates an app.' }],
    },
  ],
};

describe('DocsView', () => {
  beforeEach(() => {
    // DocsApp fetches doc.json; return the small index. VersionBadge also fetches
    // (releases) — leave any non-doc request pending so it never resolves.
    global.fetch = vi.fn((input: RequestInfo | URL) => {
      if (String(input).includes('doc.json')) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(DOC_INDEX) } as Response);
      }
      return new Promise<Response>(() => {});
    }) as unknown as typeof fetch;
  });

  it('renders the inline React API reference from the fetched doc.json', async () => {
    const { container } = render(<DocsView lib={EXPRESS} />);
    expect(container.querySelector('#view-docs')).not.toBeNull();
    expect(
      screen.getByRole('heading', { level: 2, name: /API documentation/ }),
    ).toBeInTheDocument();

    // DocsApp fetches asynchronously, then renders the package view + symbols.
    expect(await screen.findByRole('heading', { name: /package express/ })).toBeInTheDocument();
    expect(container.querySelector('#sym-New'), 'func New symbol card').not.toBeNull();
    expect(container.querySelector('#sym-Router'), 'type Router symbol card').not.toBeNull();

    // The secondary link to the raw generated static HTML remains.
    expect(screen.getByRole('link', { name: /Raw generated HTML/ })).toHaveAttribute('href', './api/');
  });
});
