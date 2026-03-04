/**
 * Main Application Logic for Astrolink Captive Portal
 */

document.addEventListener('DOMContentLoaded', () => {
    // --- Application State ---
    const state = {
        userMac: null,
        gatewayToken: null,
        currentTxid: null,
        pollingInterval: null,
        countdownInterval: null,
        timeLeft: 600, // 10 minutes in seconds for Walled Garden
        redirectUrl: null
    };

    // --- DOM Elements ---
    const screens = {
        plans: document.getElementById('screen-plans'),
        payment: document.getElementById('screen-payment'),
        success: document.getElementById('screen-success'),
        userData: document.getElementById('screen-user-data')
    };

    const ui = {
        planList: document.getElementById('plan-list'),
        countdown: document.getElementById('countdown'),
        qrImage: document.getElementById('qr-image'),
        pixCodeInput: document.getElementById('pix-code'),
        btnCopy: document.getElementById('btn-copy'),
        btnCopyMain: document.getElementById('btn-copy-main'),
        btnCancel: document.getElementById('btn-cancel'),
        btnStart: document.getElementById('btn-start'),
        successTime: document.getElementById('success-time'),
        successExpire: document.getElementById('success-expire'),
        successIcon: document.getElementById('success-icon'),
        // Voucher Elements
        inputVoucher: document.getElementById('inputVoucher'),
        btnVoucher: document.getElementById('btn-voucher'),
        voucherError: document.getElementById('voucher-error'),

        // PIX User Data form
        formUserData: document.getElementById('form-user-data'),
        inputNome: document.getElementById('input-nome'),
        inputSobrenome: document.getElementById('input-sobrenome'),
        inputCpf: document.getElementById('input-cpf'),
        btnCancelUserData: document.getElementById('btn-cancel-user-data')
    };

    // --- Initialization ---
    async function init() {
        extractUrlParams();
        await loadSettings();
        setupEventListeners();
        loadPlans();

        // For testing/mocking purposes: Add a hidden trigger to force payment success
        addDebugTrigger();
    }

    // --- Core Functions ---

    /**
     * Extract MAC and Token from URL
     */
    function extractUrlParams() {
        // Exemplo: http://10.0.0.10:8000/?mac=00:1A:2B:3C:4D:5E&tok=xyz
        const urlParams = new URLSearchParams(window.location.search);
        state.userMac = urlParams.get('mac') || '00:00:00:00:00:00'; // Default just for preview
        state.gatewayToken = urlParams.get('tok') || 'demo_token';
        console.log(`Loaded portal for MAC: ${state.userMac}, Token: ${state.gatewayToken}`);
    }

    /**
     * Load Dynamic Settings for White-Labeling
     */
    async function loadSettings() {
        try {
            const settings = await API.getSettings();

            if (settings['LOGO_URL']) {
                const logoImgs = document.querySelectorAll('img[alt="Astrolink Logo"]');
                logoImgs.forEach(img => img.src = settings['LOGO_URL']);
            }

            if (settings['PROVIDER_NAME']) {
                document.title = settings['PROVIDER_NAME'] + ' - Wi-Fi Access';
            }

            if (settings['WELCOME_MSG']) {
                const title = document.querySelector('h1');
                if (title) title.textContent = settings['WELCOME_MSG'];
            } else if (settings['PROVIDER_NAME']) {
                const title = document.querySelector('h1');
                if (title) title.textContent = settings['PROVIDER_NAME'] + ' Wi-Fi';
            }

            if (settings['FOOTER_TEXT']) {
                const footerText = document.getElementById('footer-text');
                if (footerText) footerText.innerHTML = settings['FOOTER_TEXT'];
            }

            if (settings['TERMS_OF_USE']) {
                const termsModalBody = document.getElementById('terms-modal-body');
                const termsTriggerContainer = document.getElementById('terms-trigger-container');
                if (termsModalBody) {
                    termsModalBody.innerHTML = settings['TERMS_OF_USE'];
                }
                if (termsTriggerContainer) {
                    termsTriggerContainer.classList.remove('hidden');
                }
            }

            if (settings['REDIRECT_URL']) {
                state.redirectUrl = settings['REDIRECT_URL'];
            }

            if (settings['DEFAULT_LANGUAGE']) {
                document.documentElement.lang = settings['DEFAULT_LANGUAGE'];
            }

            if (settings['THEME_MODE']) {
                const htmlElem = document.documentElement;
                if (settings['THEME_MODE'] === 'dark') {
                    htmlElem.classList.add('dark');
                } else if (settings['THEME_MODE'] === 'light') {
                    htmlElem.classList.remove('dark');
                } else {
                    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
                        htmlElem.classList.add('dark');
                    } else {
                        htmlElem.classList.remove('dark');
                    }
                }
            }

            if (settings['LOGIN_VOUCHER_ENABLED'] === 'false') {
                const voucherSection = document.getElementById('voucher-section');
                if (voucherSection) voucherSection.style.display = 'none';

                const divider = document.querySelector('.my-8');
                if (divider) divider.style.display = 'none';
            }

            if (settings['LOGIN_PURCHASE_ENABLED'] === 'false') {
                const planListInfo = document.getElementById('plan-list');
                const pDesc = document.querySelector('p.text-slate-400.mt-2.text-sm');
                if (planListInfo) planListInfo.style.display = 'none';
                if (pDesc) pDesc.style.display = 'none';

                const divider = document.querySelector('.my-8');
                if (divider) divider.style.display = 'none';
            }

            if (settings['BACKGROUND_URL']) {
                document.body.style.backgroundImage = `url(${settings['BACKGROUND_URL']})`;
                document.body.style.backgroundSize = 'cover';
                document.body.style.backgroundPosition = 'center';
                document.body.style.backgroundAttachment = 'fixed';
                // Dim the glow effects out slightly since there is a background image now
                const glowEffects = document.querySelector('.fixed.inset-0.z-0');
                if (glowEffects) glowEffects.style.opacity = '0.3';
            }

            if (settings['PRIMARY_COLOR']) {
                const hex = settings['PRIMARY_COLOR'];
                const style = document.createElement('style');
                style.innerHTML = `
                    .bg-brand-500 { background-color: ${hex} !important; }
                    .bg-brand-600 { background-color: ${hex} !important; filter: brightness(0.85); }
                    .text-brand-500, .text-brand-400 { color: ${hex} !important; }
                    .border-brand-500 { border-color: ${hex} !important; }
                    .ring-brand-500 { --tw-ring-color: ${hex} !important; }
                    .from-brand-500 { --tw-gradient-from: ${hex} var(--tw-gradient-from-position) !important; }
                    .to-brand-500 { --tw-gradient-to: ${hex} var(--tw-gradient-to-position) !important; }
                    .via-brand-500 { --tw-gradient-stops: var(--tw-gradient-from), ${hex} var(--tw-gradient-via-position), var(--tw-gradient-to) !important; }
                    .shadow-\\[0_0_15px_rgba\\(0\\,229\\,255\\,0\\.3\\)\\] { box-shadow: 0 0 15px ${hex}40 !important; }
                    .hover\\:shadow-\\[0_0_25px_rgba\\(0\\,229\\,255\\,0\\.5\\)\\]:hover { box-shadow: 0 0 25px ${hex}80 !important; }
                    .drop-shadow-\\[0_0_8px_rgba\\(0\\,229\\,255\\,0\\.3\\)\\] { filter: drop-shadow(0 0 8px ${hex}40) !important; }
                `;
                document.head.appendChild(style);
            }
        } catch (e) {
            console.error("Erro ao carregar configurações do portal:", e);
        }
    }

    /**
     * Set up all button listeners
     */
    function setupEventListeners() {
        ui.btnCopy.addEventListener('click', copyPixCode);
        ui.btnCopyMain.addEventListener('click', copyPixCode);

        ui.btnCancel.addEventListener('click', () => {
            stopPolling();
            stopCountdown();
            switchScreen('plans');
        });

        ui.btnStart.addEventListener('click', () => {
            // Re-direct to OpenNDS Auth to finalize connection
            // e.g. window.location.href = `http://10.0.0.1/opennds_auth?tok=${state.gatewayToken}`;
            const targetUrl = `http://10.0.0.1/opennds_auth?tok=${state.gatewayToken}`;
            console.log(`[OpenNDS] Granting access: ${targetUrl}`);

            // For demo purposes, we'll just show an alert or redirect to google or configured URL
            alert('Em ambiente de produção, isto liberaria o acesso no OpenNDS.');
            const redirectUrl = state.redirectUrl || "https://www.google.com";
            window.location.href = redirectUrl;
        });

        ui.btnVoucher.addEventListener('click', handleVoucherSubmit);

        // Allow pressing Enter on the voucher input
        if (ui.inputVoucher) {
            ui.inputVoucher.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') handleVoucherSubmit();
            });
        }

        // Add User Data form events
        if (ui.btnCancelUserData) {
            ui.btnCancelUserData.addEventListener('click', () => {
                switchScreen('plans');
            });
        }

        if (ui.formUserData) {
            // Apply CPF mask as user types
            ui.inputCpf.addEventListener('input', function (e) {
                let value = e.target.value.replace(/\D/g, "");
                if (value.length > 11) value = value.slice(0, 11);

                if (value.length > 9) {
                    value = value.replace(/(\d{3})(\d{3})(\d{3})(\d{1,2})/, "$1.$2.$3-$4");
                } else if (value.length > 6) {
                    value = value.replace(/(\d{3})(\d{3})(\d{1,3})/, "$1.$2.$3");
                } else if (value.length > 3) {
                    value = value.replace(/(\d{3})(\d{1,3})/, "$1.$2");
                }
                e.target.value = value;
            });

            // Handle submission to generate PIX
            ui.formUserData.addEventListener('submit', async (e) => {
                e.preventDefault();
                await submitUserDataAndGeneratePix();
            });
        }
    }

    /**
     * Load Plan list from API and render them
     */
    async function loadPlans() {
        ui.planList.innerHTML = '<div class="text-center py-4 text-slate-400"><i class="fa-solid fa-circle-notch fa-spin text-2xl text-brand-500 mb-2"></i><p>Carregando planos...</p></div>';

        try {
            const plans = await API.getPlans();
            ui.planList.innerHTML = ''; // Clear loader

            plans.forEach((plan, index) => {
                const delay = index * 100; // Staggered animation

                const card = document.createElement('div');
                card.className = `plan-card group relative bg-slate-900/60 backdrop-blur-md border ${plan.popular ? 'border-brand-500/50 shadow-[0_0_20px_rgba(0,229,255,0.1)]' : 'border-white/5'} hover:border-brand-500/50 hover:shadow-[0_0_25px_rgba(0,229,255,0.2)] rounded-3xl p-5 cursor-pointer overflow-hidden fade-in transition-all duration-300`;
                card.style.animationDelay = `${delay}ms`;

                const popularBadge = plan.popular ?
                    `<div class="absolute top-0 right-0 bg-brand-500 text-slate-950 text-[10px] font-bold px-3 py-1 rounded-bl-xl shadow-[0_0_10px_rgba(0,229,255,0.3)] z-20">RECOMENDADO</div>` : '';

                let customHtml = '';
                if (plan.isCustom) {
                    customHtml = `
                    <div class="custom-hours-container mt-4 hidden transition-all duration-300 relative z-20" onclick="event.stopPropagation()">
                        <label class="block text-xs text-slate-400 uppercase tracking-widest font-semibold mb-2 ml-1">Quantas horas?</label>
                        <div class="flex gap-2">
                            <input type="number" min="1" max="72" value="1" class="custom-hours-input w-24 bg-slate-950/80 border border-slate-700/80 rounded-xl py-2 px-3 text-white font-mono text-center focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 shadow-inner">
                            <button class="btn-confirm-custom flex-1 bg-brand-600 hover:bg-brand-500 text-slate-950 font-bold shadow-[0_0_10px_rgba(0,229,255,0.2)] hover:shadow-[0_0_15px_rgba(0,229,255,0.4)] rounded-xl py-2 transition-all active:scale-[0.98]">
                                Confirmar
                            </button>
                        </div>
                    </div>
                    `;
                }

                card.innerHTML = `
                    <div class="absolute inset-0 bg-gradient-to-br from-brand-500/5 to-purple-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none z-0"></div>
                    ${popularBadge}
                    <div class="flex justify-between items-start relative z-10 ${plan.popular ? 'mt-3 sm:mt-1' : ''}">
                        <div class="flex flex-col mb-1">
                            <h3 class="text-[1.15rem] font-bold text-white tracking-tight leading-tight group-hover:text-brand-400 transition-colors">${plan.name}</h3>
                            <div class="flex items-center text-xs text-slate-400 gap-1 mt-1 font-medium">
                                <i class="fa-regular fa-clock text-brand-500/70"></i> ${plan.duration}
                            </div>
                        </div>
                        <div class="text-right">
                            <span class="text-brand-400 font-bold ${plan.isCustom ? 'text-xl' : 'text-2xl'} tracking-tight drop-shadow-[0_0_8px_rgba(0,229,255,0.3)]">${plan.price}</span>
                        </div>
                    </div>
                    <p class="text-sm text-slate-400/80 leading-snug mt-2 relative z-10">${plan.description}</p>
                    ${customHtml}
                `;

                if (plan.isCustom) {
                    const container = card.querySelector('.custom-hours-container');
                    const btnConfirm = card.querySelector('.btn-confirm-custom');
                    const inputHours = card.querySelector('.custom-hours-input');

                    card.addEventListener('click', () => {
                        // Expand the container
                        document.querySelectorAll('.custom-hours-container').forEach(c => c.classList.add('hidden'));
                        container.classList.remove('hidden');
                        inputHours.focus();
                    });

                    btnConfirm.addEventListener('click', (e) => {
                        e.stopPropagation();
                        // Overwrite plan details to pass to handlePlanSelection
                        const hours = parseInt(inputHours.value) || 1;
                        const dynamicPlan = {
                            ...plan,
                            duration: `${hours} Horas`,
                            dynamicHours: hours
                        };
                        handlePlanSelection(dynamicPlan);
                    });
                } else {
                    card.addEventListener('click', () => handlePlanSelection(plan));
                }

                ui.planList.appendChild(card);
            });

        } catch (error) {
            console.error("Error loading plans:", error);
            ui.planList.innerHTML = '<div class="text-center py-4 text-red-400"><p>Erro ao carregar planos. Tente novamente.</p></div>';
        }
    }

    /**
     * Handle Plan Selection (Click)
     */
    async function handlePlanSelection(plan) {
        // Visual feedback on click
        const cards = document.querySelectorAll('.plan-card');
        cards.forEach(c => c.style.opacity = '0.5');

        // Transition to User Data screen to safely collect info for MP first
        window.__selectedPlan = plan;

        setTimeout(() => {
            cards.forEach(c => c.style.opacity = '1');
            switchScreen('userData');
            ui.inputNome.focus();
        }, 150);
    }

    /**
     * Submits user data and actual generates the PIX payload
     */
    async function submitUserDataAndGeneratePix() {
        const btn = document.getElementById('btn-submit-user-data');
        btn.innerHTML = '<i class="fa-solid fa-circle-notch fa-spin"></i> Processando...';
        btn.disabled = true;

        const nome = ui.inputNome.value.trim();
        const sobrenome = ui.inputSobrenome.value.trim();
        const cpf = ui.inputCpf.value.replace(/\D/g, "");

        if (cpf.length !== 11) {
            alert("Por favor, introduza um CPF válido com 11 algarismos.");
            btn.innerHTML = 'Continuar Pagamento <i class="fa-solid fa-arrow-right"></i>';
            btn.disabled = false;
            ui.inputCpf.focus();
            return;
        }

        try {
            // Request PIX Generation with the collected customer data
            const data = await API.gerarPix(state.userMac, window.__selectedPlan.id, window.__selectedPlan.dynamicHours, nome, sobrenome, cpf);
            state.currentTxid = data.txid;

            // Populate Payment Screen
            ui.qrImage.src = data.qr_code || `https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${encodeURIComponent(data.pix_copia_cola)}`;
            ui.pixCodeInput.value = data.pix_copia_cola;

            // Switch UI
            switchScreen('payment');

            // Start Timers
            startCountdown();
            startPolling();

        } catch (error) {
            console.error("Error generating PIX:", error);
            alert("Erro ao processar pagamento com o Mercado Pago. Tente novamente ou verifique os seus dados.");
        } finally {
            btn.innerHTML = 'Continuar Pagamento <i class="fa-solid fa-arrow-right"></i>';
            btn.disabled = false;
        }
    }

    /**
     * Handle Voucher Submission
     */
    async function handleVoucherSubmit() {
        const codigo = ui.inputVoucher.value.trim();

        if (!codigo) return;

        // UI Feedback
        ui.btnVoucher.disabled = true;
        ui.btnVoucher.innerHTML = '<i class="fa-solid fa-circle-notch fa-spin"></i> Validando...';
        ui.voucherError.classList.add('hidden');

        try {
            const data = await API.resgatarVoucher(state.userMac, codigo);

            if (data.sucesso) {
                // Store plan info for success screen
                window.__selectedPlan = data.plano;
                showSuccessScreen();
            } else {
                ui.voucherError.textContent = data.erro || "Código inválido.";
                ui.voucherError.classList.remove('hidden');
                ui.inputVoucher.focus();
            }
        } catch (error) {
            console.error("Erro ao validar voucher:", error);
            ui.voucherError.textContent = "Erro de conexão. Tente novamente.";
            ui.voucherError.classList.remove('hidden');
        } finally {
            // Restore button
            ui.btnVoucher.disabled = false;
            ui.btnVoucher.textContent = 'Ativar Internet';
        }
    }

    /**
     * Copy PIX to Clipboard with visual feedback
     */
    function copyPixCode() {
        ui.pixCodeInput.select();
        ui.pixCodeInput.setSelectionRange(0, 99999); /* For mobile devices */

        navigator.clipboard.writeText(ui.pixCodeInput.value).then(() => {
            // Button 1
            const originalHtmlMain = ui.btnCopyMain.innerHTML;
            ui.btnCopyMain.innerHTML = '<i class="fa-solid fa-check"></i><span>Copiado!</span>';
            ui.btnCopyMain.classList.replace('bg-brand-600', 'bg-emerald-600');
            ui.btnCopyMain.classList.replace('hover:bg-brand-500', 'hover:bg-emerald-500');

            // Reset after 3 seconds
            setTimeout(() => {
                ui.btnCopyMain.innerHTML = originalHtmlMain;
                ui.btnCopyMain.classList.replace('bg-emerald-600', 'bg-brand-600');
                ui.btnCopyMain.classList.replace('hover:bg-emerald-500', 'hover:bg-brand-500');
            }, 3000);
        }).catch(err => {
            console.error('Failed to copy text: ', err);
            alert("Erro ao copiar. Por favor, selecione e copie o texto manualmente.");
        });
    }

    /**
     * Switch visible screens with animation
     */
    function switchScreen(targetScreenName) {
        Object.keys(screens).forEach(key => {
            if (key === targetScreenName) {
                screens[key].classList.remove('hidden');
                // Force triggering animation
                screens[key].style.opacity = '0';
                setTimeout(() => {
                    screens[key].style.transition = 'opacity 0.3s ease-in, transform 0.3s ease-out';
                    screens[key].style.opacity = '1';
                    screens[key].style.transform = 'translateY(0)';
                }, 10);
            } else {
                screens[key].classList.add('hidden');
                screens[key].style.opacity = '0';
                screens[key].style.transform = 'translateY(10px)';
            }
        });
    }

    /**
     * Countdown Timer Logic
     */
    function startCountdown() {
        state.timeLeft = 600; // Reset to 10 min
        updateCountdownUI();

        if (state.countdownInterval) clearInterval(state.countdownInterval);

        state.countdownInterval = setInterval(() => {
            state.timeLeft--;
            updateCountdownUI();

            if (state.timeLeft <= 0) {
                stopCountdown();
                stopPolling();
                alert("O tempo do Jardim Murado expirou. Você será desconectado.");
                // Em produção, isso iria redirecionar para uma página de retry ou bloquear navegação
                window.location.reload();
            }
        }, 1000);
    }

    function updateCountdownUI() {
        const minutes = Math.floor(state.timeLeft / 60);
        const seconds = state.timeLeft % 60;
        ui.countdown.textContent = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;

        if (state.timeLeft < 60) {
            ui.countdown.classList.add('text-red-400');
            ui.countdown.classList.remove('text-amber-300');
        }
    }

    function stopCountdown() {
        if (state.countdownInterval) clearInterval(state.countdownInterval);
    }

    /**
     * Polling Logic to check payment status
     */
    function startPolling() {
        if (state.pollingInterval) clearInterval(state.pollingInterval);

        state.pollingInterval = setInterval(async () => {
            try {
                const res = await API.verificarPagamento(state.currentTxid);
                if (res.status === 'pago') {
                    stopPolling();
                    stopCountdown();
                    showSuccessScreen();
                }
            } catch (error) {
                console.error("Polling error:", error);
            }
        }, 5000); // 5 seconds interval
    }

    function stopPolling() {
        if (state.pollingInterval) clearInterval(state.pollingInterval);
    }

    /**
     * Show Success Final Screen
     */
    function showSuccessScreen() {
        const plan = window.__selectedPlan || { duration: 'Tempo Contratado' };

        // Calculate expiration mock
        const now = new Date();
        const expire = new Date(now.getTime() + 24 * 60 * 60 * 1000); // Mocking +24h just for display

        ui.successTime.textContent = plan.duration;
        ui.successExpire.textContent = `${expire.toLocaleDateString('pt-BR')} às ${expire.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}`;

        switchScreen('success');

        // Animate icon
        setTimeout(() => {
            ui.successIcon.style.transform = 'scale(1)';
        }, 100);
    }

    /**
     * Hidden debug trigger to force payment success
     * Click on the Pix Icon
     */
    function addDebugTrigger() {
        const pixIcons = document.querySelectorAll('.fa-pix');
        pixIcons.forEach(icon => {
            icon.addEventListener('dblclick', () => {
                console.log("[DEBUG] Forcing payment success!");
                window.__forcePaymentSuccess = true;
                // It will be caught by the next polling tick
            });
        });
    }

    // Start App
    init();
});
