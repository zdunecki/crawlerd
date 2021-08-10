function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

function rand(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

function avoidScrollLockIn() {
    const body = document.querySelector('body')
    const computedStyles = window.getComputedStyle(body, null) || {}

    // to avoid scroll lock-in
    if (computedStyles.position === "fixed") {
        body.style.position = "inherit"
    }
}

// TODO: pass token
const API = {
    // TODO: use real api instead of mock
    async patchIndices(links) {
        await sleep(1000)

    }
};

(async () => {
    const links = {}

    await new Promise((resolve, reject) => {
        const maxLastScrollHeight = 3
        const maxNoMoreContent = 3

        let searching
        let lastScrollHeight = 0
        let lastScrollHeightTry = 0
        let maxScrollHeight = 0
        let noMoreContentTry = 0

        // TODO: find better solution
        const linksObserver = new MutationObserver(function () {
            if (!searching) {
                searching = true
                const as = document.querySelectorAll('a')

                if (!as || !as.length) {
                    searching = false
                    return
                }

                as.forEach(a => {
                    links[a.href] = a.href
                })

                searching = false
            }
        });

        linksObserver.observe(document, {
            subtree: true,
            childList: true
        });

        let scrollID = setInterval(() => {
            avoidScrollLockIn()

            // TODO: improve scroll - smooth like a human
            window.scrollTo(0, document.body.scrollHeight);
            lastScrollHeight = document.body.scrollHeight

            if (lastScrollHeight > maxScrollHeight) {
                maxScrollHeight = lastScrollHeight
            }

            if (lastScrollHeight === document.body.scrollHeight) {
                if (lastScrollHeightTry >= maxLastScrollHeight) {
                    if (maxScrollHeight === document.body.scrollHeight) {
                        noMoreContentTry++

                        if (noMoreContentTry >= maxNoMoreContent) {
                            // TODO: clear interval if there's no more scrollable content
                            // TODO: currently it's working only by checking last max scrollheight but we must reset counter if there's network requests
                            // 1) check if there's no network xhr/fetch requests
                            // 2) check if scrollHeight is same for X seconds
                            clearInterval(scrollID);
                            resolve()
                        }
                    }

                    window.scrollTo(0, document.body.scrollHeight - (document.body.scrollHeight / 4));
                    lastScrollHeightTry = 0
                    return
                }
                lastScrollHeightTry++
            }
        }, rand(600, 1200))

        setInterval(async () => {
            const copy = {...links}

            // TODO: dont' send same links between snapshots
            await API.patchIndices(copy)

            for (let href in copy) {
                delete links[href]
            }

        }, 2000)
    })

    // TODO: dont' return links because it's already send by api - send status or something else
    console.log(links, 'finished')

    return {
        status: 'OK'
    }
})()

