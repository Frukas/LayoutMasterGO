function productsComponent() {
    return {
        products: [],
        searchQuery: '',
        page: 1,
        pageSize: 10,
        newProduct: { name: '', jan_code: '' },
        loading: false,

        async init() {
            await this.loadProducts();
        },

        async loadProducts() {
            this.loading = true;
            try {
                const q = encodeURIComponent(this.searchQuery.trim());
                const res = await apiRequest(`/products?page=${this.page}&pageSize=${this.pageSize}&search=${q}`);
                
                if (Array.isArray(res)) {
                    this.products = res;
                } else if (res && Array.isArray(res.data)) {
                    this.products = res.data;
                } else {
                    this.products = [];
                }
            } catch (err) {
                console.error("Erro ao carregar produtos:", err);
                showToast("Erro ao carregar produtos: " + err.message, "danger");
            } finally {
                this.loading = false;
            }
        },

        async onSearch() {
            this.page = 1;
            await this.loadProducts();
        },

        async changePage(delta) {
            if (this.page + delta < 1) return;
            this.page += delta;
            await this.loadProducts();
        },

        async createProduct() {
            try {
                await apiRequest('/products', 'POST', this.newProduct);
                this.newProduct = { name: '', jan_code: '' };
                closeForm(null, 'modalProduct');
                showToast("Produto criado com sucesso!", "success");
                await this.loadProducts();
            } catch (err) {
                showToast(err.message, "danger");
            }
        },

        async deleteProduct(id) {
            if (!confirm(`Deseja excluir este produto?`)) return;
            try {
                await apiRequest(`/products/${id}`, 'DELETE');
                showToast("Produto excluído com sucesso!", "success");
                await this.loadProducts();
            } catch (err) {
                showToast(err.message, "danger");
            }
        }
    };
}

document.addEventListener('alpine:init', () => {
    Alpine.data('productsComponent', productsComponent);
});