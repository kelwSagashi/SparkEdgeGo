"use client"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "./ui/sidebar"
import { Link, useLocation } from "react-router-dom"
import type { JSX } from "react"

export function NavMain({
  items,
}: {
  items: {
    title: string
    url: string
    icon?: JSX.Element
  }[]
}) {
  const location = useLocation();

  return (
    <SidebarMenu>
      {items.map((item) => {
        const isActive = location.pathname === item.url || location.pathname.startsWith(item.url + '/');
        return (
          <SidebarMenuItem key={item.title}>
            <Link to={item.url}>
              <SidebarMenuButton
                tooltip={item.title}
                className={`h-9 gap-3 rounded-lg transition-all duration-200 ${
                  isActive
                    ? 'bg-white/[0.08] text-white font-medium'
                    : 'text-zinc-500 hover:text-zinc-300 hover:bg-white/[0.04]'
                }`}
              >
                {item.icon}
                <span className="text-sm">{item.title}</span>
              </SidebarMenuButton>
            </Link>
          </SidebarMenuItem>
        );
      })}
    </SidebarMenu>
  )
}

