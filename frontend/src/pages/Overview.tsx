import InstanceTable from "@/components/instance-table";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useNavigate } from "react-router-dom";

export default function Overview() {
    const navigate = useNavigate();
    
    return (
        <main className="grow px-10 py-6 w-full">
            <div className="flex flex-row justify-between">
                <div>
                    <h1 className="text-primary text-2xl">Visão Geral</h1>
                    <span className="text-primary/60">Instâncias criadas por você.</span>
                </div>
                <div>
                    <Button
                        onClick={() => navigate('/instances/new')}
                        className="gap-2 bg-white text-zinc-900 hover:bg-zinc-200 font-medium"
                    >
                        <Plus size={16} />
                        Nova Instância
                    </Button>
                </div>
            </div>
            <section>
                <InstanceTable />
            </section>
        </main>
    )
}
