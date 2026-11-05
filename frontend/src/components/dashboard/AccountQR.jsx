import { useEffect, useRef, useState } from 'react';
import QRCode from 'qrcode';

export default function AccountQR({ user, showDetails = true, size = 180 }) {
  const canvasRef = useRef(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!canvasRef.current || !user?.account_number) return;
    const payload = JSON.stringify({
      bank: 'MIeBank',
      account_number: user.account_number,
      name: user.name,
    });

    QRCode.toCanvas(canvasRef.current, payload, {
      width: size,
      margin: 2,
      color: {
        dark: '#163a9f',
        light: '#ffffff',
      },
    }).catch((err) => {
      console.error('QR code generation error:', err);
    });
  }, [user, size]);

  function handleCopy() {
    if (!user?.account_number) return;
    navigator.clipboard.writeText(user.account_number);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  function handleDownload() {
    if (!canvasRef.current) return;
    const link = document.createElement('a');
    link.download = `MIeBank-QR-${user.account_number}.png`;
    link.href = canvasRef.current.toDataURL('image/png');
    link.click();
  }

  return (
    <div className="account-qr-card">
      <div className="qr-wrapper">
        <canvas ref={canvasRef} className="qr-canvas" width={size} height={size} />
      </div>

      {showDetails && (
        <div className="qr-meta">
          <p className="qr-account-num">
            Account: <strong>{user?.account_number}</strong>
          </p>
          <p className="qr-account-holder">{user?.name}</p>
          <div className="qr-actions">
            <button type="button" className="btn btn-ghost-dark btn-sm" onClick={handleCopy}>
              {copied ? '✓ Copied!' : 'Copy Account #'}
            </button>
            <button type="button" className="btn btn-primary btn-sm" onClick={handleDownload}>
              Download QR
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
