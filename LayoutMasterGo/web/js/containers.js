function containersComponent() {
    return {
        containers: [],
        availableProducts: [],
        searchQuery: '',
        productSearchQuery: '', 
        showProductDropdown: false,
        page: 1,
        pageSize: 10,
        viewMode: 'list', // Modos: 'list', 'details', 'create'
        selectedContainer: null,
        
        // Modal de adicionar item a contêiner existente
        newItem: { product_id: '', quantity: 1 }, 
        
        // Form de criação de novo contêiner
        newContainer: { name: '', data: '', status: 'Pending', items: [] },
        tempItem: { product_id: '', quantity: 1 },

        loading: false,

        init: async function() {
            await Promise.all([
                this.loadContainers(),
                this.loadAvailableProducts()
            ]);
        },

        loadContainers: async function() {
            this.loading = true;
            try {
                const q = encodeURIComponent(this.searchQuery.trim());
                const res = await apiRequest(`/containers?page=${this.page}&pageSize=${this.pageSize}&search=${q}`);
                if (Array.isArray(res)) {
                    this.containers = res;
                } else if (res && Array.isArray(res.data)) {
                    this.containers = res.data;
                } else if (res && Array.isArray(res.items)) {
                    this.containers = res.items;
                } else {
                    this.containers = [];
                }
            } catch (err) {
                console.error("Erro ao carregar contêineres:", err);
                showToast("Erro ao carregar contêineres: " + err.message, "danger");
            } finally {
                this.loading = false;
            }
        },

        onSearch: async function() {
            this.page = 1;
            await this.loadContainers();
        },

        changePage: async function(delta) {
            if (this.page + delta < 1) return;
            this.page += delta;
            await this.loadContainers();
        },

        loadAvailableProducts: async function() {
            try {
                const q = encodeURIComponent(this.productSearchQuery.trim());
                const res = await apiRequest(`/products?page=1&pageSize=50&search=${q}`);
                if (Array.isArray(res)) {
                    this.availableProducts = res;
                } else if (res && Array.isArray(res.data)) {
                    this.availableProducts = res.data;
                } else {
                    this.availableProducts = [];
                }
            } catch (err) {
                console.error("Erro ao carregar lista de produtos:", err);
            }
        },

        onProductSearch: async function() {
            this.tempItem.product_id = '';
            this.newItem.product_id = '';
            await this.loadAvailableProducts();
            this.showProductDropdown = true;
        },

        selectProduct: function(p) {
            const pId = p.ID || p.id;
            this.tempItem.product_id = pId;
            this.newItem.product_id = pId;
            
            const jan = p.jan_code || p.JanCode || '-';
            const name = p.name || p.Name || '';
            this.productSearchQuery = `${jan} - ${name}`;
            this.showProductDropdown = false;
        },

        getProduct: function(item) {
            const pId = item.product_id || item.ProductID;
            if (item.product || item.Product) return item.product || item.Product;
            return this.availableProducts.find(p => (p.ID || p.id) == pId) || {};
        },

        getJanCode: function(item) {
            const p = this.getProduct(item);
            return p.jan_code || p.JanCode || '-';
        },

        getProductName: function(item) {
            const p = this.getProduct(item);
            return p.name || p.Name || 'Produto não identificado';
        },

        get totalQuantity() {
            if (!this.selectedContainer) return 0;
            const items = this.selectedContainer.item_containers || this.selectedContainer.items || this.selectedContainer.ItemContainers || [];
            return items.reduce((sum, item) => sum + (parseInt(item.quantity || item.Quantity, 10) || 0), 0);
        },

        viewDetails: async function(containerId) {
            this.loading = true;
            try {
                this.selectedContainer = await apiRequest(`/containers/${containerId}`);
                this.viewMode = 'details';
            } catch (err) {
                showToast(err.message, "danger");
            } finally {
                this.loading = false;
            }
        },

        backToList: function() {
            this.viewMode = 'list';
            this.selectedContainer = null;
            this.loadContainers();
        },

        openCreateMode: function() {
            this.newContainer = { name: '', data: '', status: 'Pending', items: [] };
            this.tempItem = { product_id: '', quantity: 1 };
            this.productSearchQuery = '';
            this.showProductDropdown = false;
            this.loadAvailableProducts(); 
            this.viewMode = 'create';
        },

        addTempItem: async function() {
            if (!this.tempItem.product_id) return showToast("Selecione um produto da lista.", "warning");
            if (this.tempItem.quantity <= 0) return showToast("A quantidade deve ser maior que zero.", "warning");

            const pId = parseInt(this.tempItem.product_id, 10);
            const qty = parseInt(this.tempItem.quantity, 10);

            const product = this.availableProducts.find(p => (p.ID || p.id) == pId) || {};
            const pName = product.name || product.Name || 'Produto Desconhecido';
            const pJan = product.jan_code || product.JanCode || '-';

            const existing = this.newContainer.items.find(i => i.product_id === pId);
            if (existing) {
                existing.quantity += qty;
            } else {
                this.newContainer.items.push({ 
                    product_id: pId, 
                    quantity: qty,
                    name: pName,
                    jan_code: pJan
                });
            }

            this.tempItem = { product_id: '', quantity: 1 };
            this.productSearchQuery = '';
            this.showProductDropdown = false;
            await this.loadAvailableProducts(); 
        },

        removeTempItem: function(index) {
            this.newContainer.items.splice(index, 1);
        },

        get newContainerTotalQty() {
            return this.newContainer.items.reduce((sum, i) => sum + i.quantity, 0);
        },

        saveNewContainer: async function() {
            if (!this.newContainer.name.trim()) return showToast("O nome do contêiner é obrigatório.", "warning");
            if (!this.newContainer.data) return showToast("A data do contêiner é obrigatória.", "warning");

            this.loading = true;
            try {
                let formattedDate = this.newContainer.data;
                if (formattedDate.length === 10) {
                    formattedDate += 'T00:00:00Z'; 
                }

                const payload = {
                    name: this.newContainer.name.trim(),
                    data: formattedDate,
                    status: this.newContainer.status || 'Pending'
                };

                const created = await apiRequest('/containers', 'POST', payload);
                const containerId = created.ID || created.id;

                if (!containerId) throw new Error("ID do contêiner não retornado pelo servidor.");

                if (this.newContainer.items.length > 0) {
                    const promises = this.newContainer.items.map(item => {
                        const itemPayload = {
                            product_id: item.product_id,
                            quantity: item.quantity
                        };
                        return apiRequest(`/containers/${containerId}/items`, 'POST', itemPayload);
                    });
                    await Promise.all(promises);
                }

                showToast("Contêiner criado com sucesso!", "success");
                await this.viewDetails(containerId);

            } catch (err) {
                showToast(err.message, "danger");
            } finally {
                this.loading = false;
            }
        },

        addItem: async function() {
            if (!this.newItem.product_id) return showToast("Selecione um produto da lista.", "warning");
            
            try {
                const containerId = this.selectedContainer.ID || this.selectedContainer.id;
                const payload = {
                    product_id: parseInt(this.newItem.product_id, 10),
                    quantity: parseInt(this.newItem.quantity, 10)
                };
                await apiRequest(`/containers/${containerId}/items`, 'POST', payload);
                
                this.newItem = { product_id: '', quantity: 1 };
                this.productSearchQuery = '';
                this.showProductDropdown = false;
                await this.loadAvailableProducts(); 
                
                closeForm(null, 'modalAddItem');
                showToast("Produto adicionado com sucesso!", "success");
                await this.viewDetails(containerId);
            } catch (err) {
                showToast(err.message, "danger");
            }
        },

        removeItem: async function(productId) {
            if (!confirm(`Remover este produto do contêiner?`)) return;
            try {
                const containerId = this.selectedContainer.ID || this.selectedContainer.id;
                await apiRequest(`/containers/${containerId}/items/${productId}`, 'DELETE');
                showToast("Produto removido com sucesso!", "success");
                await this.viewDetails(containerId);
            } catch (err) {
                showToast(err.message, "danger");
            }
        }
    };
}

document.addEventListener('alpine:init', () => {
    Alpine.data('containersComponent', containersComponent);
});