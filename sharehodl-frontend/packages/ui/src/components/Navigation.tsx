import Link from 'next/link';

export function Navigation() {
    const navItems = [
        { href: 'https://sharehodl.com', label: 'Governance', icon: '🏛️' },
        { href: 'https://trade.sharehodl.com', label: 'Trading', icon: '📈' },
        { href: 'https://scan.sharehodl.com', label: 'Explorer', icon: '🔍' },
        { href: 'https://wallet.sharehodl.com', label: 'Wallet', icon: '💼' },
        { href: 'https://business.sharehodl.com', label: 'Business', icon: '🏢' },
    ];

    return (
        <nav style={{
            backgroundColor: '#fff',
            borderBottom: '1px solid #e5e7eb',
            padding: '1rem 0',
            marginBottom: '2rem'
        }}>
            <div style={{
                maxWidth: '1200px',
                margin: '0 auto',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '0 2rem'
            }}>
                <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px'
                }}>
                    <span style={{ fontSize: '24px', fontWeight: 'bold' }}>ShareHODL</span>
                </div>
                
                <div style={{
                    display: 'flex',
                    gap: '2rem',
                    alignItems: 'center'
                }}>
                    {navItems.map((item) => (
                        <a
                            key={item.href}
                            href={item.href}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '6px',
                                textDecoration: 'none',
                                color: '#374151',
                                fontSize: '14px',
                                fontWeight: '500',
                                padding: '8px 12px',
                                borderRadius: '6px',
                                border: '1px solid #e5e7eb',
                                backgroundColor: '#f9fafb',
                                transition: 'all 0.2s'
                            }}
                            onMouseEnter={(e) => {
                                e.currentTarget.style.backgroundColor = '#f3f4f6';
                                e.currentTarget.style.borderColor = '#d1d5db';
                            }}
                            onMouseLeave={(e) => {
                                e.currentTarget.style.backgroundColor = '#f9fafb';
                                e.currentTarget.style.borderColor = '#e5e7eb';
                            }}
                        >
                            <span>{item.icon}</span>
                            <span>{item.label}</span>
                        </a>
                    ))}
                </div>
            </div>
        </nav>
    );
}