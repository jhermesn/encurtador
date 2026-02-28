import { useNavigate } from 'react-router-dom'

export default function NotFound() {
  const navigate = useNavigate()

  return (
    <div className="flex flex-1 items-center justify-center px-4 py-12">
      <div className="w-full max-w-sm">
        <div className="rounded-2xl border border-zinc-800 bg-zinc-900 p-6 sm:p-8 text-center">
          <span className="text-4xl">🔗</span>
          <h2 className="mt-3 text-xl font-semibold text-zinc-50">
            Link não encontrado
          </h2>
          <p className="mt-2 text-sm text-zinc-400">
            Este link não existe ou já expirou.
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
