import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth-store';
import { Button } from '@/components/ui/button';
import {
  Mail, Lock, Zap, Loader2, UserPlus, AlertCircle, Eye, EyeOff, User
} from 'lucide-react';

const inputCls =
  'w-full px-4 py-3 bg-white/[0.04] border border-white/[0.1] rounded-xl text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] focus:bg-white/[0.06] transition-all';
const labelCls = 'block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2';

export default function RegisterPage() {
  const navigate = useNavigate();
  const { register: doRegister, loading, error, user } = useAuthStore();
  
  const [firstName, setFirstName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    if (user) navigate('/workflow');
  }, [user, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const ok = await doRegister(email, password, firstName);
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
          <h1 className="text-2xl font-semibold text-white tracking-tight">Crie sua conta</h1>
          <p className="text-sm text-zinc-500 mt-2">
            Comece a monitorar seus dispositivos com Spark Edge.
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
                <User size={11} className="inline mr-1" />
                Nome Completo
              </label>
              <input
                type="text"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                placeholder="Seu nome"
                className={inputCls}
                disabled={loading}
              />
            </div>

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
                  autoComplete="new-password"
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
              <p className="text-[10px] text-zinc-600 mt-1.5">Mínimo de 6 caracteres</p>
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
                  <UserPlus size={18} />
                )}
                {loading ? 'Criando conta...' : 'Criar conta'}
              </Button>
            </div>
          </form>

          {/* Footer inside card */}
          <div className="mt-6 pt-5 border-t border-white/[0.06] text-center">
            <p className="text-sm text-zinc-500">
              Já tem uma conta?{' '}
              <Link to="/login" className="text-violet-400 hover:text-violet-300 font-medium transition-colors">
                Fazer login
              </Link>
            </p>
          </div>
        </div>

        {/* Terms */}
        <p className="mt-8 text-[11px] text-zinc-600 text-center max-w-[300px] mx-auto leading-relaxed">
          Ao se cadastrar, você concorda com nossos{' '}
          <Link to="/#" className="text-zinc-500 hover:text-zinc-400 underline underline-offset-2">Termos de Serviço</Link> e{' '}
          <Link to="/#" className="text-zinc-500 hover:text-zinc-400 underline underline-offset-2">Política de Privacidade</Link>.
        </p>
      </div>
    </div>
  );
}
