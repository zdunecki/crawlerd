const API_URL = window.CRAWLERD_API_URL || "http://localhost:8080/v1"
const RUN_ID = window.CRAWLERD_RUN_ID || ""
const DEPTH = window.CRAWLERD_DEPTH || ""

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
function API(token) {
    return {
        RequestQueue: {
            // TODO: use real api instead of mock
            async add(links) {
                const body = []

                for (const link of links) {
                    // TODO: be sure that links are absolute
                    body.push({
                        run_id: RUN_ID,
                        url: link,
                        depth: parseInt(DEPTH)
                    })
                }

                await fetch(`${API_URL}/request-queue/batch`, {
                    method: "POST",
                    body: JSON.stringify(body)
                })
            }
        },
    };
}

const api = API();

(async () => {
    console.log("start", JSON.stringify({
        API_URL,
        RUN_ID,
        href: window.location.href
    }))
    const links = new Set()

    await new Promise((resolve, reject) => {
        const maxLastScrollHeight = 3
        const maxNoMoreContent = 3

        let searching
        let lastScrollHeight = 0
        let lastScrollHeightTry = 0
        let maxScrollHeight = 0
        let noMoreContentTry = 0

        function searchLinks() {
            searching = true
            const as = document.querySelectorAll('a')

            if (!as || !as.length) {
                searching = false
                return
            }

            as.forEach(a => {
                links.add(a.href)
            })

            searching = false
        }

        // search before mutation observer
        searchLinks()

        // TODO: find better solution
        const linksObserver = new MutationObserver(function () {
            if (!searching) {
                searchLinks()
            }
        });

        linksObserver.observe(document, {
            subtree: true,
            childList: true
        });

        let scrollID = setInterval(async () => {
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
            const copy = [...new Set(links)]

            if (!copy.length) {
                return
            }

            // TODO: dont' send same links between snapshots
            await api.RequestQueue.add(copy)

            copy.forEach(c => links.delete(c))
        }, 2000)
    })

    // TODO: dont' return links because it's already send by api - send status or something else
    console.log(links, 'finished')

    return {
        status: 'OK'
    }
})()

