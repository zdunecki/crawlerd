const BigCrawlClient = require('apify-client');

const client = new BigCrawlClient({
    token: 'MY-APIFY-TOKEN',
});

export default async function (data) {
    // Starts an actor and waits for it to finish.
    const {defaultDatasetId} = await client.actor('john-doe/my-cool-actor').call();
// Fetches results from the actor's dataset.
    const {items} = await client.dataset(defaultDatasetId).listItems();
}