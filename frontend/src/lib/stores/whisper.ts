import { writable, get } from 'svelte/store';
import { container } from '../services';
import { auth } from './auth';

export type WhisperStatus = 'idle' | 'recording' | 'processing' | 'completed' | 'failed';

interface WhisperState {
    status: WhisperStatus;
    taskId: string | null;
    targetTaskId: string | null; // Context ID (null means global NEW planet)
    result: string | null;
    error: string | null;
}

function createWhisperStore() {
    const { subscribe, set, update } = writable<WhisperState>({
        status: 'idle',
        taskId: null,
        targetTaskId: null,
        result: null,
        error: null
    });

    let mediaRecorder: MediaRecorder | null = null;
    let chunks: Blob[] = [];

    // Helper to get API URL
    const getApiUrl = () => {
        try {
            return container.getService<string>('API_URL');
        } catch (e) {
            return '/api/v1';
        }
    };

    // Placeholder key derivation (32 bytes)
    // Matches the "demo" key used in backend for Stage 2 verification
    async function getDemoMasterKeyRaw(): Promise<string> {
        return "0000000000000000000000000000000000000000000000000000000000000000";
    }

    async function getCryptoKey(): Promise<CryptoKey> {
        const hexKey = await getDemoMasterKeyRaw();
        const keyData = new Uint8Array(hexKey.match(/.{1,2}/g)!.map(byte => parseInt(byte, 16)));
        return await crypto.subtle.importKey(
            "raw",
            keyData,
            "AES-GCM",
            false,
            ["encrypt"]
        );
    }

    return {
        subscribe,
        reset: () => set({ status: 'idle', taskId: null, targetTaskId: null, result: null, error: null }),
        startRecording: async (targetTaskId: string | null = null) => {
            try {
                const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
                mediaRecorder = new MediaRecorder(stream);
                chunks = [];
                mediaRecorder.ondataavailable = (e) => {
                    if (e.data.size > 0) {
                        chunks.push(e.data);
                        console.log(`🎙️ Received audio chunk: ${e.data.size} bytes. Total chunks: ${chunks.length}`);
                    }
                };
                mediaRecorder.onstop = async () => {
                    const blob = new Blob(chunks, { type: mediaRecorder?.mimeType || 'audio/webm' });
                    console.log(`🎙️ Recording stopped. Final blob size: ${blob.size} bytes.`);
                    if (blob.size < 100) {
                        console.warn("⚠️ Recording is too small. Something is wrong with the audio capture.");
                    }
                    await whisper.process(blob);
                };
                mediaRecorder.start(100); 
                update(s => ({ ...s, status: 'recording', targetTaskId, error: null, result: null }));
            } catch (err: any) {
                console.error("Recording error:", err);
                update(s => ({ ...s, status: 'failed', error: "Microphone access denied or error." }));
            }
        },
        stopRecording: () => {
            if (mediaRecorder && mediaRecorder.state !== 'inactive') {
                mediaRecorder.stop();
                mediaRecorder.stream.getTracks().forEach(t => t.stop());
            }
        },
        process: async (blob: Blob) => {
            update(s => ({ ...s, status: 'processing' }));
            const apiUrl = getApiUrl();
            const token = get(auth).token;

            try {
                const arrayBuffer = await blob.arrayBuffer();
                const key = await getCryptoKey();
                const iv = crypto.getRandomValues(new Uint8Array(12)); // 12-byte nonce for GCM
                const encrypted = await crypto.subtle.encrypt(
                    { name: "AES-GCM", iv },
                    key,
                    arrayBuffer
                );

                // Combine IV (12) + Ciphertext + Tag (GCM)
                const combined = new Uint8Array(iv.length + encrypted.byteLength);
                combined.set(iv);
                combined.set(new Uint8Array(encrypted), iv.length);

                const formData = new FormData();
                formData.append('file', new Blob([combined]), 'voice.wav.enc');

                const hexKey = await getDemoMasterKeyRaw();
                const res = await fetch(`${apiUrl}/ai/whisper`, {
                    method: 'POST',
                    headers: { 
                        'Authorization': `Bearer ${token}`,
                        'X-Master-Key': hexKey
                    },
                    body: formData
                });

                if (!res.ok) throw new Error("Upload to Whisper-Engine failed.");
                const data = await res.json();
                update(s => ({ ...s, taskId: data.task_id }));

                // Start polling
                whisper.poll(data.task_id);
            } catch (err: any) {
                console.error("Whisper processing error:", err);
                update(s => ({ ...s, status: 'failed', error: err.message }));
            }
        },
        poll: async (id: string) => {
            const apiUrl = getApiUrl();
            const token = get(auth).token;
            
            const interval = setInterval(async () => {
                try {
                    const res = await fetch(`${apiUrl}/ai/whisper/status/${id}`, {
                        headers: { 'Authorization': `Bearer ${token}` }
                    });
                    if (!res.ok) throw new Error("Status check failed");
                    
                    const task = await res.json();
                    if (task.Status === 'completed') {
                        clearInterval(interval);
                        update(s => ({ ...s, status: 'completed', result: task.Result }));
                    } else if (task.Status === 'failed') {
                        clearInterval(interval);
                        update(s => ({ ...s, status: 'failed', error: task.Error }));
                    }
                } catch (e) {
                    console.error("Polling error:", e);
                    clearInterval(interval);
                    update(s => ({ ...s, status: 'failed', error: "Polling lost connection." }));
                }
            }, 1500);
        }
    };
}

export const whisper = createWhisperStore();
