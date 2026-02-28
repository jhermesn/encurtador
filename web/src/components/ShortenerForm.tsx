import { useState, useEffect, useRef, useCallback } from 'react'
import { createURL, checkSlug, type CreateURLResponse } from '../api/urls'
import ResultCard from './ResultCard'
import { translateError } from '../utils/errors'

const TTL_OPTIONS = [
  { value: '1h',    label: '1 Hora'   },
  { value: '24h',   label: '1 Dia'    },
  { value: '168h',  label: '1 Semana' },
  { value: '720h',  label: '1 Mês'    },
  { value: '8760h', label: '1 Ano'    },
]

const INPUT_CLASS =
  'w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3.5 py-2.5 text-sm text-zinc-50 placeholder:text-zinc-500 outline-none transition focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500'

type SlugStatus = 'idle' | 'checking' | 'available' | 'taken'

export default function ShortenerForm() {
  const [targetURL, setTargetURL]       = useState('')
  const [slug, setSlug]                 = useState('')
  const [ttl, setTtl]                   = useState('24h')
  const [password, setPassword]         = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [slugStatus, setSlugStatus]     = useState<SlugStatus>('idle')
  const [suggestion, setSuggestion]     = useState('')
  const [loading, setLoading]           = useState(false)
  const [error, setError]               = useState('')
  const [result, setResult]             = useState<CreateURLResponse | null>(null)

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const checkSlugAvailability = useCallback(async (value: string) => {
    if (!value) {
      setSlugStatus('idle')
      setSuggestion('')
      return
    }
    setSlugStatus('checking')
    try {
      const res = await checkSlug(value)
      setSlugStatus(res.available ? 'available' : 'taken')
      setSuggestion(res.suggestion ?? '')
    } catch {
      setSlugStatus('idle')
    }
  }, [])

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => checkSlugAvailability(slug), 500)
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [slug, checkSlugAvailability])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await createURL({
        target_url: targetURL,
        slug: slug || undefined,
        ttl,
        password: password || undefined,
      })
      setResult(res)
    } catch (err) {
      setError(err instanceof Error ? translateError(err.message) : 'Algo deu errado.')
    } finally {
      setLoading(false)
    }
  }

  if (result) {
    return (
      <ResultCard
        result={result}
        onCreateAnother={() => {
          setResult(null)
          setTargetURL('')
          setSlug('')
          setTtl('24h')
          setPassword('')
          setShowPassword(false)
          setSlugStatus('idle')
          setSuggestion('')
        }}
      />
    )
  }

  return (
    <div className="mx-auto flex w-full max-w-lg flex-col px-4 py-10 sm:py-16">
      <div className="mb-8 text-center">
        <h1 className="text-3xl font-bold tracking-tight text-zinc-50">
          Encurte sua URL
        </h1>
        <p className="mt-2 text-sm text-zinc-400">
          Gere um link curto em segundos — com expiração e proteção opcionais
        </p>
      </div>

      <form
        onSubmit={handleSubmit}
        className="rounded-2xl border border-zinc-800 bg-zinc-900 p-6 sm:p-8"
      >
        <div className="mb-5">
          <label className="mb-1.5 block text-sm font-medium text-zinc-300">
            URL longa <span className="text-red-400">*</span>
          </label>
          <input
            type="url"
            required
            placeholder="https://exemplo.com/caminho/muito/longo"
            value={targetURL}
            onChange={e => setTargetURL(e.target.value)}
            className={INPUT_CLASS}
          />
        </div>

        <div className="mb-5">
          <label className="mb-1.5 block text-sm font-medium text-zinc-300">
            Slug personalizado{' '}
            <span className="font-normal text-zinc-600">(opcional)</span>
          </label>
          <input
            type="text"
            placeholder="meu-slug"
            value={slug}
            onChange={e =>
              setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))
            }
            className={INPUT_CLASS}
          />
          {slug && (
            <p
              className={`mt-1.5 text-xs ${
                slugStatus === 'available'
                  ? 'text-emerald-400'
                  : slugStatus === 'taken'
                  ? 'text-red-400'
                  : 'text-zinc-500'
              }`}
            >
              {slugStatus === 'checking' && '⟳ Verificando…'}
              {slugStatus === 'available' && '✓ Disponível'}
              {slugStatus === 'taken' &&
                (suggestion
                  ? `✗ Ocupado — sugestão: ${suggestion}`
                  : '✗ Ocupado')}
            </p>
          )}
        </div>

        <div className="mb-5">
          <label className="mb-1.5 block text-sm font-medium text-zinc-300">
            Expira em <span className="text-red-400">*</span>
          </label>
          <select
            value={ttl}
            onChange={e => setTtl(e.target.value)}
            className={INPUT_CLASS}
          >
            {TTL_OPTIONS.map(opt => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div className="mb-6">
          <button
            type="button"
            onClick={() => {
              setShowPassword(!showPassword)
              if (showPassword) setPassword('')
            }}
            className="flex items-center gap-2 text-sm text-zinc-400 transition-colors hover:text-zinc-200"
          >
            <span>{showPassword ? '🔓' : '🔒'}</span>
            {showPassword
              ? 'Remover proteção de senha'
              : 'Adicionar proteção de senha'}
          </button>
          {showPassword && (
            <div className="mt-3">
              <input
                type="password"
                placeholder="Senha para este link"
                value={password}
                onChange={e => setPassword(e.target.value)}
                className={INPUT_CLASS}
              />
            </div>
          )}
        </div>

        {error && (
          <p className="mb-4 rounded-lg border border-red-800 bg-red-950/40 px-4 py-3 text-sm text-red-400">
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={loading}
          className="w-full rounded-lg bg-zinc-50 px-4 py-2.5 text-sm font-semibold text-zinc-950 transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? 'Encurtando…' : 'Encurtar URL'}
        </button>
      </form>
    </div>
  )
}
