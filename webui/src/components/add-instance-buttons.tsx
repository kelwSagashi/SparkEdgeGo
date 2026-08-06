"use client"
import { useState } from "react";
import {
  ChevronDownIcon,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "./ui/button-group"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useNavigate } from "react-router-dom";

export function AddInstanceButtonGroup({
  setIsServerDialogOpen,
  setIsDeviceDialogOpen
}: {
  setIsServerDialogOpen: React.Dispatch<React.SetStateAction<boolean>>
  setIsDeviceDialogOpen: React.Dispatch<React.SetStateAction<boolean>>
}) {
  const navigate = useNavigate();

  const handleNavigateToEditor = () => {
    navigate(`/workflow/new`);
  };

  return (
    <ButtonGroup>
      <Button
        variant="outline"
        className="bg-accent-foreground border-border/30 rounded text-primary hover:bg-accent-foreground/70 hover:text-primary"
        onClick={() => handleNavigateToEditor()}
      >
        Adicionar Instância
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="pl-2! bg-accent-foreground border-border/30 rounded text-primary hover:bg-accent-foreground/70 hover:text-primary">
            <ChevronDownIcon />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="px-0 [--radius:1rem] bg-background border-border/30 rounded text-primary">
          <DropdownMenuGroup>
            <DropdownMenuItem
              variant="default"
              className="text-input rounded-none focus:bg-muted/10 focus:text-primary"
              onSelect={() => setIsServerDialogOpen(true)}
            >
              Criar Servidor
            </DropdownMenuItem>
            <DropdownMenuItem
              variant="default"
              className="text-input rounded-none focus:bg-muted/10 focus:text-primary"
              onSelect={() => setIsDeviceDialogOpen(true)}
            >
              Criar Dispositivo
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>


    </ButtonGroup>
  )
}
