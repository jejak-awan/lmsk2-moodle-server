// LMS Manager Frontend JavaScript

// Global variables
let refreshInterval;
let isLoggedIn = false;

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

// Initialize application
function initializeApp() {
    // Check if we're on login page or dashboard
    if (document.getElementById('login-form')) {
        initializeLogin();
    } else if (document.getElementById('stats-grid')) {
        initializeDashboard();
    }
}

// Initialize login page
function initializeLogin() {
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
}

// Initialize dashboard
function initializeDashboard() {
    // Initialize theme
    initializeTheme();
    
    // Wait for icons to load, then render
    setTimeout(() => {
        renderAllIcons();
    }, 100);
    
    // Start auto-refresh
    startAutoRefresh();
    
    // Load initial data
    loadDashboardData();
    
    // Set up event listeners
    setupEventListeners();
    
    // Initialize navigation
    initializeNavigation();
}

// Handle login form submission
async function handleLogin(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const loginData = {
        username: formData.get('username'),
        password: formData.get('password')
    };
    
    showLoading();
    
    try {
        const response = await fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(loginData)
        });
        
        const result = await response.json();
        
        if (response.ok) {
            // Store token
            localStorage.setItem('auth_token', result.token);
            isLoggedIn = true;
            
            showToast('Login successful!', 'success');
            
            // Redirect to dashboard
            setTimeout(() => {
                window.location.href = '/';
            }, 1000);
        } else {
            showToast(result.error || 'Login failed', 'error');
        }
    } catch (error) {
        showToast('Network error. Please try again.', 'error');
    } finally {
        hideLoading();
    }
}

// Load dashboard data
async function loadDashboardData() {
    try {
        const response = await fetch('/api/stats', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            updateDashboard(data);
        } else if (response.status === 401) {
            // Redirect to login
            window.location.href = '/login';
        }
    } catch (error) {
        console.error('Failed to load dashboard data:', error);
    }
}

// Update dashboard with new data
function updateDashboard(data) {
    // Update system stats
    if (data.cpu_usage !== undefined) {
        updateStatCard('cpu-usage', data.cpu_usage + '%');
        updateStatBar('cpu-bar', data.cpu_usage);
    }
    
    if (data.memory_usage !== undefined) {
        updateStatCard('memory-usage', data.memory_usage + '%');
        updateStatBar('memory-bar', data.memory_usage);
    }
    
    if (data.disk_usage !== undefined) {
        updateStatCard('disk-usage', data.disk_usage + '%');
        updateStatBar('disk-bar', data.disk_usage);
    }
    
    if (data.uptime !== undefined) {
        updateStatCard('uptime', formatDuration(data.uptime));
    }
    
    // Update Moodle status
    if (data.moodle_status) {
        updateMoodleStatus(data.moodle_status);
    }
}

// Update stat card
function updateStatCard(elementId, value) {
    const element = document.getElementById(elementId);
    if (element) {
        element.textContent = value;
    }
}

// Update stat bar
function updateStatBar(elementId, percentage) {
    const element = document.getElementById(elementId);
    if (element) {
        element.style.width = percentage + '%';
        
        // Change color based on percentage
        if (percentage > 80) {
            element.style.background = 'linear-gradient(90deg, #dc3545, #c82333)';
        } else if (percentage > 60) {
            element.style.background = 'linear-gradient(90deg, #ffc107, #e0a800)';
        } else {
            element.style.background = 'linear-gradient(90deg, #28a745, #20c997)';
        }
    }
}

// Update Moodle status
function updateMoodleStatus(status) {
    const statusDot = document.getElementById('moodle-status-dot');
    const statusText = document.getElementById('moodle-status-text');
    const versionElement = document.getElementById('moodle-version');
    const uptimeElement = document.getElementById('moodle-uptime');
    const pidElement = document.getElementById('moodle-pid');
    
    if (statusDot && statusText) {
        if (status.running) {
            statusDot.classList.add('running');
            statusText.textContent = 'Running';
        } else {
            statusDot.classList.remove('running');
            statusText.textContent = 'Stopped';
        }
    }
    
    if (versionElement && status.version) {
        versionElement.textContent = status.version;
    }
    
    if (uptimeElement && status.uptime) {
        uptimeElement.textContent = formatDuration(status.uptime);
    }
    
    if (pidElement && status.process_id) {
        pidElement.textContent = status.process_id;
    }
}

// Start Moodle
async function startMoodle() {
    showLoading();
    
    try {
        const response = await fetch('/api/moodle/start', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            }
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showToast(result.message || 'Moodle started successfully', 'success');
            // Refresh status after a delay
            setTimeout(loadDashboardData, 2000);
        } else {
            showToast(result.error || 'Failed to start Moodle', 'error');
        }
    } catch (error) {
        showToast('Network error. Please try again.', 'error');
    } finally {
        hideLoading();
    }
}

// Stop Moodle
async function stopMoodle() {
    if (!confirm('Are you sure you want to stop Moodle?')) {
        return;
    }
    
    showLoading();
    
    try {
        const response = await fetch('/api/moodle/stop', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            }
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showToast(result.message || 'Moodle stopped successfully', 'success');
            // Refresh status after a delay
            setTimeout(loadDashboardData, 2000);
        } else {
            showToast(result.error || 'Failed to stop Moodle', 'error');
        }
    } catch (error) {
        showToast('Network error. Please try again.', 'error');
    } finally {
        hideLoading();
    }
}

// Restart Moodle
async function restartMoodle() {
    if (!confirm('Are you sure you want to restart Moodle?')) {
        return;
    }
    
    showLoading();
    
    try {
        const response = await fetch('/api/moodle/restart', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            }
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showToast(result.message || 'Moodle restarted successfully', 'success');
            // Refresh status after a delay
            setTimeout(loadDashboardData, 3000);
        } else {
            showToast(result.error || 'Failed to restart Moodle', 'error');
        }
    } catch (error) {
        showToast('Network error. Please try again.', 'error');
    } finally {
        hideLoading();
    }
}

// Refresh alerts
async function refreshAlerts() {
    try {
        const response = await fetch('/api/alerts', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            }
        });
        
        if (response.ok) {
            const alerts = await response.json();
            updateAlertsList(alerts);
        }
    } catch (error) {
        console.error('Failed to load alerts:', error);
    }
}

// Update alerts list
function updateAlertsList(alerts) {
    const alertsList = document.getElementById('alerts-list');
    if (!alertsList) return;
    
    if (alerts.length === 0) {
        alertsList.innerHTML = '<p class="text-center">No alerts</p>';
        return;
    }
    
    alertsList.innerHTML = alerts.map(alert => `
        <div class="alert-item ${alert.severity}">
            <div class="alert-header">
                <span class="alert-type">${alert.type}</span>
                <span class="alert-time">${formatTimestamp(alert.timestamp)}</span>
            </div>
            <div class="alert-message">${alert.message}</div>
        </div>
    `).join('');
}

// Refresh logs
async function refreshLogs() {
    try {
        const response = await fetch('/api/logs?limit=50', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            }
        });
        
        if (response.ok) {
            const logs = await response.json();
            updateLogsList(logs);
        }
    } catch (error) {
        console.error('Failed to load logs:', error);
    }
}

// Update logs list
function updateLogsList(logs) {
    const logsList = document.getElementById('logs-list');
    if (!logsList) return;
    
    if (logs.length === 0) {
        logsList.innerHTML = '<p class="text-center">No logs</p>';
        return;
    }
    
    logsList.innerHTML = logs.map(log => `
        <div class="log-item">
            <span class="log-time">${formatTimestamp(log.timestamp)}</span>
            <span class="log-level ${log.level}">${log.level.toUpperCase()}</span>
            <span class="log-message">${log.message}</span>
        </div>
    `).join('');
}

// Start auto-refresh
function startAutoRefresh() {
    // Refresh every 30 seconds
    refreshInterval = setInterval(() => {
        loadDashboardData();
        refreshAlerts();
        refreshLogs();
    }, 30000);
}

// Stop auto-refresh
function stopAutoRefresh() {
    if (refreshInterval) {
        clearInterval(refreshInterval);
    }
}

// Setup event listeners
function setupEventListeners() {
    // Logout button
    const logoutBtn = document.querySelector('button[onclick="logout()"]');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', logout);
    }
    
    // Moodle action buttons
    const startBtn = document.querySelector('button[onclick="startMoodle()"]');
    if (startBtn) {
        startBtn.addEventListener('click', startMoodle);
    }
    
    const stopBtn = document.querySelector('button[onclick="stopMoodle()"]');
    if (stopBtn) {
        stopBtn.addEventListener('click', stopMoodle);
    }
    
    const restartBtn = document.querySelector('button[onclick="restartMoodle()"]');
    if (restartBtn) {
        restartBtn.addEventListener('click', restartMoodle);
    }
    
    // Refresh buttons
    const refreshAlertsBtn = document.querySelector('button[onclick="refreshAlerts()"]');
    if (refreshAlertsBtn) {
        refreshAlertsBtn.addEventListener('click', refreshAlerts);
    }
    
    const refreshLogsBtn = document.querySelector('button[onclick="refreshLogs()"]');
    if (refreshLogsBtn) {
        refreshLogsBtn.addEventListener('click', refreshLogs);
    }
}

// Logout function
async function logout() {
    try {
        await fetch('/logout', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            }
        });
    } catch (error) {
        console.error('Logout error:', error);
    } finally {
        // Clear token and redirect
        localStorage.removeItem('auth_token');
        window.location.href = '/login';
    }
}

// Show loading overlay
function showLoading() {
    const overlay = document.getElementById('loading-overlay');
    if (overlay) {
        overlay.style.display = 'flex';
    }
}

// Hide loading overlay
function hideLoading() {
    const overlay = document.getElementById('loading-overlay');
    if (overlay) {
        overlay.style.display = 'none';
    }
}

// Show toast notification
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    if (!container) return;
    
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

// Format duration
function formatDuration(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (days > 0) {
        return `${days}d ${hours}h ${minutes}m`;
    } else if (hours > 0) {
        return `${hours}h ${minutes}m`;
    } else if (minutes > 0) {
        return `${minutes}m ${secs}s`;
    } else {
        return `${secs}s`;
    }
}

// Format timestamp
function formatTimestamp(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleString();
}

// Utility functions
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Error handling
window.addEventListener('error', function(e) {
    console.error('JavaScript error:', e.error);
    showToast('An error occurred. Please refresh the page.', 'error');
});

// Handle unhandled promise rejections
window.addEventListener('unhandledrejection', function(e) {
    console.error('Unhandled promise rejection:', e.reason);
    showToast('An error occurred. Please try again.', 'error');
});

// Cleanup on page unload
window.addEventListener('beforeunload', function() {
    stopAutoRefresh();
});

// Navigation Functions
function initializeNavigation() {
    // Set active nav item
    const activeItem = document.querySelector('.nav-item.active');
    if (activeItem) {
        activeItem.classList.add('active');
    }
}

function showDashboard() {
    setActiveNavItem('Dashboard');
    updateNavbarTitle('Dashboard');
    // Show dashboard content
    showContent('dashboard');
}

function showSystem() {
    setActiveNavItem('System Status');
    updateNavbarTitle('System Status');
    showContent('system');
}

function showMoodle() {
    setActiveNavItem('Moodle Management');
    updateNavbarTitle('Moodle Management');
    showContent('moodle');
}

function showBackups() {
    setActiveNavItem('Backups');
    updateNavbarTitle('Backups');
    showContent('backups');
}

function showLogs() {
    setActiveNavItem('System Logs');
    updateNavbarTitle('System Logs');
    showContent('logs');
}

function showAlerts() {
    setActiveNavItem('Alerts');
    updateNavbarTitle('Alerts');
    showContent('alerts');
}

function showUsers() {
    setActiveNavItem('Users');
    updateNavbarTitle('User Management');
    showContent('users');
}

function showSettings() {
    setActiveNavItem('Settings');
    updateNavbarTitle('Settings');
    showContent('settings');
}

function setActiveNavItem(itemText) {
    // Remove active class from all nav items
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });
    
    // Add active class to clicked item
    document.querySelectorAll('.nav-item').forEach(item => {
        if (item.textContent.trim() === itemText) {
            item.classList.add('active');
        }
    });
}

function updateNavbarTitle(title) {
    const navbarTitle = document.querySelector('.navbar-title');
    if (navbarTitle) {
        navbarTitle.textContent = title;
    }
}

function showContent(contentType) {
    // Hide all content sections
    const sections = document.querySelectorAll('.section');
    sections.forEach(section => {
        section.style.display = 'none';
    });
    
    // Show specific content based on type
    // For now, show all sections (dashboard view)
    if (contentType === 'dashboard') {
        sections.forEach(section => {
            section.style.display = 'block';
        });
    }
    
    // Load specific data if needed
    if (contentType === 'logs') {
        refreshLogs();
    } else if (contentType === 'alerts') {
        refreshAlerts();
    }
}

// Theme Functions
function initializeTheme() {
    // Get saved theme or default to light
    const savedTheme = localStorage.getItem('theme') || 'light';
    setTheme(savedTheme);
}

function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    setTheme(newTheme);
}

function setTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
    
    // Update theme icon
    const themeIcon = document.getElementById('theme-icon');
    if (themeIcon) {
        themeIcon.innerHTML = theme === 'dark' ? getIcon('sun') : getIcon('moon');
    }
}

// Icon Functions
function renderAllIcons() {
    // Check if Icons object exists
    if (typeof Icons === 'undefined') {
        console.error('Icons object not found. Make sure icons.js is loaded.');
        return;
    }
    
    // Render all icons with data-icon attribute
    const iconElements = document.querySelectorAll('[data-icon]');
    console.log(`Found ${iconElements.length} icon elements to render`);
    
    iconElements.forEach(element => {
        const iconName = element.getAttribute('data-icon');
        console.log(`Rendering icon: ${iconName}`);
        
        if (iconName && Icons[iconName]) {
            element.innerHTML = Icons[iconName];
            console.log(`Successfully rendered icon: ${iconName}`);
        } else {
            console.warn(`Icon not found: ${iconName}`);
        }
    });
}

function renderIcon(element, iconName) {
    if (element && Icons[iconName]) {
        element.innerHTML = Icons[iconName];
    }
}
