export default function Message({ id, text, type = 'error' }) {
  return (
    <p id={id} className={`message ${text ? type : ''}`} role="alert">
      {text}
    </p>
  );
}
