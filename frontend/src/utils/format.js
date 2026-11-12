const currency = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });

export const formatMoney = (cents) => currency.format(cents / 100);

export const formatDate = (iso) =>
  new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });

/**
 * Formats a 16-digit card number in groups of 4: "5342 0123 4567 8901".
 * If masked is true, masks middle 8 digits: "5342 •••• •••• 8901".
 */
export function formatCardNumber(raw, masked = false) {
  if (!raw) return '';
  const digits = String(raw).replace(/\D/g, '');
  if (digits.length === 0) return '';
  if (masked) {
    if (digits.length <= 4) return digits;
    const first4 = digits.slice(0, 4);
    const last4 = digits.slice(-4);
    return `${first4} •••• •••• ${last4}`;
  }
  return digits.replace(/(\d{4})(?=\d)/g, '$1 ').trim();
}

/**
 * Formats an IBAN into groups of 4 characters: "EG96 MIEB 0000 0001 2345 6789".
 */
export function formatIBAN(raw) {
  if (!raw) return '';
  const clean = String(raw).replace(/[^A-Za-z0-9]/g, '').toUpperCase();
  return clean.replace(/(.{4})(?=.)/g, '$1 ').trim();
}
