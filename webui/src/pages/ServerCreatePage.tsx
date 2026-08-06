import ServerStepForm from "@/components/add-server/server-form";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import type { ServerTypeReturningValues } from "@/types/db";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

export default function InstanceCreatePage() {
  const navigate = useNavigate();

  const handleClose = useCallback(() => {
    navigate("/servers");
  }, []);
  //grow px-8 py-6 w-full max-w-[900px] mx-auto pb-24
  return (
    <main className="grow px-8 py-6 w-full mx-auto pb-24">
      {/* Header */}
      <div className="flex items-center gap-3 mb-8">
        <button
          onClick={() => navigate("/servers")}
          className="p-2 rounded-lg hover:bg-white/5 text-zinc-400 hover:text-white transition-colors"
        >
          <ArrowLeft size={18} />
        </button>
        <div>
          <h1 className="text-xl font-semibold text-white tracking-tight">
            Adicionar Novo Serviço / Servidor
          </h1>
          <p className="text-sm text-zinc-500">
            Cadastre um novo servidor para enviar dados monitorados.
          </p>
        </div>
      </div>
      <ServerStepForm onClose={handleClose} />
    </main>
  );
}

