export async function uploadImage(file: File, csrfToken: string, alt = '') {
  const fd = new FormData()
  fd.set('image', file)
  fd.set('alt', alt)
  const res = await fetch('/admin/api/imagenes', {
    method: 'POST',
    credentials: 'include',
    headers: {
      'X-CSRF-Token': csrfToken,
    },
    body: fd,
  })
  const data = (await res.json()) as { basePath?: string; alt?: string; error?: string }
  if (!res.ok) {
    throw new Error(data.error ?? 'No se pudo subir imagen')
  }
  return data as { basePath: string; alt: string }
}

