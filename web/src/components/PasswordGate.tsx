import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { unlockURL } from '../api/urls'

export default function PasswordGate() {
  const { slug }            = useParams<{ slug: string }>()
  const [password, setPassword] = useState('')
  const [error, setError]       = useState('')
  const [loading, setLoading]   = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const { target_url } = await unlockURL(slug!, password)
      window.location.href = target_url
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Algo deu errado')
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-1 items-center justify-center px-4 py-12">
      <div className="w-full max-w-sm">
        <div className="rounded-2xl border border-zinc-800 bg-zinc-900 p-6 sm:p-8">
          <div className="mb-6 text-center">
            <span className="text-4xl">🔒</span>
            <h1 className="mt-3 text-xl font-semibold text-zinc-50">
              Senha necessária
            </h1>
            <p className="mt-1 text-sm text-zinc-400">
              Este link está protegido por senha
            </p>
            <code className="mt-2 inline-block rounded-md bg-zinc-800 px-2.5 py-1 text-xs text-zinc-400">
              /{slug}
            </code>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <input
              type="password"
              required
              placeholder="Digite a senha"
              value={password}
              onChange={e => setPassword(e.target.value)}
              autoFocus
              className="w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3.5 py-2.5 text-sm text-zinc-50 placeholder:text-zinc-500 outline-none transition focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
            />

            {error && (
              <p className="rounded-lg border border-red-800 bg-red-950/40 px-4 py-2.5 text-sm text-red-400">
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full rounded-lg bg-zinc-50 px-4 py-2.5 text-sm font-semibold text-zinc-950 transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-50"
            >
              {loading ? 'Verificando…' : 'Continuar'}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
