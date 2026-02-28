import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import ShortenerForm from './components/ShortenerForm'
import PasswordGate from './components/PasswordGate'
import ManagePage from './components/ManagePage'
import NotFound from './components/NotFound'

function Header() {
  return (
    <header className="border-b border-zinc-800/60">
      <div className="mx-auto flex max-w-screen-sm items-center justify-between px-4 py-3">
        <Link to="/" className="flex items-center opacity-90 transition-opacity hover:opacity-100">
          <img
            src={`${import.meta.env.BASE_URL}linklogo-nobg.png`}
            alt="Encurtador"
            className="h-8 w-auto"
          />
        </Link>
        <a
          href="https://jhermesn.dev"
          className="text-xs text-zinc-500 transition-colors hover:text-zinc-300"
        >
          jhermesn.dev →
        </a>
      </div>
    </header>
  )
}

function Footer() {
  return (
    <footer className="border-t border-zinc-800/60 py-6">
      <p className="text-center text-xs text-zinc-600">
        <a href="https://jhermesn.dev" className="hover:underline">Jorge Hermes</a> &copy; 2026. Todos os Direitos Reservados.
      </p>
    </footer>
  )
}

export default function App() {
  return (
    <BrowserRouter basename="/encurtador">
      <div className="flex min-h-screen flex-col">
        <Header />
        <main className="flex-1">
          <Routes>
            <Route path="/" element={<ShortenerForm />} />
            <Route path="/gate/:slug" element={<PasswordGate />} />
            <Route path="/manage/:slug" element={<ManagePage />} />
            <Route path="*" element={<NotFound />} />
          </Routes>
        </main>
        <Footer />
      </div>
    </BrowserRouter>
  )
}
