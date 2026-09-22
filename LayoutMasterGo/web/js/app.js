const API_BASE = '/api/v1';

// Função global de notificação por Toast
function showToast(message, type = 'danger') {
    const toastEl = document.getElementById('appToast');
    const toastMessage = document.getElementById('toastMessage');

    if (!toastEl || !toastMessage) {
        alert(message);
        return;
    }

    toastMessage.textContent = message;
    // Tipos suportados pelo Bootstrap: 'danger', 'success', 'warning', 'info'
    toastEl.className = `toast align-items-center text-white bg-${type} border-0`;

    const bsToast = new bootstrap.Toast(toastEl, { delay: 4000 });
    bsToast.show();
}

async function apiRequest(endpoint, method = 'GET', body = null) {
    const options = { method, headers: { 'Content-Type': 'application/json' } };
    if (body) options.body = JSON.stringify(body);
    
    const response = await fetch(`${API_BASE}${endpoint}`, options);
    if (!response.ok) {
        const errorText = await response.text();
        let cleanMessage = `Erro ${response.status}`;

        try {
            const parsed = JSON.parse(errorText);
            if (parsed && parsed.error) {
                cleanMessage = parsed.error;
            } else if (parsed && parsed.message) {
                cleanMessage = parsed.message;
            } else {
                cleanMessage = errorText;
            }
        } catch (e) {
            cleanMessage = errorText || `Erro ${response.status}`;
        }

        throw new Error(cleanMessage);
    }
    if (response.status === 204) return true;
    return await response.json();
}

function appRouter() {
    return {
        activeTab: 'containers',
        async init() {
            await this.navigate('containers');
        },
        async navigate(tab) {
            this.activeTab = tab;
            const viewArea = document.getElementById('view-area');
            if (!viewArea) return;

            try {
                const response = await fetch(`/static/templates/${tab}.html`);
                if (!response.ok) throw new Error('Template não encontrado');
                
                const html = await response.text();
                viewArea.innerHTML = html;   // Injeta o HTML diretamente no DOM
                Alpine.initTree(viewArea);    // Garante a compilação reativa imediata
            } catch (err) {
                console.error("Erro na navegação:", err);
                viewArea.innerHTML = `<div class="alert alert-danger m-3">Erro ao carregar o módulo <strong>${tab}</strong>.</div>`;
            }
        }
    };
}

function closeForm(form, modalId) {
    if (form && typeof form.reset === 'function') form.reset();
    const modalEl = document.getElementById(modalId);
    if (modalEl) {
        const modal = bootstrap.Modal.getInstance(modalEl);
        if (modal) modal.hide();
    }
}

document.addEventListener('alpine:init', () => {
    Alpine.data('appRouter', appRouter);
});