/**
 * Instance Edit Page
 * Page for editing existing instances
 */

import { useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import InstanceStepForm from "@/components/add-instance/instance-step-form";

export default function InstanceEditPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();

  const handleClose = useCallback(() => {
    navigate("/instances");
  }, [navigate]);

  if (!id) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <p className="text-zinc-500">ID da instância não fornecido.</p>
      </div>
    );
  }

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
            Editar Instância
          </h1>
          <p className="text-sm text-zinc-500">
            Atualize a configuração desta instância de automação.
          </p>
        </div>
      </div>

      {/* Form */}
      <div className="max-w-6xl">
        <InstanceStepForm instanceId={id} onClose={handleClose} />
      </div>
    </main>
  );
}

