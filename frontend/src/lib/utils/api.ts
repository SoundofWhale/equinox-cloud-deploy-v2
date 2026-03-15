export async function uploadForOCR(file: File): Promise<{ text: string; summary: string }> {
    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch('/api/v1/ai/ocr', {
        method: 'POST',
        body: formData
    });

    if (!response.ok) {
        throw new Error('OCR upload failed');
    }

    return response.json();
}

export async function uploadFile(taskId: string, file: File): Promise<{ filename: string; url: string }> {
    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch(`/api/v1/upload/${taskId}`, {
        method: 'POST',
        body: formData
    });

    if (!response.ok) {
        throw new Error('File upload failed');
    }

    return response.json();
}



export async function queryAI(role: 'cto' | 'zen', prompt: string): Promise<{ response: string }> {
    const response = await fetch('/api/v1/ai/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role, prompt })
    });

    if (!response.ok) {
        throw new Error('AI query failed');
    }

    return response.json();
}
