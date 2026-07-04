import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { screen } from '@testing-library/react';
import { Features } from '../../../src/components/Features';
import { EXPRESS } from '../../../src/data';

describe('Features', () => {
  it('renders the features heading and one bullet per feature', () => {
    const { container } = render(<Features lib={EXPRESS} />);
    expect(screen.getByRole('heading', { name: 'Features' })).toBeInTheDocument();
    expect(container.querySelectorAll('ul.feat li').length).toBe(EXPRESS.features.length);
  });
});
