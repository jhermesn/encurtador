import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { CreateURLResponse } from '../api/urls'

interface Props {
  result: CreateURLResponse
  onCreateAnother: () => void
}

function CopyButton({
  text,
  className,
}: {
  text: string
  className?: string
}) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      onClick={copy}
      className={
        className ??
        'shrink-0 rounded-md border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-xs font-medium text-zinc-300 transition hover:bg-zinc-700'
      }
    >
      {copied ? '✓ Copiado' : 'Copiar'}
    </button>
  )
}

function formatDate(iso: string) {
  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(iso))
}

export default function ResultCard({ result, onCreateAnother }: Props) {
  const navigate = useNavigate()
  const [tokenCopied, setTokenCopied] = useState(false)

  async function copyToken() {
    await navigator.clipboard.writeText(result.manage_token)
    setTokenCopied(true)
  }

  return (
    <div className="mx-auto flex w-full max-w-lg flex-col px-4 py-10 sm:py-16">
      <div className="mb-6 text-center">
        <span className="text-4xl">✓</span>
        <h2 className="mt-2 text-xl font-semibold text-zinc-50">
          URL encurtada!
        </h2>
      </div>

      <div className="rounded-2xl border border-zinc-800 bg-zinc-900 p-6 sm:p-8 space-y-5">
        <div>
          <p className="mb-1.5 text-xs font-medium uppercase tracking-wide text-zinc-500">
            URL encurtada
          </p>
          <div className="flex items-center gap-2 rounded-lg border border-zinc-700 bg-zinc-800 px-3.5 py-2.5">
            <a
              href={result.short_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex-1 truncate text-sm font-medium text-zinc-100 hover:text-white hover:underline"
            >
              {result.short_url}
            </a>
            <CopyButton text={result.short_url} />
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-3 text-sm text-zinc-500">
          {result.protected && (
            <span className="flex items-center gap-1 rounded-full border border-zinc-700 bg-zinc-800 px-2.5 py-0.5 text-xs font-medium text-zinc-300">
              🔒 Protegido por senha
            </span>
          )}
          <span className="text-xs">Expira em {formatDate(result.expires_at)}</span>
        </div>

        {/* Management token — displayed once */}
        <div className="rounded-xl border border-amber-800/50 bg-amber-950/30 p-4">
          <div className="mb-2 flex items-center gap-2">
            <span>⚠️</span>
            <p className="text-sm font-semibold text-amber-300">
              Salve seu token de gerenciamento
            </p>
          </div>
          <p className="mb-3 text-xs leading-relaxed text-amber-400/80">
            Este token é exibido <strong>apenas uma vez</strong>. Guarde-o em um
            lugar seguro — você precisará dele para expirar esta URL antes do
            prazo.
          </p>
          <div className="flex items-center gap-2 rounded-lg border border-amber-800/40 bg-zinc-900 px-3 py-2">
            <code className="flex-1 break-all text-xs text-zinc-300">
              {result.manage_token}
            </code>
            <button
              onClick={copyToken}
              className="shrink-0 rounded-md border border-amber-800/50 bg-amber-950/40 px-3 py-1.5 text-xs font-medium text-amber-300 transition hover:bg-amber-950/60"
            >
              {tokenCopied ? '✓ Copiado' : 'Copiar'}
            </button>
          </div>
        </div>

        <div className="flex gap-3 pt-1">
          <button
            onClick={() => navigate(`/manage/${result.slug}`)}
            className="flex-1 rounded-lg border border-zinc-700 px-4 py-2.5 text-sm font-medium text-zinc-300 transition hover:bg-zinc-800"
          >
            Gerenciar
          </button>
          <button
            onClick={onCreateAnother}
            className="flex-1 rounded-lg bg-zinc-50 px-4 py-2.5 text-sm font-semibold text-zinc-950 transition hover:bg-white"
          >
            Criar Outra
          </button>
        </div>
      </div>
    </div>
  )
}
