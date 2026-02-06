/**
 * Error Boundary Component
 * Catches React errors and displays a fallback UI
 */

import { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
    errorInfo: null
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error, errorInfo: null };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.setState({ error, errorInfo });
  }

  private handleReset = () => {
    // Clear localStorage and reload
    try {
      localStorage.clear();
    } catch (e) {
      console.error('Failed to clear localStorage:', e);
    }
    window.location.reload();
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div style={{
          minHeight: '100vh',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '20px',
          backgroundColor: '#0D1117',
          color: 'white',
          textAlign: 'center'
        }}>
          <h1 style={{ fontSize: '24px', marginBottom: '16px' }}>Something went wrong</h1>
          <p style={{ color: '#8b949e', marginBottom: '20px' }}>
            The app encountered an error. Please try resetting.
          </p>

          {this.state.error && (
            <div style={{
              background: '#161B22',
              padding: '16px',
              borderRadius: '8px',
              marginBottom: '20px',
              maxWidth: '100%',
              overflow: 'auto',
              textAlign: 'left'
            }}>
              <p style={{ color: '#ef4444', fontFamily: 'monospace', fontSize: '12px', whiteSpace: 'pre-wrap' }}>
                {this.state.error.toString()}
              </p>
              {this.state.errorInfo && (
                <p style={{ color: '#8b949e', fontFamily: 'monospace', fontSize: '10px', marginTop: '8px', whiteSpace: 'pre-wrap' }}>
                  {this.state.errorInfo.componentStack}
                </p>
              )}
            </div>
          )}

          <button
            onClick={this.handleReset}
            style={{
              padding: '12px 24px',
              background: 'linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%)',
              border: 'none',
              borderRadius: '12px',
              color: 'white',
              fontSize: '16px',
              fontWeight: '600',
              cursor: 'pointer'
            }}
          >
            Reset App
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}
