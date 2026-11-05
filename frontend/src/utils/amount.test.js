import { describe, expect, it } from 'vitest';
import { validateAmount, validatePaymentAmount } from './amount.js';

describe('validateAmount', () => {
  it.each(['20', '50', '1000000'])('accepts %s', (v) => {
    expect(validateAmount(v)).toBeNull();
  });

  it.each([
    ['', /valid amount/],
    ['0', /valid amount/],
    ['-5', /valid amount/],
    ['abc', /valid amount/],
    ['20.5', /whole number/],
    ['100.01', /whole number/],
    ['19', /minimum/],
    ['1000001', /limit/],
  ])('rejects %s', (v, msg) => {
    expect(validateAmount(v)).toMatch(msg);
  });
});

describe('validatePaymentAmount', () => {
  it.each(['1', '5', '20', '1000000'])('accepts %s', (v) => {
    expect(validatePaymentAmount(v, 1)).toBeNull();
  });

  it.each([
    ['', /valid amount/],
    ['0', /valid amount/],
    ['-5', /valid amount/],
    ['1.5', /whole number/],
    ['1000001', /limit/],
  ])('rejects %s', (v, msg) => {
    expect(validatePaymentAmount(v, 1)).toMatch(msg);
  });
});
