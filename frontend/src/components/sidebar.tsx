import React from 'react';
import { 
  Cpu, Package, FileText, History, Server, Smartphone, 
  FolderKanban, Settings, LogOut, ChevronRight, Clock, ShieldAlert, Zap
} from 'lucide-react';
import { Sidebar, SidebarContent, SidebarGroup, SidebarGroupLabel, SidebarHeader, SidebarRail } from './ui/sidebar';
import { NavMain } from './nav-main';
import { useAuthStore } from '@/stores/auth-store';
import { SidebarMenu, SidebarMenuItem, SidebarMenuButton, SidebarMenuAction } from './ui/sidebar';
import { useNavigate, useLocation } from 'react-router-dom';

const mainNav = [
  { url: '/instances', icon: <Cpu size={18} className="text-emerald-400" />, title: 'Instâncias' },
  { url: '/script-hub', icon: <Package size={18} className="text-violet-400" />, title: 'Script Hub' },
];

const dataNav = [
  { url: '/history', icon: <Clock size={18} className="text-pink-400" />, title: 'Histórico Execuções' },
  { url: '/credentials', icon: <ShieldAlert size={18} className="text-violet-400" />, title: 'Credenciais' },
  { url: '/data', icon: <FileText size={18} className="text-sky-400" />, title: 'Dados Locais' },
];

const managementNav = [
  { url: '/servers', icon: <Server size={18} className="text-orange-400" />, title: 'Servidores' },
  { url: '/devices', icon: <Smartphone size={18} className="text-pink-400" />, title: 'Dispositivos' },
  { url: '/projects', icon: <FolderKanban size={18} className="text-cyan-400" />, title: 'Projetos' },
  { url: '/settings/cloud', icon: <Zap size={18} className="text-emerald-400" />, title: 'Conectar ao Spark' },
];

export default function SideBar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();

  const handleSignOut = async () => {
    await logout();
    navigate('/login');
  };

  const initials = React.useMemo(() => {
    const name = user?.first_name ?? user?.email ?? '';
    const parts = name.split(/\s+/).filter(Boolean);
    if (parts.length === 0) return '';
    if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }, [user]);

  return (
    <Sidebar 
      className='border-r border-white/[0.06] h-full bg-zinc-950/80 backdrop-blur-xl'
      variant='sidebar' 
      collapsible="icon" 
      {...props}
    >
      <SidebarHeader className='flex items-center gap-3 p-4 border-b border-white/[0.06]'>
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-cyan-500 flex items-center justify-center">
            <span className="text-white font-bold text-[10px] leading-tight text-center">SPARK<br/>EDGE</span>
          </div>
          <span className="text-white font-semibold text-sm tracking-tight group-data-[collapsible=icon]:hidden">Spark Edge</span>
        </div>
      </SidebarHeader>
      <SidebarContent className='h-full overflow-hidden flex flex-col'>
        <SidebarGroup>
          <SidebarGroupLabel className="text-[10px] uppercase tracking-widest text-zinc-600 font-semibold px-3">Principal</SidebarGroupLabel>
          <NavMain items={mainNav} />
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel className="text-[10px] uppercase tracking-widest text-zinc-600 font-semibold px-3">Dados</SidebarGroupLabel>
          <NavMain items={dataNav} />
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel className="text-[10px] uppercase tracking-widest text-zinc-600 font-semibold px-3">Gestão</SidebarGroupLabel>
          <NavMain items={managementNav} />
        </SidebarGroup>

        <div className="mt-auto border-t border-white/[0.06]">
          <SidebarGroup>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton 
                  tooltip="Configurações" 
                  onClick={() => navigate('/settings')}
                  className={`h-10 text-zinc-400 hover:text-white hover:bg-white/[0.04] transition-all ${location.pathname === '/settings' ? 'text-white bg-white/[0.06]' : ''}`}
                >
                  <div className=''>

                  <Settings size={16} />
                  </div>
                  <span className="text-sm">Configurações</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton 
                  tooltip={user ? (user.first_name ?? user.email) : 'Not signed'} 
                  onClick={() => navigate('/settings')} 
                  className="group-data-[collapsible=icon]:p-0! h-10 hover:bg-white/[0.04] transition-all"
                >
                  <div className=''>

                  <div className='w-8 h-8 rounded-full bg-gradient-to-br from-violet-500 to-fuchsia-500 flex items-center justify-center text-white font-semibold text-[10px] select-none shrink-0'>
                    <span>{initials}</span>
                  </div>
                  </div>
                  <span className="flex flex-col min-w-0">
                    <span className='text-white text-xs font-medium truncate'>{user?.first_name ?? user?.email}</span>
                    <span className='text-zinc-500 text-[10px] truncate'>{user?.email}</span>
                  </span>
                </SidebarMenuButton>
                <SidebarMenuAction 
                  title='Sign out' 
                  aria-label='Sign out' 
                  onClick={handleSignOut} 
                  className='w-7 h-7 text-zinc-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all'
                >
                  <LogOut size={14} />
                </SidebarMenuAction>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>
        </div>
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  );
}
