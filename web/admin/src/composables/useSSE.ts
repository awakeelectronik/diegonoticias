import { ref } from 'vue'

export function useSSE() {
  const streaming = ref(false)
  let controller: AbortController | null = null

  async function start(url: string, body: unknown, onData: (chunk: string) => void, onDone: () => void, onError: (msg: string) => void) {
    controller = new AbortController()
    streaming.value = true
    try {
      const res = await fetch(url, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
        signal: controller.signal,
      })
      if (!res.ok || !res.body) {
        throw new Error('No se pudo iniciar streaming')
      }
      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const parts = buffer.split('\n\n')
        buffer = parts.pop() ?? ''
        for (const p of parts) {
          const lines = p.split('\n')
          const event = lines.find((l) => l.startsWith('event:'))?.slice(6).trim() ?? 'message'
          const data = lines.filter((l) => l.startsWith('data:')).map((l) => l.slice(5)).join('\n')
          if (event === 'error') {
            onError(data || 'Error de streaming')
            continue
          }
          if (event === 'done') {
            onDone()
            continue
          }
          onData(data)
        }
      }
      onDone()
    } catch (e) {
      if ((e as Error).name !== 'AbortError') {
        onError(e instanceof Error ? e.message : 'Error inesperado')
      }
    } finally {
      streaming.value = false
      controller = null
    }
  }

  function stop() {
    controller?.abort()
  }

  return { streaming, start, stop }
}

