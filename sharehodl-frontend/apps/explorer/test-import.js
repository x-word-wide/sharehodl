try {
    const ui = require('../../packages/ui/dist/index.js');
    console.log('Available exports:', Object.keys(ui));
    console.log('Header type:', typeof ui.Header);
    console.log('Header function exists:', !!ui.Header);
} catch (error) {
    console.error('Import error:', error.message);
}