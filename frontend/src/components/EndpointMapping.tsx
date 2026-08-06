import React from 'react';
import { useDrop } from 'react-dnd';
import { ItemTypes } from '@/lib/constants';
import { X, ArrowRight, CornerDownRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from './ui/button';

interface MappingItemProps {
  field: string;
  value: string;
  onRemove: () => void;
  onDrop: (path: string) => void;
  onRenameField: (oldField: string, newField: string) => void
}

function MappingItem({ field, value, onRemove, onDrop, onRenameField }: MappingItemProps) {
  const [{ isOver, canDrop }, drop] = useDrop(() => ({
    accept: ItemTypes.OUTPUT_VALUE,
    drop: (item: { value: string }) => onDrop(item.value),
    collect: (monitor) => ({
      isOver: monitor.isOver(),
      canDrop: monitor.canDrop(),
    }),
  }), [onDrop]);

  return (
    <div 
      ref={drop as any}
      className={cn(
        "flex items-center gap-3 p-2 rounded-lg border transition-all",
        isOver ? "bg-emerald-500/10 border-emerald-500/50 ring-1 ring-emerald-500/20" : 
        canDrop ? "bg-white/[0.04] border-white/20 border-dashed" : "bg-white/[0.02] border-white/[0.05]"
      )}
    >
      <div className="flex-1 min-w-0">
        <label className="text-[10px] uppercase tracking-wider text-zinc-500 font-semibold mb-1 block">Campo de Destino</label>
        <input 
          type="text" 
          value={field}
          onChange={(e) => onRenameField(field, e.target.value)}
          placeholder="Nome do campo de destino..."
          className="flex-1 bg-transparent border-none text-xs text-white placeholder:text-zinc-600 focus:ring-0"
        />
      </div>
      
      <ArrowRight className="text-zinc-700 shrink-0" size={14} />
      
      <div className="flex-[1.5] min-w-0">
        <label className="text-[10px] uppercase tracking-wider text-zinc-500 font-semibold mb-1 block">Fonte (JsonPath)</label>
        <div className={cn(
            "text-sm px-2 py-1.5 rounded bg-black/40 border border-white/[0.05] min-h-[28px] truncate",
            value ? "text-emerald-400 font-mono" : "text-zinc-600 italic"
        )}>
          {value || "Arraste um campo aqui..."}
        </div>
      </div>

      <Button 
        variant="ghost" 
        size="icon" 
        onClick={onRemove}
        className="h-8 w-8 text-zinc-600 hover:text-red-400 hover:bg-red-400/10"
      >
        <X size={14} />
      </Button>
    </div>
  );
}

interface EndpointMappingProps {
  mapping: Record<string, string>;
  onChange: (mapping: Record<string, string>) => void;
  suggestedFields?: string[]; // Based on endpoint payload schema if available
}

export function EndpointMapping({ mapping, onChange, suggestedFields = [] }: EndpointMappingProps) {
  const handleUpdateField = (field: string, path: string) => {
    onChange({ ...mapping, [field]: path });
  };

  const handleRemoveField = (field: string) => {
    const newMapping = { ...mapping };
    delete newMapping[field];
    onChange(newMapping);
  };

  const [newFieldName, setNewFieldName] = React.useState('');

  const handleAddField = () => {
    if (!newFieldName || mapping[newFieldName]) return;
    onChange({ ...mapping, [newFieldName]: '' });
    setNewFieldName('');
  };

  const handleRenameField = (oldField: string, newField: string) => {
    // Não renomeia se o novo nome já existe (exceto o próprio campo)
    if (!newField || (newField !== oldField && mapping[newField] !== undefined)) return;

    // Preserva a ordem das chaves substituindo in-place
    const entries = Object.entries(mapping).map(([k, v]) =>
      k === oldField ? [newField, v] : [k, v]
    );
    onChange(Object.fromEntries(entries));
  };

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        {Object.entries(mapping).map(([field, value]) => (
          <MappingItem 
            key={field}
            field={field}
            value={value}
            onRemove={() => handleRemoveField(field)}
            onDrop={(path) => handleUpdateField(field, path)}
            onRenameField={handleRenameField}
          />
        ))}

        {Object.keys(mapping).length === 0 && (
          <div className="py-8 text-center border-2 border-dashed border-white/[0.03] rounded-xl text-zinc-600">
            <CornerDownRight size={24} className="mx-auto mb-2 opacity-20" />
            <p className="text-xs">Nenhum mapeamento definido.<br/>Adicione campos de destino abaixo.</p>
          </div>
        )}
      </div>

      <div className="flex gap-2 p-3 bg-white/[0.02] border border-white/[0.05] rounded-xl">
        <input 
          type="text" 
          value={newFieldName}
          onChange={(e) => setNewFieldName(e.target.value)}
          placeholder="Nome do campo de destino..."
          className="flex-1 bg-transparent border-none text-xs text-white placeholder:text-zinc-600 focus:ring-0"
          onKeyDown={(e) => e.key === 'Enter' && handleAddField()}
        />
        <Button 
          size="sm" 
          onClick={handleAddField}
          disabled={!newFieldName || !!mapping[newFieldName]}
          className="bg-zinc-800 hover:bg-zinc-700 text-white text-[11px] h-8"
        >
          Adicionar Campo
        </Button>
      </div>

      {suggestedFields.length > 0 && suggestedFields.some(f => !mapping[f]) && (
        <div className="space-y-2">
          <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-bold">Sugestões (do schema)</p>
          <div className="flex flex-wrap gap-2">
            {suggestedFields.filter(f => !mapping[f]).map(field => (
              <button
                key={field}
                onClick={() => onChange({ ...mapping, [field]: '' })}
                className="px-2 py-1 rounded-md bg-white/[0.04] border border-white/[0.08] text-[10px] text-zinc-400 hover:bg-white/[0.08] transition-colors"
              >
                + {field}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

