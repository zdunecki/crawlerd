import React from "react";
import {renderToStaticMarkup as html} from 'react-dom/server'; // TODO: create generator without explicit call ReactDOMServer?

const internalLink1Name = "search-me"
const internalLink1Link = "/" + internalLink1Name

const internalLink2Name = "cool-article"
const internalLink2Link = "/" + internalLink2Name

// const externalServer1URL = "http://url-to-another-server:8282"
const externalServer1URL = "http://localhost:8282"

function PageWithOneExternalLink() {
    return <html>
    <head>

    </head>
    <body>
    <h1>This is a heading</h1>
    <p>This is a paragraph.</p>
    <a href={externalServer1URL}>Url to another mock server</a>
    </body>
    </html>
}

function PageWithOneInternalLink({internalLinkName}) {
    return <html>
    <head>

    </head>
    <body>
    <h1>This is a heading</h1>
    <p>This is a paragraph.</p>
    <a href={'/' + internalLinkName}>Search me</a>
    </body>
    </html>
}

function PageWithNoLinks() {
    return <html>
    <head>

    </head>
    <body>
    <h1>This is a heading</h1>
    <p>This is a paragraph.</p>
    <p>Page with no links</p>
    </body>
    </html>
}

// TODO: externalLink props?
function PageWithInternalAndExternalLink({internalLinkName}) {
    return <html>
    <head>

    </head>
    <body>
    <h1>This is a heading</h1>
    <p>This is a paragraph.</p>
    <a href={'/' + internalLinkName}>Search me</a>
    <div>
        <div>
            <h1> Heading </h1>
            <a href={externalServer1URL}>Url to another mock server</a>
        </div>
    </div>
    </body>
    </html>
}

// TODO: more robust expect, like {links: [], requestToPages: []}
// TODO: outside_network should be checked only on backend side
// TODO: tests with errors on pages (not found, internal error or something else)
// TODO: tests sometimes failed if run multiple instead of single
export function TestData({rootServer}) {
    return [
        {
            start_url: rootServer + "/some-url",
            description: "one depth only",
            max_depth: 1,
            pages: {
                [rootServer + "/some-url"]: {
                    body: html(<PageWithInternalAndExternalLink
                        internalLinkName={internalLink1Name}
                    />)
                },
                [rootServer + internalLink1Link]: {
                    body: html(<PageWithNoLinks/>)
                },
                [rootServer + internalLink2Link]: {
                    body: html(<PageWithNoLinks/>),
                },
                [externalServer1URL]: {
                    body: html(<PageWithOneInternalLink internalLinkName={internalLink2Name}/>),
                    // outside_network: true,
                },
            },
            expect: [
                rootServer + internalLink1Link,
                externalServer1URL
            ]
        },
        {
            start_url: rootServer + "/some-url",
            description: "depth level 2 but deeper pages don't have links",
            max_depth: 2,
            pages: {
                [rootServer + "/some-url"]: {
                    body: html(<PageWithInternalAndExternalLink
                        internalLinkName={internalLink1Name}
                    />)
                },
                [rootServer + internalLink1Link]: {
                    body: html(<PageWithNoLinks/>)
                },
                [rootServer + internalLink2Link]: {
                    body: html(<PageWithNoLinks/>),
                },
                [externalServer1URL]: {
                    body: html(<PageWithNoLinks/>),
                    // outside_network: true,
                },
            },
            expect: [
                rootServer + internalLink1Link,
                externalServer1URL
            ]
        },
        {
            start_url: rootServer + "/some-url",
            description: "depth level 2 and page level 2 has link",
            max_depth: 2,
            pages: {
                [rootServer + "/some-url"]: {
                    body: html(<PageWithInternalAndExternalLink
                        internalLinkName={internalLink1Name}
                    />)
                },
                [rootServer + internalLink1Link]: {
                    body: html(<PageWithOneInternalLink internalLinkName={internalLink2Name}/>)
                },
                [externalServer1URL]: {
                    body: html(<PageWithNoLinks/>),
                }
            },
            expect: [
                rootServer + internalLink1Link,
                externalServer1URL,
                rootServer + internalLink2Link
            ]
        },
        {
            start_url: rootServer + "/some-url",
            description: "depth level 2 and follow only root server links",
            max_depth: 2,
            pages: {
                [rootServer + "/some-url"]: {
                    body: html(<PageWithOneExternalLink/>)
                },
                [externalServer1URL]: {
                    body: html(<PageWithOneInternalLink internalLinkName={internalLink1Name}/>),
                }
            },
            follow_links: [
                {
                    match: rootServer
                }
            ],
            expect: [
                externalServer1URL,
            ]
        },
        {
            start_url: rootServer + "/some-url",
            description: "depth level 2 and follow only root server links and scrape only root server links",
            max_depth: 2,
            pages: {
                [rootServer + "/some-url"]: {
                    body: html(<PageWithOneExternalLink/>)
                },
                [externalServer1URL]: {
                    body: html(<PageWithOneInternalLink internalLinkName={internalLink1Name}/>),
                }
            },
            follow_links: [
                {
                    match: rootServer
                }
            ],
            scrape_links_pattern: rootServer,
            expect: [

            ]
        },
        {
            start_url: rootServer + "/some-url",
            description: "depth level 2 and page level 2 has link with root server follow_links",
            max_depth: 2,
            follow_links: [
                {
                    match: new RegExp(rootServer).toString()
                }
            ],
            pages: {
                [rootServer + "/some-url"]: {
                    body: html(<PageWithInternalAndExternalLink
                        internalLinkName={internalLink1Name}
                    />)
                },
                [rootServer + internalLink1Link]: {
                    body: html(<PageWithOneInternalLink internalLinkName={internalLink2Name}/>)
                },
                [externalServer1URL]: {
                    body: html(<PageWithNoLinks/>),
                }
            },
            expect: [
                rootServer + internalLink1Link,
                externalServer1URL,
                rootServer + internalLink2Link
            ]
        }
    ]
}

