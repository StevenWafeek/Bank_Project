import { describe, expect, it } from 'vitest';
import { formatCardNumber, formatIBAN, formatMoney } from './format.js';

describe('formatCardNumber', () => {
  it('formats clean 16 digits into 4 groups', () => {
    expect(formatCardNumber('5342012345678901')).toBe('5342 0123 4567 8901');
  });

  it('formats and masks middle digits', () => {
    expect(formatCardNumber('5342012345678901', true)).toBe('5342 •••• •••• 8901');
  });

  it('handles already spaced input correctly', () => {
    expect(formatCardNumber('5342 0123 4567 8901')).toBe('5342 0123 4567 8901');
  });

  it('handles empty or null gracefully', () => {
    expect(formatCardNumber('')).toBe('');
    expect(formatCardNumber(null)).toBe('');
  });
});

describe('formatIBAN', () => {
  it('formats IBAN with 4-character blocks', () => {
    expect(formatIBAN('EG96MIEB0000000123456789')).toBe('EG96 MIEB 0000 0001 2345 6789');
  });

  it('normalizes lowercase and removes existing spaces', () => {
    expect(formatIBAN('eg96 mieb 0000 0001 2345 6789')).toBe('EG96 MIEB 0000 0001 2345 6789');
  });

  it('handles empty or null gracefully', () => {
    expect(formatIBAN('')).toBe('');
    expect(formatIBAN(null)).toBe('');
  });
});

describe('formatMoney', () => {
  it('formats cents to currency string', () => {
    expect(formatMoney(10000)).toBe('$100.00');
    expect(formatMoney(0)).toBe('$0.00');
  });
});
