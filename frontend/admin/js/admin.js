/**
 * Astrolink Admin Panel Logic
 */

// State
let authToken = localStorage.getItem('astrolink_admin_token');
let idleTime = 0;
let idleInterval;
const TIMEOUT_MINUTES = 15;

function resetTimer() {
    idleTime = 0;
}

function startIdleTimer() {
    document.addEventListener('mousemove', resetTimer);
    document.addEventListener('keypress', resetTimer);
    document.addEventListener('click', resetTimer);
    document.addEventListener('scroll', resetTimer);

    if (idleInterval) clearInterval(idleInterval);
    idleInterval = setInterval(() => {
        if (!authToken) return;
        idleTime++;
        if (idleTime >= TIMEOUT_MINUTES * 60) {
            showToast('Sessão expirada por inatividade', 'error');
            logout();
        }
    }, 1000);
}

function stopIdleTimer() {
    if (idleInterval) clearInterval(idleInterval);
    document.removeEventListener('mousemove', resetTimer);
    document.removeEventListener('keypress', resetTimer);
    document.removeEventListener('click', resetTimer);
    document.removeEventListener('scroll', resetTimer);
}

// DOM Elements
const loginContainer = document.getElementById('login-container');
const dashboardLayout = document.getElementById('dashboard-layout');
const loginForm = document.getElementById('login-form');
const loginError = document.getElementById('login-error');

// Initialization
document.addEventListener('DOMContentLoaded', () => {
    // Theme Init
    if (localStorage.theme === 'light') {
        document.documentElement.classList.remove('dark');
        const btnIcon = document.querySelector('#btn-theme-toggle i');
        if (btnIcon) {
            btnIcon.classList.remove('fa-sun');
            btnIcon.classList.add('fa-moon');
        }
    } else {
        document.documentElement.classList.add('dark');
        const btnIcon = document.querySelector('#btn-theme-toggle i');
        if (btnIcon) {
            btnIcon.classList.remove('fa-moon');
            btnIcon.classList.add('fa-sun');
        }
    }

    // Nav Event Listeners
    document.querySelectorAll('.nav-item').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const target = link.getAttribute('data-target');
            if (target) nav(target);
        });
    });

    if (authToken) {
        // Try to load dashboard. If token is invalid, backend will return 401 and we logout
        showDashboard();
        loadDashboardStats();
        startIdleTimer();
    } else {
        showLogin();
    }
});

function toggleTheme() {
    const btnIcon = document.querySelector('#btn-theme-toggle i');
    if (document.documentElement.classList.contains('dark')) {
        document.documentElement.classList.remove('dark');
        localStorage.theme = 'light';
        if (btnIcon) {
            btnIcon.classList.remove('fa-sun');
            btnIcon.classList.add('fa-moon');
        }
    } else {
        document.documentElement.classList.add('dark');
        localStorage.theme = 'dark';
        if (btnIcon) {
            btnIcon.classList.remove('fa-moon');
            btnIcon.classList.add('fa-sun');
        }
    }
}

// Navigation
function nav(section) {
    const activeClasses = ['bg-cyan-500/10', 'border-cyan-400', 'text-cyan-400', 'drop-shadow-[0_0_8px_rgba(34,211,238,0.5)]', 'rounded-r-xl'];
    const inactiveClasses = ['text-slate-400', 'hover:bg-slate-800/50', 'hover:text-cyan-300', 'border-transparent', 'hover:border-slate-700', 'rounded-xl'];

    // Hide all sections
    ['overview', 'vouchers', 'planos', 'users', 'reports', 'walledgarden', 'blacklist', 'contas-fixas', 'logs', 'settings'].forEach(s => {
        const sec = document.getElementById(`sec-${s}`);
        if (sec) sec.classList.add('hidden');
    });

    // Reset all nav items
    document.querySelectorAll('.nav-item').forEach(link => {
        link.classList.remove(...activeClasses);
        link.classList.add(...inactiveClasses);

        if (link.getAttribute('data-target') === section) {
            link.classList.remove(...inactiveClasses);
            link.classList.add(...activeClasses);
        }
    });

    // Show target
    const targetSec = document.getElementById(`sec-${section}`);
    if (targetSec) targetSec.classList.remove('hidden');

    // Load data based on section
    if (section === 'overview') loadDashboardStats();
    if (section === 'vouchers') loadVouchers();
    if (section === 'planos') loadPlanos();
    if (section === 'users') loadActiveUsers();
    if (section === 'reports') loadReports();
    if (section === 'walledgarden') loadWalledGarden();
    if (section === 'blacklist') loadBlacklist();
    if (section === 'logs') loadLogs();
    if (section === 'settings') loadSettings();
}

// UI State
function showDashboard() {
    loginContainer.classList.add('hidden');
    document.body.classList.remove('justify-center');
    document.body.classList.add('justify-start');
    dashboardLayout.classList.remove('hidden');
    nav('overview');
}

function showLogin() {
    dashboardLayout.classList.add('hidden');
    document.body.classList.remove('justify-start');
    document.body.classList.add('justify-center');
    loginContainer.classList.remove('hidden');
}

function logout() {
    localStorage.removeItem('astrolink_admin_token');
    authToken = null;
    stopIdleTimer();
    showLogin();
}

function showToast(msg, type = 'success') {
    const toast = document.getElementById('toast');
    const icon = document.getElementById('toast-icon');
    document.getElementById('toast-msg').innerText = msg;

    if (type === 'success') {
        icon.className = 'fas fa-check-circle text-cyan-400 text-xl';
        toast.className = 'fixed bottom-5 right-5 transform translate-y-0 opacity-100 transition-all duration-300 bg-slate-900/80 backdrop-blur-md text-white px-6 py-4 rounded-xl shadow-2xl border border-cyan-500/30 shadow-[0_0_15px_rgba(6,182,212,0.2)] flex items-center gap-3 z-50 border-l-4 border-l-cyan-400';
    } else {
        icon.className = 'fas fa-exclamation-circle text-red-400 text-xl glow-red';
        toast.className = 'fixed bottom-5 right-5 transform translate-y-0 opacity-100 transition-all duration-300 bg-slate-900/80 backdrop-blur-md text-white px-6 py-4 rounded-xl shadow-2xl border border-red-500/30 shadow-[0_0_15px_rgba(239,68,68,0.2)] flex items-center gap-3 z-50 border-l-4 border-l-red-500';
    }

    setTimeout(() => {
        toast.classList.remove('translate-y-0', 'opacity-100');
        toast.classList.add('translate-y-20', 'opacity-0');
    }, 3000);
}

// API Helper
async function apiFetch(endpoint, options = {}) {
    const headers = {
        ...(authToken ? { 'Authorization': `Bearer ${authToken}` } : {})
    };

    if (!(options.body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
    }

    try {
        const response = await fetch(`/api/admin${endpoint}`, { ...options, headers });
        const data = await response.json();

        if (response.status === 401) {
            logout();
            throw new Error("Sessão expirada");
        }

        if (!response.ok) {
            throw new Error(data.detail || "Erro na API");
        }

        return data;
    } catch (error) {
        console.error("API Error:", error);
        throw error;
    }
}

// Login Form Submit
loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const btn = loginForm.querySelector('button');
    const originalText = btn.innerHTML;

    try {
        loginError.classList.add('hidden');
        btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Autenticando...';
        btn.disabled = true;

        const payload = {
            username: document.getElementById('username').value,
            password: document.getElementById('password').value
        };

        // Needs to match the Body(...) expectation or JSON post
        // The router expects JSON body for the login payload as defined
        const data = await apiFetch('/login', {
            method: 'POST',
            body: JSON.stringify(payload)
        });

        authToken = data.access_token;
        localStorage.setItem('astrolink_admin_token', authToken);

        document.getElementById('password').value = '';
        showDashboard();
        startIdleTimer();

    } catch (error) {
        loginError.innerText = error.message;
        loginError.classList.remove('hidden');
    } finally {
        btn.innerHTML = originalText;
        btn.disabled = false;
    }
});

// Load Dashboard Stats
async function loadDashboardStats() {
    try {
        const stats = await apiFetch('/dashboard-stats');

        // Format Currency
        const formatter = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });

        document.getElementById('stat-revenue').innerText = formatter.format(stats.faturamento_hoje);
        document.getElementById('stat-online').innerText = stats.online_agora;
        document.getElementById('stat-vouchers').innerText = stats.vouchers_disponiveis;
    } catch (e) {
        console.error("Failed to load stats");
    }
}

// Vouchers Management
let lastGeneratedVouchers = [];

function toggleUniversalFields() {
    const isUniversal = document.getElementById('voucher-is-universal').checked;
    const fields = document.getElementById('universal-fields');
    const qtyInput = document.getElementById('voucher-qtd');
    if (isUniversal) {
        fields.classList.remove('hidden');
        qtyInput.value = 1;
        qtyInput.disabled = true;
    } else {
        fields.classList.add('hidden');
        qtyInput.disabled = false;
    }
}

document.getElementById('generate-voucher-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const btn = document.getElementById('btn-generate');
    const originalText = btn.innerHTML;

    try {
        btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Gerando...';
        btn.disabled = true;

        const isUniversal = document.getElementById('voucher-is-universal') ? document.getElementById('voucher-is-universal').checked : false;

        const payload = {
            duracao_minutos: parseInt(document.getElementById('voucher-duration').value),
            quantidade: isUniversal ? 1 : parseInt(document.getElementById('voucher-qtd').value),
            is_universal: isUniversal,
            max_uses: isUniversal ? parseInt(document.getElementById('voucher-max-uses').value) : 1,
            nome: isUniversal ? document.getElementById('voucher-nome').value : null
        };

        const data = await apiFetch('/gerar-vouchers', {
            method: 'POST',
            body: JSON.stringify(payload)
        });

        showToast(data.mensagem);

        // Update list
        lastGeneratedVouchers = data.vouchers;
        const listEl = document.getElementById('generated-list');
        listEl.innerHTML = data.vouchers.map(v => `<div class="py-1 tracking-widest text-emerald-300 font-bold">${v}</div>`).join('');
        document.getElementById('generated-results').classList.remove('hidden');

        // Refresh table
        loadVouchers();
        // Refresh stats
        loadDashboardStats();

    } catch (e) {
        showToast(e.message, 'error');
    } finally {
        btn.innerHTML = originalText;
        btn.disabled = false;
    }
});

function copyGenerated() {
    const text = lastGeneratedVouchers.join('\n');
    navigator.clipboard.writeText(text).then(() => {
        showToast("Códigos copiados para a área de transferência!");
    });
}

function printGenerated() {
    if (!lastGeneratedVouchers || lastGeneratedVouchers.length === 0) return;

    const iframe = document.createElement('iframe');
    iframe.style.position = 'fixed';
    iframe.style.right = '0';
    iframe.style.bottom = '0';
    iframe.style.width = '0';
    iframe.style.height = '0';
    iframe.style.border = '0';
    document.body.appendChild(iframe);

    const doc = iframe.contentWindow.document;

    const providerName = document.getElementById('cfg-provider-name') && document.getElementById('cfg-provider-name').value ? document.getElementById('cfg-provider-name').value : 'Astrolink Telecom';
    const planSelect = document.getElementById('voucher-duration');
    const planText = planSelect.options[planSelect.selectedIndex].text;

    let cardsHtml = '';

    lastGeneratedVouchers.forEach(code => {
        cardsHtml += `
            <div class="card">
                <div class="header">${providerName}</div>
                <div class="plan-name">Acesso Wi-Fi - ${planText}</div>
                <div class="code-box">PIN: <strong>${code}</strong></div>
                <div class="footer">Conecte-se à rede WiFi e digite o PIN acima.</div>
            </div>
        `;
    });

    const html = `
    <html>
    <head>
        <title>Imprimir Vouchers</title>
        <style>
            @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap');
            body { font-family: 'Inter', sans-serif; background: white; margin: 0; padding: 20px; }
            .grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; }
            .card { 
                border: 2px dashed #cbd5e1; 
                border-radius: 12px; 
                padding: 15px; 
                text-align: center; 
                page-break-inside: avoid;
            }
            .header { font-size: 18px; font-weight: bold; color: #0f172a; margin-bottom: 5px; }
            .plan-name { font-size: 14px; color: #64748b; margin-bottom: 15px; }
            .code-box { 
                background: #f8fafc; 
                border: 1px solid #e2e8f0; 
                padding: 10px; 
                border-radius: 8px; 
                font-family: monospace; 
                font-size: 20px; 
                letter-spacing: 2px;
                color: #0284c7;
                margin-bottom: 15px;
            }
            .footer { font-size: 11px; color: #94a3b8; }
            @media print {
                @page { margin: 1cm; }
            }
        </style>
    </head>
    <body>
        <div class="grid">
            ${cardsHtml}
        </div>
        <script>
            window.onload = function() {
                setTimeout(function() {
                    window.print();
                }, 500);
            }
        </script>
    </body>
    </html>
    `;

    doc.open();
    doc.write(html);
    doc.close();

    setTimeout(() => {
        document.body.removeChild(iframe);
    }, 5000);
}

async function loadVouchers() {
    const tbody = document.getElementById('vouchers-table-body');
    tbody.innerHTML = '<tr><td colspan="4" class="p-4 text-center text-slate-500"><i class="fas fa-spinner fa-spin mr-2"></i> Carregando...</td></tr>';

    try {
        const vouchers = await apiFetch('/vouchers');

        if (vouchers.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" class="p-8 text-center text-slate-500 bg-slate-800/50">Nenhum voucher histórico encontrado.</td></tr>';
            return;
        }

        tbody.innerHTML = vouchers.map(v => {
            let statusBadge = '';
            if (v.status === 'disponivel') {
                statusBadge = '<span class="px-2 py-1 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-md text-xs font-medium">Disponível</span>';
            } else if (v.status === 'esgotado') {
                statusBadge = '<span class="px-2 py-1 bg-orange-500/10 text-orange-400 border border-orange-500/20 rounded-md text-xs font-medium">Esgotado</span>';
            } else {
                statusBadge = '<span class="px-2 py-1 bg-red-500/10 text-red-400 border border-red-500/20 rounded-md text-xs font-medium">Usado</span>';
            }

            let usageText = '';
            let userText = '';

            if (v.is_universal) {
                const icon = '<i class="fas fa-globe text-purple-400 mr-1" title="Universal"></i>';
                usageText = `${v.uses_count} / ${v.max_uses} usos`;
                v.codigo = `${icon} ${v.codigo}`;
                userText = `<br><span class="text-xs text-slate-500">Múltiplos usuários</span>`;
            } else {
                usageText = v.status === 'usado' ? v.usado_em : '-';
                userText = v.usado_por !== '-' ? `<br><span class="text-xs text-slate-500">MAC: ${v.usado_por}</span>` : '';
            }

            return `
                <tr class="hover:bg-slate-700/30 transition-colors">
                    <td class="p-4 font-mono font-bold text-white">${v.codigo}</td>
                    <td class="p-4">${v.duracao}</td>
                    <td class="p-4">${statusBadge}</td>
                    <td class="p-4 text-slate-400">${usageText}${userText}</td>
                </tr>
            `;
        }).join('');

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="4" class="p-4 text-center text-red-400">Erro ao carregar: ${e.message}</td></tr>`;
    }
}

// Planos Management
async function loadPlanos() {
    const tbody = document.getElementById('planos-table-body');
    if (!tbody) return;

    tbody.innerHTML = '<tr><td colspan="7" class="p-8 text-center text-slate-500"><i class="fas fa-spinner fa-spin text-2xl mb-3"></i><br>Buscando planos...</td></tr>';

    try {
        const planos = await apiFetch('/planos');

        if (planos.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="p-8 text-center text-slate-500 bg-slate-800/50">Nenhum plano cadastrado.</td></tr>';
            return;
        }

        tbody.innerHTML = planos.map(p => `
            <tr class="hover:bg-slate-700/30 transition-colors">
                <td class="p-4 font-mono text-slate-500 dark:text-slate-400 text-xs">#${p.id}</td>
                <td class="p-4 font-medium text-slate-900 dark:text-white">${p.nome}</td>
                <td class="p-4 text-emerald-500 font-bold">R$ ${p.preco.toFixed(2)}</td>
                <td class="p-4 text-slate-600 dark:text-slate-300"><i class="far fa-clock text-slate-400 mr-2"></i>${p.duracao_minutos}</td>
                <td class="p-4 text-slate-600 dark:text-slate-300"><i class="fas fa-mobile-alt text-slate-400 mr-2"></i>${p.max_devices}</td>
                <td class="p-4">
                    <button onclick="togglePlanStatus(${p.id}, ${p.ativo})" class="px-3 py-1 text-xs font-bold uppercase rounded-md border ${p.ativo ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20 hover:bg-emerald-500 hover:text-white' : 'bg-slate-500/10 text-slate-500 border-slate-500/20 hover:bg-slate-500 hover:text-white'} transition-all">
                        ${p.ativo ? 'Ativo' : 'Inativo'}
                    </button>
                </td>
                <td class="p-4 text-right">
                    <div class="flex items-center justify-end gap-2">
                        <button onclick='openPlanModal(${JSON.stringify(p).replace(/'/g, "&#39;")})' class="px-2 py-1.5 bg-cyan-500/10 text-cyan-500 hover:bg-cyan-500 hover:text-white rounded transition-colors" title="Editar">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button onclick="deletePlan(${p.id})" class="px-2 py-1.5 bg-red-500/10 text-red-500 hover:bg-red-500 hover:text-white rounded transition-colors" title="Excluir">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="7" class="p-4 text-center text-red-400">Erro ao carregar: ${e.message}</td></tr>`;
    }
}

function openPlanModal(plan = null) {
    const modal = document.getElementById('plan-modal');
    const title = document.getElementById('plan-modal-title');
    const form = document.getElementById('plan-form');

    form.reset();

    if (plan) {
        title.innerHTML = '<i class="fas fa-edit text-cyan-500 text-xl"></i> Editar Plano';
        document.getElementById('plan-id').value = plan.id;
        document.getElementById('plan-nome').value = plan.nome;
        document.getElementById('plan-duracao').value = plan.duracao_minutos;
        document.getElementById('plan-preco').value = plan.preco;
        document.getElementById('plan-devices').value = plan.max_devices;
        document.getElementById('plan-ativo').checked = plan.ativo;
    } else {
        title.innerHTML = '<i class="fas fa-plus-circle text-emerald-500 text-xl"></i> Novo Plano';
        document.getElementById('plan-id').value = '';
    }

    modal.classList.remove('hidden');
}

function closePlanModal() {
    document.getElementById('plan-modal').classList.add('hidden');
}

async function savePlan() {
    const id = document.getElementById('plan-id').value;
    const isEdit = id !== '';

    const payload = {
        nome: document.getElementById('plan-nome').value,
        duracao_minutos: parseInt(document.getElementById('plan-duracao').value),
        max_devices: parseInt(document.getElementById('plan-devices').value),
        preco: parseFloat(document.getElementById('plan-preco').value),
        ativo: document.getElementById('plan-ativo').checked
    };

    if (!payload.nome || isNaN(payload.duracao_minutos) || isNaN(payload.preco)) {
        showToast('Preencha os campos obrigatórios corretamente', 'error');
        return;
    }

    try {
        const method = isEdit ? 'PUT' : 'POST';
        const endpoint = isEdit ? `/planos/${id}` : '/planos';

        await apiFetch(endpoint, {
            method: method,
            body: JSON.stringify(payload)
        });

        showToast(`Plano ${isEdit ? 'atualizado' : 'criado'} com sucesso!`);
        closePlanModal();
        loadPlanos();

    } catch (e) {
        showToast(e.message, 'error');
    }
}

async function deletePlan(id) {
    if (!confirm('Deseja realmente excluir este plano? Se houver transações atreladas, a exclusão pode falhar.')) return;

    try {
        await apiFetch(`/planos/${id}`, { method: 'DELETE' });
        showToast('Plano excluído com sucesso!');
        loadPlanos();
    } catch (e) {
        showToast(e.message, 'error');
    }
}

async function togglePlanStatus(id, currentStatus) {
    try {
        await apiFetch(`/planos/${id}`, {
            method: 'PUT',
            body: JSON.stringify({ ativo: !currentStatus }) // Using partial update assumption or resending old logic? Backend route might need full. We will update the backend to allow PATCH or just full PUT later.
        });
        showToast(`Status alterado com sucesso!`);
        loadPlanos();
    } catch (e) {
        // Fallback or simpler approach for toggle -> we will build a specific patch route
        try {
            await apiFetch(`/planos/${id}/toggle`, { method: 'PATCH' });
            showToast(`Status alterado com sucesso!`);
            loadPlanos();
        } catch (err) {
            showToast(err.message, 'error');
        }
    }
}

let activeUsersData = [];

// Active Users Management
async function loadActiveUsers() {
    const tbody = document.getElementById('users-table-body');
    tbody.innerHTML = '<tr><td colspan="5" class="p-8 text-center text-slate-500"><i class="fas fa-spinner fa-spin text-2xl mb-3"></i><br>Buscando usuários...</td></tr>';

    try {
        const users = await apiFetch('/usuarios-ativos');
        activeUsersData = users; // Store for filtering
        renderActiveUsers();

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="5" class="p-4 text-center text-red-400">Erro ao buscar rede: ${e.message}</td></tr>`;
    }
}

function renderActiveUsers() {
    const tbody = document.getElementById('users-table-body');
    const filterText = document.getElementById('users-filter-input') ? document.getElementById('users-filter-input').value.toLowerCase() : '';

    let filteredUsers = activeUsersData;

    if (filterText) {
        filteredUsers = activeUsersData.filter(u =>
            u.mac_address.toLowerCase().includes(filterText) ||
            (u.ip_atual && u.ip_atual.toLowerCase().includes(filterText)) ||
            (u.nome && u.nome.toLowerCase().includes(filterText))
        );
    }

    if (filteredUsers.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="p-12 text-center text-slate-500 bg-slate-800/50"><div class="text-4xl mb-3">📡</div>Nenhum cliente encontrado.</td></tr>';
        return;
    }

    tbody.innerHTML = filteredUsers.map(u => `
        <tr class="hover:bg-slate-700/30 transition-colors">
            <td class="p-4 font-mono text-cyan-400">${u.mac_address}</td>
            <td class="p-4 text-slate-400 font-mono text-xs">${u.ip_atual || '-'}</td>
            <td class="p-4 font-medium text-white">${u.tempo_restante}</td>
            <td class="p-4 text-slate-500 text-xs">${u.fim_acesso}</td>
            <td class="p-4 text-right">
                <div class="flex items-center justify-end gap-2">
                    <button onclick="openUserDetails(${JSON.stringify(u).replace(/"/g, '&quot;')})" class="px-3 py-1.5 bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500 hover:text-white border border-cyan-500/20 rounded-lg transition-all text-xs font-bold uppercase tracking-wider flex items-center gap-2">
                        <i class="fas fa-eye"></i> Perfil
                    </button>
                    <button onclick="kickUser('${u.mac_address}')" class="px-3 py-1.5 bg-red-500/10 text-red-400 hover:bg-red-500 hover:text-white border border-red-500/20 rounded-lg transition-all text-xs font-bold uppercase tracking-wider flex items-center gap-2">
                        <i class="fas fa-bolt"></i>
                    </button>
                </div>
            </td>
        </tr>
    `).join('');
}

async function kickUser(mac) {
    if (!confirm(`Deseja realmente desconectar e bloquear o acesso do dispositivo com MAC ${mac}?`)) return;

    try {
        const data = await apiFetch('/derrubar-usuario', {
            method: 'POST',
            body: JSON.stringify({ mac_address: mac })
        });

        showToast(data.mensagem);
        loadActiveUsers();
        loadDashboardStats(); // update online counter

    } catch (e) {
        showToast(e.message, 'error');
    }
}

// User Details Modal Logic
async function openUserDetails(userObj) {
    const modal = document.getElementById('user-details-modal');
    const content = document.getElementById('user-details-content');
    const btnDisconnect = document.getElementById('btn-modal-disconnect');

    modal.classList.remove('hidden');
    content.innerHTML = '<div class="text-center text-slate-500 py-8"><i class="fas fa-spinner fa-spin text-3xl mb-4"></i><p>Buscando histórico completo do servidor...</p></div>';

    const formatValue = (val) => val ? val : '<span class="text-slate-400 dark:text-slate-600">N/A</span>';

    try {
        const data = await apiFetch(`/usuario/${encodeURIComponent(userObj.mac_address)}`);

        let sessionHtml = '';
        if (data.sessao_atual) {
            sessionHtml = `
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                <div class="bg-slate-50 dark:bg-slate-950/50 p-4 rounded-xl border border-slate-200/60 dark:border-white/5">
                    <p class="text-xs text-slate-400 mb-1 uppercase tracking-wider">Tempo Restante Estimado</p>
                    <p class="text-md font-medium text-emerald-500 dark:text-emerald-400 block"><i class="far fa-clock mr-1"></i> ${formatValue(data.sessao_atual.tempo_restante)}</p>
                </div>
                <div class="bg-slate-50 dark:bg-slate-950/50 p-4 rounded-xl border border-slate-200/60 dark:border-white/5">
                    <p class="text-xs text-slate-400 mb-1 uppercase tracking-wider">Data/Hora Expiração</p>
                    <p class="text-md font-medium text-slate-700 dark:text-slate-300 block"><i class="far fa-calendar-alt mr-1"></i> ${formatValue(data.sessao_atual.fim_acesso)}</p>
                </div>
            </div>`;
        } else {
            sessionHtml = '<div class="p-4 bg-slate-100 dark:bg-slate-800/50 rounded-xl text-center text-slate-500 mb-6 border border-slate-200 dark:border-white/5">Nenhuma sessão ativa neste momento.</div>';
        }

        let historyHtml = '<div class="text-sm text-slate-500 text-center py-4">Sem histórico recente do sistema.</div>';
        if (data.historico_eventos && data.historico_eventos.length > 0) {
            historyHtml = `
            <div class="overflow-x-auto border border-slate-200/60 dark:border-white/5 rounded-xl">
                <table class="w-full text-left text-sm">
                    <thead class="bg-slate-100 dark:bg-slate-900/80 text-slate-600 dark:text-slate-400">
                        <tr>
                            <th class="p-3 font-medium">Data/Hora</th>
                            <th class="p-3 font-medium">Evento</th>
                            <th class="p-3 font-medium">Descrição</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-200/50 dark:divide-white/5">
                        ${data.historico_eventos.map(h => `
                            <tr class="hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                                <td class="p-3 whitespace-nowrap text-slate-500 dark:text-slate-400">${h.data_hora}</td>
                                <td class="p-3 font-mono font-bold ${h.evento.includes('FAIL') ? 'text-orange-400' : 'text-cyan-500'}">${h.evento}</td>
                                <td class="p-3 text-slate-600 dark:text-slate-300">${h.descricao || '-'}</td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>`;
        }

        content.innerHTML = `
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                <div class="bg-slate-50 dark:bg-slate-950/50 p-4 rounded-xl border border-slate-200/60 dark:border-white/5">
                    <p class="text-xs text-slate-400 mb-1 uppercase tracking-wider">MAC Address</p>
                    <p class="font-mono text-lg font-bold text-cyan-500 dark:text-cyan-400 flex items-center gap-2">
                        ${formatValue(data.mac_address)}
                        <button onclick="navigator.clipboard.writeText('${data.mac_address}'); showToast('MAC copiado!')" class="text-sm text-slate-400 hover:text-cyan-400 transition-colors"><i class="fas fa-copy"></i></button>
                    </p>
                </div>
                <div class="bg-slate-50 dark:bg-slate-950/50 p-4 rounded-xl border border-slate-200/60 dark:border-white/5">
                    <p class="text-xs text-slate-400 mb-1 uppercase tracking-wider">Endereço IP (Sessão Atual)</p>
                    <p class="font-mono text-lg font-bold text-slate-700 dark:text-slate-300">${formatValue(data.sessao_atual?.ip_atual)}</p>
                </div>
            </div>
            
            <h4 class="text-sm font-semibold text-slate-900 dark:text-white uppercase mb-3 border-b border-slate-200/60 dark:border-white/10 pb-2"><i class="fas fa-wifi text-emerald-500 mr-2"></i>Status da Conexão</h4>
            ${sessionHtml}
            
            <h4 class="text-sm font-semibold text-slate-900 dark:text-white uppercase mb-3 border-b border-slate-200/60 dark:border-white/10 pb-2 mt-2"><i class="fas fa-history text-cyan-500 mr-2"></i>Últimas Ações (Logs de Auditoria)</h4>
            ${historyHtml}
        `;

        btnDisconnect.onclick = () => {
            closeUserDetails();
            kickUser(data.mac_address);
        };

        const btnBan = document.getElementById('btn-modal-ban');
        if (btnBan) {
            btnBan.onclick = async () => {
                if (!confirm(`Deseja BANIR permanentemente o MAC ${data.mac_address}?`)) return;
                closeUserDetails();
                try {
                    const payload = { mac_address: data.mac_address, motivo: "Banido via Perfil do Cliente" };
                    const res = await apiFetch('/blacklist', {
                        method: 'POST',
                        body: JSON.stringify(payload)
                    });
                    showToast(res.mensagem);
                    loadActiveUsers();
                    if (!document.getElementById('sec-blacklist').classList.contains('hidden')) {
                        loadBlacklist();
                    }
                } catch (err) {
                    showToast(err.message, 'error');
                }
            };
        }

        const btnQos = document.getElementById('btn-modal-qos');
        if (btnQos) {
            btnQos.onclick = async () => {
                const limit = prompt(`Defina o limite de banda para o MAC ${data.mac_address} (Ex: 10M, 50M, 100M):\\nDeixe em branco para remover o limite.`);
                if (limit === null) return; // cancelled

                try {
                    const payload = { mac_address: data.mac_address, limit: limit.trim().toUpperCase() };
                    const res = await apiFetch('/apply-qos', {
                        method: 'POST',
                        body: JSON.stringify(payload)
                    });
                    showToast(res.mensagem || "QoS Aplicado via SSH (Simulado)");
                } catch (err) {
                    showToast(err.message, 'error');
                }
            };
        }

        if (!data.sessao_atual) {
            btnDisconnect.classList.add('opacity-50', 'cursor-not-allowed');
            btnDisconnect.disabled = true;
            if (btnQos) {
                btnQos.classList.add('opacity-50', 'cursor-not-allowed');
                btnQos.disabled = true;
            }
        } else {
            btnDisconnect.classList.remove('opacity-50', 'cursor-not-allowed');
            btnDisconnect.disabled = false;
            if (btnQos) {
                btnQos.classList.remove('opacity-50', 'cursor-not-allowed');
                btnQos.disabled = false;
            }
        }

    } catch (e) {
        content.innerHTML = `<div class="p-4 text-center text-red-500"><i class="fas fa-exclamation-triangle text-3xl mb-2"></i><br>Erro ao carregar dados do usuário: ${e.message}</div>`;
    }
}

function closeUserDetails() {
    const modal = document.getElementById('user-details-modal');
    modal.classList.add('hidden');
}

// -------------------------------------------------------------
// ADVANCED FEATURES
// -------------------------------------------------------------

// ==== Relatórios Financeiros ====
let revenueChartInstance = null;
async function loadReports() {
    try {
        const data = await apiFetch('/relatorios/financeiro');

        const formatter = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });
        document.getElementById('report-total-revenue').innerText = formatter.format(data.total);
        document.getElementById('report-pix-revenue').innerText = formatter.format(data.receita_pix);
        document.getElementById('report-voucher-revenue').innerText = formatter.format(data.receita_voucher);

        // Render Chart
        const ctx = document.getElementById('revenueChart').getContext('2d');

        if (revenueChartInstance) {
            revenueChartInstance.destroy();
        }

        revenueChartInstance = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['PIX', 'Dinheiro (Vouchers)'],
                datasets: [{
                    data: [data.receita_pix, data.receita_voucher],
                    backgroundColor: ['#34d399', '#22d3ee'],
                    borderWidth: 0,
                    hoverOffset: 4
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: { color: '#cbd5e1' }
                    }
                }
            }
        });

        // Add Mock Heatmap Logic
        renderHeatmap();

    } catch (e) {
        console.error("Failed to load reports:", e);
    }
}

function renderHeatmap() {
    const container = document.getElementById('heatmap-container');
    if (!container) return;

    let html = '';
    const days = ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb'];

    // Y-Axis Labels (Days)
    let daysHtml = `<div class="flex flex-col gap-1 w-8 pr-2 pt-[18px]">`;
    for (let d = 0; d < 7; d++) {
        daysHtml += `<div class="h-6 text-[10px] text-slate-400 flex items-center justify-end font-medium">${days[d]}</div>`;
    }
    daysHtml += `</div>`;

    // Data Columns (Hours)
    for (let h = 0; h < 24; h++) {
        html += `<div class="flex flex-col gap-1 w-6">`;
        // X-Axis Header (Hour)
        html += `<div class="text-[9px] text-slate-500 text-center mb-1">${h}h</div>`;

        for (let d = 0; d < 7; d++) {
            // Generate some logical mock data. higher during day/evening, lower at night
            let baseProb = (h > 9 && h < 23) ? 0.6 : 0.1;
            if (d === 0 || d === 6) baseProb += 0.2; // weekends are busier

            const intensity = Math.random() * baseProb;
            let colorClass = 'bg-slate-200/50 dark:bg-slate-800/50';

            if (intensity > 0.7) colorClass = 'bg-orange-600 shadow-[0_0_8px_rgba(234,88,12,0.6)] z-10';
            else if (intensity > 0.5) colorClass = 'bg-orange-500';
            else if (intensity > 0.3) colorClass = 'bg-orange-400';
            else if (intensity > 0.1) colorClass = 'bg-orange-300/60';

            html += `<div class="w-full h-6 rounded-sm ${colorClass} hover:ring-2 hover:ring-white transition-all cursor-pointer relative" title="${days[d]} às ${h}h (Pico estimado: ${Math.round(intensity * 100)}%)"></div>`;
        }
        html += `</div>`;
    }

    container.innerHTML = daysHtml + html;
}

async function exportCSV() {
    try {
        const response = await fetch(`/api/admin/relatorios/exportar-csv`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });

        if (!response.ok) {
            if (response.status === 401) logout();
            throw new Error("Erro ao baixar CSV");
        }

        // Handle download
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        // filename is provided by the server in headers, but fallback here
        a.download = `relatorio_${new Date().getTime()}.csv`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);

        showToast("Download iniciado!");
    } catch (e) {
        showToast(e.message, 'error');
    }
}

// ==== Walled Garden ====
const wgForm = document.getElementById('add-wg-rule-form');
if (wgForm) {
    wgForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        try {
            const payload = {
                tipo: document.getElementById('wg-tipo').value,
                valor: document.getElementById('wg-valor').value,
                descricao: document.getElementById('wg-descricao').value
            };

            const data = await apiFetch('/walled-garden', {
                method: 'POST',
                body: JSON.stringify(payload)
            });

            showToast(data.mensagem);

            // reset form
            document.getElementById('wg-valor').value = '';
            document.getElementById('wg-descricao').value = '';

            loadWalledGarden();
        } catch (err) {
            showToast(err.message, 'error');
        }
    });
}

async function loadWalledGarden() {
    const tbody = document.getElementById('wg-table-body');
    tbody.innerHTML = '<tr><td colspan="4" class="p-4 text-center text-slate-500"><i class="fas fa-spinner fa-spin mr-2"></i> Carregando...</td></tr>';

    try {
        const rules = await apiFetch('/walled-garden');

        if (rules.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" class="p-8 text-center text-slate-500 bg-slate-800/50">Nenhuma regra Walled Garden cadastrada.</td></tr>';
            return;
        }

        tbody.innerHTML = rules.map(r => `
            <tr class="hover:bg-slate-700/30 transition-colors">
                <td class="p-4 font-bold text-cyan-400 tracking-wider">${r.tipo}</td>
                <td class="p-4 font-mono text-white">${r.valor}</td>
                <td class="p-4 text-slate-400">${r.descricao || '-'}</td>
                <td class="p-4 text-right">
                    <button onclick="deleteWalledGarden(${r.id})" class="text-red-400 hover:text-red-300 transition-colors" title="Deletar">
                        <i class="fas fa-trash-alt"></i>
                    </button>
                </td>
            </tr>
        `).join('');

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="4" class="p-4 text-center text-red-400">Erro: ${e.message}</td></tr>`;
    }
}

async function deleteWalledGarden(id) {
    if (!confirm("Tem certeza que deseja apagar esta regra Walled Garden?")) return;
    try {
        const data = await apiFetch(`/walled-garden/${id}`, { method: 'DELETE' });
        showToast(data.mensagem);
        loadWalledGarden();
    } catch (e) {
        showToast(e.message, 'error');
    }
}

// ==== Blacklist ====
const blForm = document.getElementById('add-blacklist-form');
if (blForm) {
    blForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        try {
            const payload = {
                mac_address: document.getElementById('bl-mac').value,
                motivo: document.getElementById('bl-motivo').value
            };

            const data = await apiFetch('/blacklist', {
                method: 'POST',
                body: JSON.stringify(payload)
            });

            showToast(data.mensagem);

            // reset form
            document.getElementById('bl-mac').value = '';
            document.getElementById('bl-motivo').value = '';

            loadBlacklist();
        } catch (err) {
            showToast(err.message, 'error');
        }
    });
}

async function loadBlacklist() {
    const tbody = document.getElementById('bl-table-body');
    tbody.innerHTML = '<tr><td colspan="4" class="p-4 text-center text-slate-500"><i class="fas fa-spinner fa-spin mr-2"></i> Carregando...</td></tr>';

    try {
        const records = await apiFetch('/blacklist');

        if (records.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" class="p-8 text-center text-slate-500 bg-slate-800/50">Nenhum MAC bloqueado.</td></tr>';
            return;
        }

        tbody.innerHTML = records.map(r => `
            <tr class="hover:bg-slate-700/30 transition-colors">
                <td class="p-4 font-mono font-bold text-red-400 tracking-wider uppercase">${r.mac_address}</td>
                <td class="p-4 text-slate-300 font-mono text-xs">${r.data_bloqueio}</td>
                <td class="p-4 text-slate-400">${r.motivo || '-'}</td>
                <td class="p-4 text-right">
                    <button onclick="deleteBlacklist('${r.mac_address}')" class="text-cyan-400 hover:text-cyan-300 transition-colors" title="Desbloquear">
                        <i class="fas fa-unlock"></i> Revogar
                    </button>
                </td>
            </tr>
        `).join('');

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="4" class="p-4 text-center text-red-400">Erro: ${e.message}</td></tr>`;
    }
}

async function deleteBlacklist(mac) {
    if (!confirm(`Deseja realmente remover o bloqueio permanente do MAC ${mac}?`)) return;
    try {
        const data = await apiFetch(`/blacklist/${mac}`, { method: 'DELETE' });
        showToast(data.mensagem);
        loadBlacklist();
    } catch (e) {
        showToast(e.message, 'error');
    }
}

// ==== Contas Fixas (Mensalistas) ====
const cfForm = document.getElementById('add-conta-fixa-form');
if (cfForm) {
    cfForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        try {
            const payload = {
                mac_address: document.getElementById('cf-mac').value,
                nome: document.getElementById('cf-nome').value,
                telefone: document.getElementById('cf-telefone').value,
                email: document.getElementById('cf-email').value,
                observacoes: '',
                ativo: document.getElementById('cf-ativo').value === 'true'
            };

            const dataRenov = document.getElementById('cf-renovacao').value;
            if (dataRenov) {
                payload.data_renovacao = dataRenov;
            }

            const data = await apiFetch('/contas-fixas', {
                method: 'POST',
                body: JSON.stringify(payload)
            });

            showToast(data.mensagem);

            // reset form minimal fields
            document.getElementById('cf-mac').value = '';
            document.getElementById('cf-nome').value = '';
            document.getElementById('cf-telefone').value = '';
            document.getElementById('cf-email').value = '';

            loadContasFixas();
        } catch (err) {
            showToast(err.message, 'error');
        }
    });
}

async function loadContasFixas() {
    const tbody = document.getElementById('cf-table-body');
    if (!tbody) return;
    tbody.innerHTML = '<tr><td colspan="5" class="p-4 text-center text-slate-500"><i class="fas fa-spinner fa-spin mr-2"></i> Carregando...</td></tr>';

    try {
        const records = await apiFetch('/contas-fixas');

        if (records.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" class="p-8 text-center text-slate-500 bg-slate-800/50">Nenhuma conta fixa cadastrada.</td></tr>';
            return;
        }

        tbody.innerHTML = records.map(r => {
            const statusBadge = r.ativo
                ? '<span class="px-2 py-1 bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 rounded-md text-xs font-bold">Ativo</span>'
                : '<span class="px-2 py-1 bg-red-500/10 text-red-500 border border-red-500/20 rounded-md text-xs font-bold">Bloqueado</span>';

            const renovacaoStr = r.data_renovacao ? `<i class="far fa-calendar-alt text-slate-400"></i> ${r.data_renovacao}` : '-';

            return `
            <tr class="hover:bg-slate-700/30 transition-colors">
                <td class="p-4">
                    <button onclick="editContaFixa('${r.mac_address}', '${r.nome}', '${r.telefone || ''}', '${r.email || ''}', '${r.data_renovacao || ''}', ${r.ativo})" class="text-indigo-400 hover:text-indigo-300 transition-colors bg-indigo-500/10 px-3 py-1.5 rounded-lg border border-indigo-500/20" title="Editar">
                        <i class="fas fa-pen text-xs"></i> Editar
                    </button>
                </td>
                <td class="p-4 font-medium text-slate-900 dark:text-slate-200">
                    <div>${r.nome}</div>
                    ${r.telefone ? `<div class="text-xs text-slate-400"><i class="fas fa-phone-alt"></i> ${r.telefone}</div>` : ''}
                </td>
                <td class="p-4 font-mono font-bold text-cyan-400 tracking-wider text-xs">${r.mac_address}</td>
                <td class="p-4 text-slate-300 text-sm whitespace-nowrap">${renovacaoStr}</td>
                <td class="p-4 text-center">${statusBadge}</td>
                <td class="p-4 text-right">
                    <button onclick="deleteContaFixa('${r.mac_address}')" class="text-red-400 hover:text-red-300 transition-colors" title="Excluir">
                        <i class="fas fa-trash"></i>
                    </button>
                </td>
            </tr>
            `;
        }).join('');

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="5" class="p-4 text-center text-red-400">Erro: ${e.message}</td></tr>`;
    }
}

function editContaFixa(mac, nome, tempTelefone, tempEmail, dtRenovacao, isAtivo) {
    document.getElementById('cf-mac').value = mac;
    document.getElementById('cf-nome').value = nome;
    document.getElementById('cf-telefone').value = tempTelefone !== 'undefined' ? tempTelefone : '';
    document.getElementById('cf-email').value = tempEmail !== 'undefined' ? tempEmail : '';
    document.getElementById('cf-renovacao').value = dtRenovacao !== 'undefined' ? dtRenovacao : '';
    document.getElementById('cf-ativo').value = isAtivo ? 'true' : 'false';
    window.scrollTo({ top: document.getElementById('sec-contas-fixas').offsetTop - 20, behavior: 'smooth' });
}

async function deleteContaFixa(mac) {
    if (!confirm(`Deseja realmente apagar a conta fixa do MAC ${mac}? O usuário retornará a ser um cliente comum.`)) return;
    try {
        const data = await apiFetch(`/contas-fixas/${mac}`, { method: 'DELETE' });
        showToast(data.mensagem);
        loadContasFixas();
    } catch (e) {
        showToast(e.message, 'error');
    }
}

// ==== Logs ====
async function loadLogs() {
    const tbody = document.getElementById('logs-table-body');
    if (!tbody) return;
    tbody.innerHTML = '<tr><td colspan="4" class="p-4 text-center text-slate-500"><i class="fas fa-spinner fa-spin mr-2"></i> Buscando auditoria...</td></tr>';

    try {
        let queryParams = new URLSearchParams({ limit: 100 });

        const startDate = document.getElementById('log-start-date')?.value;
        const endDate = document.getElementById('log-end-date')?.value;
        const eventType = document.getElementById('log-event-type')?.value;

        if (startDate) queryParams.append('start_date', startDate);
        if (endDate) queryParams.append('end_date', endDate);
        if (eventType && eventType !== 'ALL') queryParams.append('event_type', eventType);

        const logs = await apiFetch(`/logs?${queryParams.toString()}`);

        if (logs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" class="p-8 text-center text-slate-500 bg-slate-800/50">Nenhum evento registrado ainda.</td></tr>';
            return;
        }

        tbody.innerHTML = logs.map(l => {
            let color = 'text-slate-400';
            if (l.event_type.includes('SUCCESS')) color = 'text-emerald-400';
            if (l.event_type.includes('FAIL')) color = 'text-red-400';
            if (l.event_type === 'PIX_CREATED') color = 'text-cyan-400';

            const dateStr = new Date(l.created_at).toLocaleString('pt-BR');

            let macColumn = `<td class="p-4 text-cyan-300">-</td>`;
            if (l.mac_address) {
                const macEscaped = l.mac_address.replace(/'/g, "\\'");
                // We create a mocked userObj for the modal to at least show the MAC 
                // since the logs table doesn't have the full session data
                const userObjStr = JSON.stringify({ mac_address: l.mac_address }).replace(/"/g, '&quot;');
                macColumn = `
                <td class="p-4">
                    <div class="flex items-center gap-2">
                        <span class="text-cyan-300 font-mono">${l.mac_address}</span>
                        <button onclick="navigator.clipboard.writeText('${macEscaped}'); showToast('MAC copiado!')" class="text-slate-500 hover:text-cyan-400 transition-colors" title="Copiar MAC">
                            <i class="fas fa-copy"></i>
                        </button>
                        <button onclick="openUserDetails(${userObjStr})" class="text-slate-500 hover:text-cyan-400 transition-colors" title="Ver Cliente">
                            <i class="fas fa-eye"></i>
                        </button>
                    </div>
                </td>`;
            }

            return `
            <tr class="hover:bg-slate-700/30 transition-colors border-b border-slate-700/50">
                <td class="p-4 text-xs tracking-wider whitespace-nowrap">${dateStr}</td>
                <td class="p-4 font-bold ${color}">${l.event_type}</td>
                ${macColumn}
                <td class="p-4 text-slate-300">${l.description || '-'}</td>
            </tr>
        `}).join('');

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="4" class="p-4 text-center text-red-400">Erro: ${e.message}</td></tr>`;
    }
}

// ==== Configurações Globais ====
async function loadSettings() {
    try {
        const settings = await apiFetch('/system-settings');
        if (settings['PROVIDER_NAME']) document.getElementById('cfg-provider-name').value = settings['PROVIDER_NAME'];
        if (settings['LOGO_URL']) document.getElementById('cfg-logo-url').value = settings['LOGO_URL'];

        if (settings['PRIMARY_COLOR']) {
            document.getElementById('cfg-primary-color').value = settings['PRIMARY_COLOR'];
            document.getElementById('cfg-primary-color-picker').value = settings['PRIMARY_COLOR'];
        }

        if (settings['THEME_MODE']) document.getElementById('cfg-theme-mode').value = settings['THEME_MODE'];
        if (settings['SECONDARY_COLOR']) {
            document.getElementById('cfg-secondary-color').value = settings['SECONDARY_COLOR'];
            document.getElementById('cfg-secondary-color-picker').value = settings['SECONDARY_COLOR'];
        }
        if (settings['WELCOME_MSG']) document.getElementById('cfg-welcome-msg').value = settings['WELCOME_MSG'];
        if (settings['FOOTER_TEXT']) document.getElementById('cfg-footer-text').value = settings['FOOTER_TEXT'];
        if (settings['TERMS_OF_USE']) document.getElementById('cfg-terms-of-use').value = settings['TERMS_OF_USE'];
        if (settings['REDIRECT_URL']) document.getElementById('cfg-redirect-url').value = settings['REDIRECT_URL'];
        if (settings['DEFAULT_LANGUAGE']) document.getElementById('cfg-default-language').value = settings['DEFAULT_LANGUAGE'];

        // Gateways and Integrations
        if (settings['MERCADOPAGO_ACCESS_TOKEN']) document.getElementById('cfg-mp-token').value = settings['MERCADOPAGO_ACCESS_TOKEN'];

        if (settings['ROUTER_IP']) document.getElementById('cfg-router-ip').value = settings['ROUTER_IP'];
        if (settings['ROUTER_USER']) document.getElementById('cfg-router-user').value = settings['ROUTER_USER'];
        if (settings['ROUTER_PASS']) document.getElementById('cfg-router-pass').value = settings['ROUTER_PASS'];

        document.getElementById('cfg-login-email').checked = settings['LOGIN_EMAIL_ENABLED'] === 'true' || settings['LOGIN_EMAIL_ENABLED'] === undefined;
        document.getElementById('cfg-login-purchase').checked = settings['LOGIN_PURCHASE_ENABLED'] === 'true' || settings['LOGIN_PURCHASE_ENABLED'] === undefined;
        document.getElementById('cfg-login-voucher').checked = settings['LOGIN_VOUCHER_ENABLED'] === 'true' || settings['LOGIN_VOUCHER_ENABLED'] === undefined;
    } catch (e) {
        console.error("Falha ao carregar configurações globais:", e);
    }
}

const settingsForm = document.getElementById('settings-form');
if (settingsForm) {
    settingsForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const btn = document.getElementById('btn-save-settings');
        const originalText = btn.innerHTML;

        try {
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Salvando...';
            btn.disabled = true;

            const formData = new FormData();
            formData.append('PROVIDER_NAME', document.getElementById('cfg-provider-name').value);
            formData.append('LOGO_URL', document.getElementById('cfg-logo-url').value);
            formData.append('PRIMARY_COLOR', document.getElementById('cfg-primary-color').value);
            formData.append('THEME_MODE', document.getElementById('cfg-theme-mode').value);
            formData.append('SECONDARY_COLOR', document.getElementById('cfg-secondary-color').value);
            formData.append('WELCOME_MSG', document.getElementById('cfg-welcome-msg').value);
            formData.append('FOOTER_TEXT', document.getElementById('cfg-footer-text').value);
            formData.append('TERMS_OF_USE', document.getElementById('cfg-terms-of-use').value);
            formData.append('REDIRECT_URL', document.getElementById('cfg-redirect-url').value);
            formData.append('DEFAULT_LANGUAGE', document.getElementById('cfg-default-language').value);

            formData.append('MERCADOPAGO_ACCESS_TOKEN', document.getElementById('cfg-mp-token').value);
            formData.append('ROUTER_IP', document.getElementById('cfg-router-ip').value);
            formData.append('ROUTER_USER', document.getElementById('cfg-router-user').value);
            formData.append('ROUTER_PASS', document.getElementById('cfg-router-pass').value);

            formData.append('LOGIN_EMAIL_ENABLED', document.getElementById('cfg-login-email').checked ? 'true' : 'false');
            formData.append('LOGIN_PURCHASE_ENABLED', document.getElementById('cfg-login-purchase').checked ? 'true' : 'false');
            formData.append('LOGIN_VOUCHER_ENABLED', document.getElementById('cfg-login-voucher').checked ? 'true' : 'false');

            const logoFileInput = document.getElementById('cfg-logo-file');
            if (logoFileInput && logoFileInput.files.length > 0) {
                formData.append('file', logoFileInput.files[0]);
            }

            const data = await apiFetch('/system-settings', {
                method: 'POST',
                body: formData
            });

            if (data.logo_url) {
                document.getElementById('cfg-logo-url').value = data.logo_url;
                document.getElementById('cfg-logo-file').value = ''; // reseta input de arquivo
            }

            showToast(data.mensagem);
        } catch (err) {
            showToast(err.message, 'error');
        } finally {
            btn.innerHTML = originalText;
            btn.disabled = false;
        }
    });

    // Sync color picker with text input
    document.getElementById('cfg-primary-color-picker').addEventListener('input', (e) => {
        document.getElementById('cfg-primary-color').value = e.target.value;
    });
    document.getElementById('cfg-primary-color').addEventListener('input', (e) => {
        document.getElementById('cfg-primary-color-picker').value = e.target.value;
    });

    document.getElementById('cfg-secondary-color-picker').addEventListener('input', (e) => {
        document.getElementById('cfg-secondary-color').value = e.target.value;
    });
    document.getElementById('cfg-secondary-color').addEventListener('input', (e) => {
        document.getElementById('cfg-secondary-color-picker').value = e.target.value;
    });

    // Logo Upload Logic
    const logoFileInput = document.getElementById('cfg-logo-file');
    const logoUrlInput = document.getElementById('cfg-logo-url');
    const btnUploadLogo = document.getElementById('btn-upload-logo');

    if (logoFileInput && logoUrlInput) {
        logoFileInput.addEventListener('change', (e) => {
            if (!e.target.files.length) return;
            const file = e.target.files[0];
            logoUrlInput.value = file.name; // Apenas mostra o nome provisório
            showToast('Arquivo selecionado. Clique em Salvar para enviar e aplicar.', 'success');
        });
    }
}

// ==== Notificações em Tempo Real ====
let pollInterval;

function toggleNotifications() {
    const dropdown = document.getElementById('notif-dropdown');
    dropdown.classList.toggle('hidden');
    // Hide badge initially when opened
    document.getElementById('notif-badge').classList.add('hidden');
}

function clearNotifications() {
    document.getElementById('notif-list').innerHTML = '<div class="p-4 text-center text-sm text-slate-500">Nenhum alerta recente.</div>';
}

async function pollNotifications() {
    if (!authToken) return;
    try {
        const notifs = await apiFetch('/notifications');
        const list = document.getElementById('notif-list');
        const badge = document.getElementById('notif-badge');

        if (notifs && notifs.length > 0) {
            badge.classList.remove('hidden');
            let html = '';
            notifs.forEach(n => {
                let color = 'text-red-400';
                if (n.event_type.includes('FAIL')) color = 'text-orange-400';
                const dateStr = new Date(n.created_at).toLocaleTimeString('pt-BR');

                html += `
                <div class="px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                    <p class="text-xs text-slate-400 font-mono mb-1">${dateStr} | <span class="${color} font-bold tracking-wider">${n.event_type}</span></p>
                    <p class="text-sm text-slate-700 dark:text-slate-300 leading-tight">${n.description || 'Alerta do sistema'}</p>
                </div>
                `;
            });
            list.innerHTML = html;
        }
    } catch (e) {
        // block silent error on polling
    }
}

// Start polling every minute if authenticated
if (authToken) {
    pollNotifications();
    pollInterval = setInterval(pollNotifications, 60000);
}

// Attach filter listeners
document.addEventListener('DOMContentLoaded', () => {
    const filterInput = document.getElementById('users-filter-input');
    if (filterInput) {
        filterInput.addEventListener('input', () => {
            renderActiveUsers();
        });
    }

    // Initialize Sortable for Overview Dashboard Grid
    const dashboardGrid = document.getElementById('dashboard-grid');
    if (dashboardGrid && typeof Sortable !== 'undefined') {
        new Sortable(dashboardGrid, {
            animation: 150,
            ghostClass: 'opacity-40',
            handle: '.cursor-move',
            store: {
                get: function (sortable) {
                    var order = localStorage.getItem('astrolink-dashboard-order');
                    return order ? order.split('|') : [];
                },
                set: function (sortable) {
                    var order = sortable.toArray();
                    localStorage.setItem('astrolink-dashboard-order', order.join('|'));
                }
            }
        });
    }
});

// ==== Backend / Diagnostics ====
async function downloadBackup() {
    try {
        const response = await fetch('/api/admin/system/backup', {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });
        if (!response.ok) {
            if (response.status === 401) logout();
            throw new Error('Erro ao gerar backup');
        }

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;

        // Extract filename from Content-Disposition header if present
        let filename = 'astrolink_backup.db';
        const disposition = response.headers.get('content-disposition');
        if (disposition && disposition.indexOf('attachment') !== -1) {
            const filenameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/;
            const matches = filenameRegex.exec(disposition);
            if (matches != null && matches[1]) {
                filename = matches[1].replace(/['"]/g, '');
            }
        }

        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        showToast('Backup baixado com sucesso!');
    } catch (e) {
        showToast(e.message, 'error');
    }
}

async function restoreBackup() {
    const input = document.getElementById('restore-file-input');
    if (!input.files || input.files.length === 0) {
        showToast('Selecione um arquivo .db', 'error');
        return;
    }

    if (!confirm("Isso irá sobrescrever TODO o banco de dados. Essa operação é irreversível.\nDeseja continuar?")) return;

    const formData = new FormData();
    formData.append('file', input.files[0]);

    try {
        const response = await fetch('/api/admin/system/restore', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${authToken}` },
            body: formData
        });

        const data = await response.json();
        if (!response.ok) throw new Error(data.detail || 'Erro ao restaurar banco.');

        showToast(data.mensagem);
        setTimeout(() => window.location.reload(), 3000); // Reload after 3s to reflect changes
    } catch (e) {
        showToast(e.message, 'error');
    }
}

async function runPing() {
    const target = document.getElementById('ping-target-input').value;
    const output = document.getElementById('ping-terminal-output');
    const btn = document.getElementById('btn-run-ping');

    if (!target) {
        showToast('Informe o IP ou Host', 'error');
        return;
    }

    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Ping';
    output.innerText = `Pinging ${target}...\n`;

    try {
        const data = await apiFetch('/system/ping', {
            method: 'POST',
            body: JSON.stringify({ target })
        });

        output.innerText += data.resultado;
    } catch (e) {
        output.innerText += `\nErro: ${e.message}`;
    } finally {
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-paper-plane"></i> Ping';
    }
}
