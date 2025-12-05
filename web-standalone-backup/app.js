// ShareHODL Atomic Swap Interface

class ShareHODLApp {
    constructor() {
        this.wallet = null;
        this.connected = false;
        this.slippageTolerance = 3;
        this.deadline = 30;
        this.expertMode = false;
        this.currentSection = 'swap';

        // Mock data for demonstration
        this.mockAssets = {
            'HODL': { balance: '1250.50', price: 1.00 },
            'APPLE': { balance: '15.75', price: 185.25 },
            'TSLA': { balance: '8.25', price: 245.80 },
            'GOOGL': { balance: '3.10', price: 2750.00 },
            'MSFT': { balance: '12.50', price: 385.60 }
        };

        this.mockMarkets = [
            { symbol: 'APPLE/HODL', lastPrice: 185.25, change: '+2.5%', volume: '125,890 HODL' },
            { symbol: 'TSLA/HODL', lastPrice: 245.80, change: '-1.2%', volume: '89,340 HODL' },
            { symbol: 'GOOGL/HODL', lastPrice: 2750.00, change: '+0.8%', volume: '45,230 HODL' },
            { symbol: 'MSFT/HODL', lastPrice: 385.60, change: '+3.1%', volume: '78,560 HODL' }
        ];

        this.mockSwapHistory = [
            { 
                time: '2024-12-04 14:30:22',
                from: '100 HODL',
                to: '0.539 APPLE',
                rate: '1 APPLE = 185.25 HODL',
                status: 'Completed'
            },
            {
                time: '2024-12-04 12:15:41',
                from: '500 HODL',
                to: '2.034 TSLA',
                rate: '1 TSLA = 245.80 HODL',
                status: 'Completed'
            }
        ];

        this.init();
    }

    init() {
        this.bindEvents();
        this.updateUI();
        this.loadMarkets();
        this.loadPortfolio();
    }

    bindEvents() {
        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                this.switchSection(e.target.getAttribute('href').substring(1));
            });
        });

        // Wallet connection
        document.getElementById('connectWallet').addEventListener('click', () => {
            this.connectWallet();
        });

        // Swap direction toggle
        document.getElementById('swapDirection').addEventListener('click', () => {
            this.swapAssetDirection();
        });

        // Asset selection
        document.getElementById('fromAsset').addEventListener('change', () => {
            this.updateFromAsset();
        });

        document.getElementById('toAsset').addEventListener('change', () => {
            this.updateToAsset();
        });

        // Amount input
        document.getElementById('fromAmount').addEventListener('input', () => {
            this.calculateSwapAmount();
        });

        // Max button
        document.getElementById('maxBtn').addEventListener('click', () => {
            this.setMaxAmount();
        });

        // Swap button
        document.getElementById('swapBtn').addEventListener('click', () => {
            this.executeSwap();
        });

        // Settings modal
        document.getElementById('settingsBtn').addEventListener('click', () => {
            this.openSettings();
        });

        document.getElementById('closeSettings').addEventListener('click', () => {
            this.closeSettings();
        });

        // Slippage buttons
        document.querySelectorAll('.slippage-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                this.setSlippage(parseFloat(btn.getAttribute('data-value')));
            });
        });

        // Custom slippage
        document.getElementById('customSlippage').addEventListener('input', (e) => {
            const value = parseFloat(e.target.value);
            if (value && value > 0 && value <= 50) {
                this.setSlippage(value);
            }
        });

        // Deadline
        document.getElementById('deadline').addEventListener('input', (e) => {
            this.deadline = parseInt(e.target.value) || 30;
        });

        // Expert mode
        document.getElementById('expertMode').addEventListener('change', (e) => {
            this.expertMode = e.target.checked;
        });
    }

    async connectWallet() {
        try {
            // Mock wallet connection - in real implementation would use Keplr/MetaMask
            this.showLoading('Connecting to wallet...');
            
            await this.delay(2000); // Simulate connection time

            this.wallet = {
                address: 'sharehodl1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0',
                hodlBalance: this.mockAssets['HODL'].balance
            };
            
            this.connected = true;
            this.hideLoading();
            this.updateWalletUI();
            this.showToast('Wallet connected successfully!', 'success');
            
        } catch (error) {
            this.hideLoading();
            this.showToast('Failed to connect wallet', 'error');
        }
    }

    updateWalletUI() {
        const connectBtn = document.getElementById('connectWallet');
        const walletInfo = document.getElementById('walletInfo');
        const walletAddress = document.getElementById('walletAddress');
        const hodlBalance = document.getElementById('hodlBalance');
        const swapBtn = document.getElementById('swapBtn');

        if (this.connected && this.wallet) {
            connectBtn.style.display = 'none';
            walletInfo.style.display = 'flex';
            walletAddress.textContent = this.wallet.address.substring(0, 12) + '...' + 
                                       this.wallet.address.substring(this.wallet.address.length - 8);
            hodlBalance.textContent = this.wallet.hodlBalance;
            swapBtn.textContent = 'Execute Atomic Swap';
            swapBtn.disabled = false;
        } else {
            connectBtn.style.display = 'block';
            walletInfo.style.display = 'none';
            swapBtn.textContent = 'Connect Wallet to Swap';
            swapBtn.disabled = true;
        }
    }

    switchSection(sectionId) {
        // Hide all sections
        document.querySelectorAll('section').forEach(section => {
            section.style.display = 'none';
        });

        // Show selected section
        document.getElementById(sectionId).style.display = 'block';

        // Update nav links
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`[href="#${sectionId}"]`).classList.add('active');

        this.currentSection = sectionId;
    }

    swapAssetDirection() {
        const fromAsset = document.getElementById('fromAsset');
        const toAsset = document.getElementById('toAsset');

        const fromValue = fromAsset.value;
        const toValue = toAsset.value;

        fromAsset.value = toValue;
        toAsset.value = fromValue;

        this.updateFromAsset();
        this.updateToAsset();
        this.calculateSwapAmount();
    }

    updateFromAsset() {
        const fromAsset = document.getElementById('fromAsset').value;
        const fromBalance = document.getElementById('fromBalance');
        
        if (this.mockAssets[fromAsset]) {
            fromBalance.textContent = this.mockAssets[fromAsset].balance + ' ' + fromAsset;
        }
        
        this.calculateSwapAmount();
    }

    updateToAsset() {
        const toAsset = document.getElementById('toAsset').value;
        const toBalance = document.getElementById('toBalance');
        
        if (this.mockAssets[toAsset]) {
            toBalance.textContent = this.mockAssets[toAsset].balance + ' ' + toAsset;
        }
        
        this.calculateSwapAmount();
    }

    calculateSwapAmount() {
        const fromAsset = document.getElementById('fromAsset').value;
        const toAsset = document.getElementById('toAsset').value;
        const fromAmount = parseFloat(document.getElementById('fromAmount').value) || 0;
        const toAmountInput = document.getElementById('toAmount');
        const exchangeRate = document.getElementById('exchangeRate');
        const estimatedFee = document.getElementById('estimatedFee');
        const priceImpact = document.getElementById('priceImpact');

        if (fromAmount === 0 || fromAsset === toAsset) {
            toAmountInput.value = '';
            exchangeRate.textContent = '-';
            return;
        }

        if (this.mockAssets[fromAsset] && this.mockAssets[toAsset]) {
            const fromPrice = this.mockAssets[fromAsset].price;
            const toPrice = this.mockAssets[toAsset].price;
            const rate = fromPrice / toPrice;
            const toAmount = fromAmount * rate;

            // Apply slippage
            const slippageAdjustedAmount = toAmount * (1 - this.slippageTolerance / 100);

            toAmountInput.value = slippageAdjustedAmount.toFixed(6);
            exchangeRate.textContent = `1 ${fromAsset} = ${rate.toFixed(6)} ${toAsset}`;
            
            // Calculate fee (0.3% of fromAmount)
            const fee = fromAmount * 0.003;
            estimatedFee.textContent = `${fee.toFixed(6)} HODL`;

            // Calculate price impact (mock calculation)
            const impact = Math.min((fromAmount / 10000) * 0.1, 5); // Max 5% impact
            priceImpact.textContent = `${impact.toFixed(2)}%`;
            priceImpact.className = impact < 1 ? 'impact-low' : impact < 3 ? 'impact-medium' : 'impact-high';
        }
    }

    setMaxAmount() {
        const fromAsset = document.getElementById('fromAsset').value;
        const fromAmount = document.getElementById('fromAmount');
        
        if (this.mockAssets[fromAsset]) {
            fromAmount.value = this.mockAssets[fromAsset].balance;
            this.calculateSwapAmount();
        }
    }

    async executeSwap() {
        if (!this.connected) {
            this.showToast('Please connect your wallet first', 'warning');
            return;
        }

        const fromAsset = document.getElementById('fromAsset').value;
        const toAsset = document.getElementById('toAsset').value;
        const fromAmount = parseFloat(document.getElementById('fromAmount').value);

        if (!fromAmount || fromAmount <= 0) {
            this.showToast('Please enter a valid amount', 'warning');
            return;
        }

        if (fromAsset === toAsset) {
            this.showToast('Cannot swap the same asset', 'warning');
            return;
        }

        try {
            this.showLoading('Executing atomic swap...');
            
            // Simulate swap execution
            await this.delay(3000);
            
            // Update mock balances
            const fromPrice = this.mockAssets[fromAsset].price;
            const toPrice = this.mockAssets[toAsset].price;
            const rate = fromPrice / toPrice;
            const toAmount = fromAmount * rate * (1 - this.slippageTolerance / 100);

            this.mockAssets[fromAsset].balance = (parseFloat(this.mockAssets[fromAsset].balance) - fromAmount).toFixed(6);
            this.mockAssets[toAsset].balance = (parseFloat(this.mockAssets[toAsset].balance) + toAmount).toFixed(6);

            this.hideLoading();
            this.showToast(`Successfully swapped ${fromAmount} ${fromAsset} for ${toAmount.toFixed(6)} ${toAsset}`, 'success');
            
            // Reset form
            document.getElementById('fromAmount').value = '';
            document.getElementById('toAmount').value = '';
            this.updateFromAsset();
            this.updateToAsset();
            
            // Add to swap history
            this.mockSwapHistory.unshift({
                time: new Date().toLocaleString(),
                from: `${fromAmount} ${fromAsset}`,
                to: `${toAmount.toFixed(6)} ${toAsset}`,
                rate: `1 ${fromAsset} = ${rate.toFixed(6)} ${toAsset}`,
                status: 'Completed'
            });

            // Update portfolio if on that section
            if (this.currentSection === 'portfolio') {
                this.loadPortfolio();
            }
            
        } catch (error) {
            this.hideLoading();
            this.showToast('Swap failed. Please try again.', 'error');
        }
    }

    openSettings() {
        document.getElementById('settingsModal').classList.add('show');
    }

    closeSettings() {
        document.getElementById('settingsModal').classList.remove('show');
    }

    setSlippage(value) {
        this.slippageTolerance = value;
        document.getElementById('slippageTolerance').textContent = value;
        
        // Update button states
        document.querySelectorAll('.slippage-btn').forEach(btn => {
            btn.classList.remove('active');
            if (parseFloat(btn.getAttribute('data-value')) === value) {
                btn.classList.add('active');
            }
        });

        this.calculateSwapAmount();
    }

    loadMarkets() {
        const tbody = document.getElementById('marketsTableBody');
        tbody.innerHTML = '';

        this.mockMarkets.forEach(market => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${market.symbol}</td>
                <td>$${market.lastPrice.toLocaleString()}</td>
                <td class="${market.change.startsWith('+') ? 'change-positive' : 'change-negative'}">${market.change}</td>
                <td>${market.volume}</td>
                <td><button class="btn-secondary" onclick="app.quickTrade('${market.symbol}')">Trade</button></td>
            `;
            tbody.appendChild(row);
        });
    }

    loadPortfolio() {
        // Update summary
        let totalValue = 0;
        Object.entries(this.mockAssets).forEach(([symbol, data]) => {
            totalValue += parseFloat(data.balance) * data.price;
        });

        document.getElementById('totalValue').textContent = '$' + totalValue.toLocaleString();
        document.getElementById('dayChange').textContent = '+2.35%'; // Mock
        document.getElementById('totalTrades').textContent = this.mockSwapHistory.length.toString();

        // Update holdings table
        const holdingsBody = document.getElementById('holdingsTableBody');
        holdingsBody.innerHTML = '';

        Object.entries(this.mockAssets).forEach(([symbol, data]) => {
            const value = parseFloat(data.balance) * data.price;
            const change = Math.random() > 0.5 ? '+' : '-';
            const changePercent = (Math.random() * 5).toFixed(2);

            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${symbol}</td>
                <td>${data.balance}</td>
                <td>${value.toFixed(2)} HODL</td>
                <td class="${change === '+' ? 'change-positive' : 'change-negative'}">${change}${changePercent}%</td>
                <td>
                    <button class="btn-secondary" onclick="app.quickSwap('${symbol}')">Swap</button>
                </td>
            `;
            holdingsBody.appendChild(row);
        });

        // Update swap history
        const historyBody = document.getElementById('swapHistoryBody');
        historyBody.innerHTML = '';

        this.mockSwapHistory.slice(0, 10).forEach(swap => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${swap.time}</td>
                <td>${swap.from}</td>
                <td>${swap.to}</td>
                <td>${swap.rate}</td>
                <td><span class="strategy-status active">${swap.status}</span></td>
            `;
            historyBody.appendChild(row);
        });
    }

    quickTrade(marketSymbol) {
        const [base, quote] = marketSymbol.split('/');
        this.switchSection('swap');
        document.getElementById('fromAsset').value = quote;
        document.getElementById('toAsset').value = base;
        this.updateFromAsset();
        this.updateToAsset();
    }

    quickSwap(asset) {
        this.switchSection('swap');
        document.getElementById('fromAsset').value = asset;
        document.getElementById('toAsset').value = asset === 'HODL' ? 'APPLE' : 'HODL';
        this.updateFromAsset();
        this.updateToAsset();
    }

    updateUI() {
        this.updateWalletUI();
    }

    showLoading(message = 'Loading...') {
        const overlay = document.getElementById('loadingOverlay');
        const text = overlay.querySelector('.loading-text');
        text.textContent = message;
        overlay.style.display = 'flex';
    }

    hideLoading() {
        document.getElementById('loadingOverlay').style.display = 'none';
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;

        container.appendChild(toast);

        // Auto remove after 5 seconds
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 5000);
    }

    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

// Initialize the app when page loads
let app;
document.addEventListener('DOMContentLoaded', () => {
    app = new ShareHODLApp();
});

// Global functions for quick access
window.app = null;
document.addEventListener('DOMContentLoaded', () => {
    window.app = new ShareHODLApp();
});

// Handle modal clicks outside content
document.addEventListener('click', (e) => {
    if (e.target.classList.contains('modal')) {
        e.target.classList.remove('show');
    }
});

// Handle escape key for modals
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        document.querySelectorAll('.modal.show').forEach(modal => {
            modal.classList.remove('show');
        });
    }
});

// Add some utility functions for demonstration
window.demoActions = {
    generateRandomSwap: () => {
        const assets = ['HODL', 'APPLE', 'TSLA', 'GOOGL', 'MSFT'];
        const fromAsset = assets[Math.floor(Math.random() * assets.length)];
        let toAsset = assets[Math.floor(Math.random() * assets.length)];
        while (toAsset === fromAsset) {
            toAsset = assets[Math.floor(Math.random() * assets.length)];
        }

        document.getElementById('fromAsset').value = fromAsset;
        document.getElementById('toAsset').value = toAsset;
        document.getElementById('fromAmount').value = (Math.random() * 100).toFixed(2);
        
        if (window.app) {
            window.app.updateFromAsset();
            window.app.updateToAsset();
            window.app.calculateSwapAmount();
        }
    },

    simulateMarketMovement: () => {
        // Randomly update asset prices
        Object.keys(window.app.mockAssets).forEach(symbol => {
            if (symbol !== 'HODL') {
                const currentPrice = window.app.mockAssets[symbol].price;
                const change = (Math.random() - 0.5) * 0.1; // ±5% max change
                window.app.mockAssets[symbol].price = currentPrice * (1 + change);
            }
        });
        
        // Update markets table if visible
        if (window.app.currentSection === 'markets') {
            window.app.loadMarkets();
        }
        
        // Recalculate swap if on swap page
        if (window.app.currentSection === 'swap') {
            window.app.calculateSwapAmount();
        }
    }
};

// Auto-update market prices every 10 seconds
setInterval(() => {
    if (window.demoActions && window.app) {
        window.demoActions.simulateMarketMovement();
    }
}, 10000);