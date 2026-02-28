import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { expireURL } from '../api/urls'
import { translateError } from '../utils/errors'

export default function ManagePage() {
  const { slug }  = useParams<{ slug: string }>()
  const navigate  = useNavigate()
  const [token, setToken]     = useState('')
  const [error, setError]     = useState('')
  const [loading, setLoading] = useState(false)
  const [expired, setExpired] = useState(false)

  async function handleExpire(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await expireURL(slug!, token)
      setExpired(true)
    } catch (err) {
      setError(err instanceof Error ? translateError(err.message) : 'Algo deu errado.')
    } finally {
      setLoading(false)
    }
  }

  if (expired) {
    return (
      <div className="flex flex-1 items-center justify-center px-4 py-12">
        <div className="w-full max-w-sm">
          <div className="rounded-2xl border border-zinc-800 bg-zinc-900 p-6 sm:p-8 text-center">
            <span className="text-4xl">✓</span>
            <h2 className="mt-3 text-xl font-semibold text-zinc-50">
              URL expirada
            </h2>
            <p className="mt-2 text-sm text-zinc-400">
              <code className="rounded bg-zinc-800 px-1.5 py-0.5 text-xs text-zinc-300">
                /{slug}
              </code>{' '}
              não está mais acessível.
            </p>
            <button
              onClick={() => navigate('/')}
              className="mt-6 w-full rounded-lg bg-zinc-50 px-4 py-2.5 text-sm font-semibold text-zinc-950 transition hover:bg-white"
            >
              Criar uma nova URL
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-1 items-center justify-center px-4 py-12">
      <div className="w-full max-w-sm">
        <div className="rounded-2xl border border-zinc-800 bg-zinc-900 p-6 sm:p-8">
          <div className="mb-6">
            <h1 className="text-xl font-semibold text-zinc-50">
              Gerenciar URL
            </h1>
            <code className="mt-1 inline-block rounded-md bg-zinc-800 px-2.5 py-1 text-xs text-zinc-400">
              /{slug}
            </code>
          </div>

          <form onSubmit={handleExpire} className="space-y-4">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-zinc-300">
                Token de gerenciamento
              </label>
              <input
                type="text"
                required
                placeholder="Cole seu token aqui"
                value={token}
                onChange={e => setToken(e.target.value)}
                className="w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3.5 py-2.5 font-mono text-xs text-zinc-50 placeholder:text-zinc-500 outline-none transition focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
              />
            </div>

            <div className="rounded-lg border border-red-900/50 bg-red-950/30 px-4 py-3 text-xs text-red-400">
              ⚠️ Expirar uma URL é <strong>permanente</strong> e não pode ser
              desfeito. Qualquer acesso ao link retornará 404.
            </div>

            {error && (
              <p className="rounded-lg border border-red-800 bg-red-950/40 px-4 py-2.5 text-sm text-red-400">
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={loading || !token}
              className="w-full rounded-lg bg-red-600 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-red-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {loading ? 'Expirando…' : 'Expirar URL Agora'}
            </button>
          </form>

          <button
            onClick={() => navigate('/')}
            className="mt-4 w-full rounded-lg border border-zinc-700 px-4 py-2 text-sm text-zinc-400 transition hover:bg-zinc-800"
          >
            Voltar para o início
          </button>
        </div>
      </div>
    </div>
  )
}
