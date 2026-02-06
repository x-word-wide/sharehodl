/**
 * Loading Screen Component
 * Uses inline styles for reliability
 */

export function LoadingScreen() {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '100vh',
      backgroundColor: '#0D1117'
    }}>
      {/* Logo */}
      <div style={{
        width: '80px',
        height: '80px',
        borderRadius: '50%',
        background: 'linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        marginBottom: '24px',
        animation: 'pulse 2s ease-in-out infinite'
      }}>
        <span style={{ fontSize: '24px', fontWeight: 'bold', color: 'white' }}>SH</span>
      </div>

      {/* App name */}
      <h1 style={{ fontSize: '24px', fontWeight: 'bold', color: 'white', marginBottom: '8px' }}>ShareHODL</h1>
      <p style={{ color: '#8b949e', fontSize: '14px', marginBottom: '32px' }}>Tokenized Equity Trading</p>

      {/* Loading spinner */}
      <div style={{
        width: '24px',
        height: '24px',
        border: '2px solid rgba(255, 255, 255, 0.2)',
        borderTopColor: 'white',
        borderRadius: '50%',
        animation: 'spin 1s linear infinite'
      }} />

      <style>{`
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.7; }
        }
      `}</style>
    </div>
  );
}
