function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

// TODO: smooth scroll

(async () => {
    const products = []

    // get product info
    function getDataFromProductCard(product) {
        const name = product.querySelector('.product-card__title')?.innerText
        const category = product.querySelector('.product-card__subtitle')?.innerText
        const variantsLength = product.querySelector('.product-card__product-count')?.innerText
        const price = product.querySelector('.product-price')?.innerText

        products.push({name, category, variantsLength, price})
    }

    //

    // observers
    const productsObserver = new MutationObserver(function (e) {
        e.forEach(function (mutation) {
            if (mutation.type != 'childList') {
                return
            }

            const productCard = mutation?.addedNodes && mutation.addedNodes[0]

            getDataFromProductCard(productCard)
        });
    });
    //

    // initial data feed
    document.querySelectorAll('.product-card').forEach(getDataFromProductCard)
    //

    // accept cookies
    document.querySelector('#hf_cookie_text_cookieAccept').click()
    //

    // observer initialization
    const productsGrid = document.querySelector('.product-grid__items')
    productsObserver.observe(productsGrid, {subtree: false, childList: true});
    //

    // scroll
    setInterval(() => {
        window.scrollTo(0, document.body.scrollHeight);
    }, 1000)
    //

    // sleep
    await sleep(1000 * 30) // TODO: sleep is sad - find better solution
    //

    return {
        url: window.location.href,
        products,
    }
})()

