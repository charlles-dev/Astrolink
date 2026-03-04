/**
 * API Module to interact with the Python Backend.
 * Since the backend is not yet implemented, this module mocks the responses
 * according to the frontend documentation.
 */

const API = {
    // Helper to handle JSON response and errors
    _fetchJSON: async (url, options = {}) => {
        try {
            const response = await fetch(url, options);
            if (!response.ok) {
                let errorData = {};
                try { errorData = await response.json(); } catch (e) { }
                throw new Error(errorData.detail || `HTTP error! status: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            console.error(`API Fetch Error [${url}]:`, error);
            throw error;
        }
    },

    /**
     * Fetch available plans from backend
     */
    getPlans: async () => {
        return await API._fetchJSON('/api/planos');
    },

    /**
     * Fetch public system settings
     */
    getSettings: async () => {
        return await API._fetchJSON('/api/settings');
    },

    /**
     * Generate PIX
     */
    gerarPix: async (userMac, planId, customHours = null, nome = null, sobrenome = null, cpf = null) => {
        const payload = {
            mac: userMac,
            plano_id: planId,
            custom_hours: customHours,
            nome: nome,
            sobrenome: sobrenome,
            cpf: cpf
        };

        return await API._fetchJSON('/api/gerar-pix', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
    },

    /**
     * Poll Payment Status
     */
    verificarPagamento: async (txid) => {
        return await API._fetchJSON(`/api/status-pix?txid=${txid}`);
    },

    /**
     * Resgatar Voucher (PIN code)
     */
    resgatarVoucher: async (userMac, codigo) => {
        const payload = {
            mac: userMac,
            codigo: codigo
        };

        // Note: For vouchers, our backend handles exceptions as 400s or returns {sucesso: false, erro: ...}
        // So we might need to handle it properly, but fetchJSON handles 200 properly.
        try {
            const response = await fetch('/api/resgatar-voucher', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            const data = await response.json();

            if (!response.ok) {
                return { sucesso: false, erro: data.detail || "Erro de servidor." };
            }
            return data;
        } catch (error) {
            return { sucesso: false, erro: "Erro de conexão." };
        }
    }
};
