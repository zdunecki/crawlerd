(() => {
    // OUTPUT/610dc02e4d0c91c8b2b4fd23/index.js
    function sleep(ms) {
        return new Promise((resolve) => setTimeout(resolve, ms));
    }
    (async () => {
        const products = [];
        function getDataFromProductCard(product) {
            const name = product.querySelector(".product-card__title")?.innerText;
            const category = product.querySelector(".product-card__subtitle")?.innerText;
            const variantsLength = product.querySelector(".product-card__product-count")?.innerText;
            const price = product.querySelector(".product-price")?.innerText;
            products.push({ name, category, variantsLength, price });
        }
        const productsObserver = new MutationObserver(function(e) {
            e.forEach(function(mutation) {
                if (mutation.type != "childList") {
                    return;
                }
                const productCard = mutation?.addedNodes && mutation.addedNodes[0];
                getDataFromProductCard(productCard);
            });
        });
        document.querySelectorAll(".product-card").forEach(getDataFromProductCard);
        document.querySelector("#hf_cookie_text_cookieAccept").click();
        const productsGrid = document.querySelector(".product-grid__items");
        productsObserver.observe(productsGrid, { subtree: false, childList: true });
        setInterval(() => {
            window.scrollTo(0, document.body.scrollHeight);
        }, 1e3);
        await sleep(1e3 * 30);
        return {
            url: window.location.href,
            products
        };
    })();
})();
