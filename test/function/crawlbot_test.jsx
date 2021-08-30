import React from "react";
import {renderToStaticMarkup as html} from 'react-dom/server'; // TODO: create generator without explicit call ReactDOMServer?

const internalLink1Name = "search-me"
const internalLink1Link = "/" + internalLink1Name

const internalLink2Name = "cool-article"
const internalLink2Link = "/" + internalLink2Name

const externalServer1URL = "http://url-to-another-server:8282"

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

// TODO: more boust expect, like {links: [], requestToPages: []}
export function TestData({fakeServer}) {
    return [
        // {
        //     description: "two",
        //     max_depth: 2,
        //     pages: {
        //         [fakeServer + internalLink1Link]: html(<PageWithNoLinks/>),
        //         [externalServer1URL]: html(<PageWithOneInternalLink internalLinkName={internalLink2Name}/>),
        //         [fakeServer + internalLink2Link]: html(<PageWithNoLinks/>),
        //     },
        //     url: fakeServer + "/some-url",
        //     body: html(<PageWithInternalAndExternalLink
        //         internalLinkName={internalLink1Name}
        //     />),
        //     expect: [
        //         fakeServer + internalLink1Link,
        //         externalServer1URL,
        //         fakeServer + internalLink2Link
        //     ]
        // },
        {
            url: fakeServer + "/some-url",
            description: "one",
            max_depth: 1,
            body: html(<PageWithInternalAndExternalLink
                internalLinkName={internalLink1Name}
            />),
            pages: {
                [fakeServer + internalLink1Link]: html(<PageWithNoLinks/>),
                [externalServer1URL]: html(<PageWithOneInternalLink internalLinkName={internalLink2Name}/>),
                [fakeServer + internalLink2Link]: html(<PageWithNoLinks/>),
            },
            expect: [
                fakeServer + internalLink1Link,
                externalServer1URL
            ]
        }
    ]
}