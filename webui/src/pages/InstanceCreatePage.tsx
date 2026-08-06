/**
 * Instance Create Page
 * Page for creating new instances
 */

import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import InstanceStepForm from "@/components/add-instance/instance-step-form";

export default function InstanceCreatePage() {
  const navigate = useNavigate();

  const handleClose = useCallback(() => {
    navigate("/instances");
  }, [navigate]);

  return (
    <main className="grow px-8 py-6 w-full mx-auto pb-24">
      {/* Header */}
      <div className="flex items-center gap-3 mb-8">
        <button
          onClick={handleClose}
          className="p-2 rounded-lg hover:bg-white/5 text-zinc-400 hover:text-white transition-colors"
          aria-label="Voltar"
        >
          <ArrowLeft size={18} />
        </button>
        <div>
          <h1 className="text-xl font-semibold text-white tracking-tight">
            Criar Nova Instância
          </h1>
          <p className="text-sm text-zinc-500">
            Configure uma nova instância de automação com triggers, destinos e
            mapeamento de dados.
          </p>
        </div>
      </div>

      {/* Form */}
      <div className="max-w-6xl">
        <InstanceStepForm onClose={handleClose} />
      </div>
    </main>
  );
}

