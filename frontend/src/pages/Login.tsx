import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth-store';
import { Button } from '@/components/ui/button';
import {
  Mail, Lock, Zap, Loader2, LogIn, AlertCircle, Eye, EyeOff
} from 'lucide-react';

const inputCls =
  'w-full px-4 py-3 bg-white/[0.04] border border-white/[0.1] rounded-xl text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] focus:bg-white/[0.06] transition-all';
const labelCls = 'block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2';

export default function LoginPage() {
  const navigate = useNavigate();
  const { login, loading, error, user } = useAuthStore();
  
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    if (user) navigate('/workflow');
  }, [user, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const ok = await login(email, password);
    if (ok) navigate('/workflow');
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-6 bg-background">
      <div className="w-full max-w-[400px]">
        {/* Header */}
        <div className="mb-8 text-center">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-fuchsia-500 flex items-center justify-center mx-auto mb-4 shadow-lg shadow-violet-500/20">
            <Zap size={20} className="text-white" />
          </div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">Bem-vindo de volta</h1>
          <p className="text-sm text-zinc-500 mt-2">
            Acesse sua conta para gerenciar seus monitores.
          </p>
        </div>

        {/* Error banner */}
        {error && (
          <div className="flex items-start gap-3 bg-red-500/10 border border-red-500/20 rounded-xl px-4 py-3 mb-6 text-sm text-red-400 animate-in fade-in slide-in-from-top-2">
            <AlertCircle size={16} className="mt-0.5 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Main Card */}
        <div className="bg-white/[0.03] border border-white/[0.08] rounded-2xl p-6 shadow-2xl shadow-black/50">
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className={labelCls}>
                <Mail size={11} className="inline mr-1" />
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="voce@exemplo.com"
                className={inputCls}
                required
                autoComplete="email"
                disabled={loading}
              />
            </div>

            <div>
              <label className={labelCls}>
                <Lock size={11} className="inline mr-1" />
                Senha
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  className={`${inputCls} pr-12`}
                  required
                  autoComplete="current-password"
                  disabled={loading}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 transition-colors px-1"
                >
                  {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </div>

            <div className="pt-2">
              <Button
                type="submit"
                disabled={loading || !email || !password}
                className="w-full gap-2 bg-white text-zinc-900 hover:bg-zinc-100 font-semibold h-12 rounded-xl transition-all active:scale-[0.98]"
              >
                {loading ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : (
                  <LogIn size={18} />
                )}
                {loading ? 'Entrando...' : 'Entrar'}
              </Button>
            </div>
          </form>

          {/* Footer inside card */}
          <div className="mt-6 pt-5 border-t border-white/[0.06] text-center">
            <p className="text-sm text-zinc-500">
              Não tem uma conta?{' '}
              <Link to="/register" className="text-violet-400 hover:text-violet-300 font-medium transition-colors">
                Criar conta
              </Link>
            </p>
          </div>
        </div>

        {/* Forgot password */}
        <div className="mt-6 text-center">
          <Link to="/#" className="text-xs text-zinc-600 hover:text-zinc-400 transition-colors">
            Esqueceu sua senha?
          </Link>
        </div>
      </div>
    </div>
  );
}
