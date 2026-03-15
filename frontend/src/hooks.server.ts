import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
    // Proxy requests starting with /api to the backend container
    if (event.url.pathname.startsWith('/api')) {
        const backendUrl = 'http://backend:8080';
        const targetPath = event.url.pathname + event.url.search;
        const fullUrl = `${backendUrl}${targetPath}`;

        console.log(`[Proxy] Forwarding ${event.url.pathname} -> ${fullUrl}`);

        try {
            const requestHeaders = new Headers(event.request.headers);
            // We might need to adjust or remove certain headers if they cause issues
            // but for a simple proxy, forwarding works best.
            
            const response = await fetch(fullUrl, {
                method: event.request.method,
                headers: requestHeaders,
                body: event.request.method !== 'GET' && event.request.method !== 'HEAD' 
                    ? await event.request.arrayBuffer() 
                    : undefined,
                // @ts-ignore - duplex is needed for streaming bodies in some node versions
                duplex: 'half'
            });

            return response;
        } catch (err) {
            console.error(`[Proxy] Error forwarding request:`, err);
            return new Response(JSON.stringify({ error: 'Gateway Error', details: String(err) }), {
                status: 502,
                headers: { 'Content-Type': 'application/json' }
            });
        }
    }

    return resolve(event);
};
