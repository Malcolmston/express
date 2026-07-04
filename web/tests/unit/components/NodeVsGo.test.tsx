import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { NodeVsGo } from '../../../src/components/NodeVsGo';
import { EXPRESS } from '../../../src/data';

describe('NodeVsGo', () => {
  it('renders the comparison heading and both language columns', () => {
    render(<NodeVsGo lib={EXPRESS} />);
    expect(screen.getByRole('heading', { name: /Node\.js/ })).toBeInTheDocument();
    expect(screen.getByText('Node.js')).toBeInTheDocument();
    expect(screen.getByText('Go')).toBeInTheDocument();
  });
});
