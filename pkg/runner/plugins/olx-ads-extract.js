(async () => {
    const ads = []

    document.querySelectorAll('.offer [data-id]').forEach(offer => {
        const title = offer.querySelector("[data-cy='listing-ad-title']").innerText
        const price = offer.querySelector(".price").innerText

        ads.push({title, price})
    })

    return {
        url: window.location.href,
        ads,
    }
})()