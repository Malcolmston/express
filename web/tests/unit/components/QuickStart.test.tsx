import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QuickStart } from '../../../src/components/QuickStart';
import { EXPRESS } from '../../../src/data';

describe('QuickStart', () => {
  it('renders the quick start and going further sections with code', () => {
    render(<QuickStart lib={EXPRESS} />);
    expect(screen.getByRole('heading', { name: 'Quick start' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Going further' })).toBeInTheDocument();
    // The highlighted quick-start snippet includes the express.New() call.
    expect(screen.getByText(/express\.New\(\)/)).toBeInTheDocument();
  });
});
