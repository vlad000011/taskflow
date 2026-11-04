import { describe, expect, it } from 'vitest';
import { statuses } from './types';

describe('TaskFlow domain helpers', () => {
  it('defines five workflow columns', () => { expect(statuses).toHaveLength(5); expect(statuses.map(s=>s.key)).toContain('done'); });
  it('contains expected user-facing status labels', () => { expect(statuses.find(s=>s.key==='in_progress')?.label).toBe('В работе'); });
  it('has unique status identifiers', () => { expect(new Set(statuses.map(s=>s.key)).size).toBe(statuses.length); });
});
