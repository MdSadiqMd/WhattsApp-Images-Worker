addEventListener('fetch', (event) => {
    event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {
    try {
        if (typeof handleRequest === 'function') {
            const result = handleRequest();
            return new Response(JSON.stringify(result), { status: 200 });
        } else {
            return new Response('Handler not available', { status: 500 });
        }
    } catch (error) {
        return new Response(`Error: ${error.message}`, { status: 500 });
    }
}

addEventListener('scheduled', (event) => {
    event.waitUntil(new Promise((resolve) => {
        if (typeof handleRequest === 'function') {
            handleRequest();
        }
        resolve();
    }));
});