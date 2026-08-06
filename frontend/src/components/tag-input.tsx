import React, { useState, useEffect, useRef, useCallback } from 'react';
import { X } from 'lucide-react';
import { useTagsStore } from '@/stores/tags-store';
import type { Tag } from '@/rest-api-client/tags.service';

interface TagInputProps {
  value: Tag[];
  onChange: (tags: Tag[]) => void;
  projectId?: string;
}

export default function TagInput({ value, onChange, projectId }: TagInputProps) {
  const { search, createTag } = useTagsStore();
  const [input, setInput] = useState('');
  const [suggestions, setSuggestions] = useState<Tag[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const doSearch = useCallback(async (q: string) => {
    const results = await search(q, projectId);
    setSuggestions(results.filter(t => !value.some(v => v.id === t.id)));
  }, [search, value, projectId]);

  useEffect(() => {
    const timer = setTimeout(() => doSearch(input), 200);
    return () => clearTimeout(timer);
  }, [input, doSearch]);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setShowSuggestions(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const addTag = (tag: Tag) => {
    onChange([...value, tag]);
    setInput('');
    setSuggestions([]);
    inputRef.current?.focus();
  };

  const removeTag = (id: string) => {
    onChange(value.filter(t => t.id !== id));
  };

  const handleKeyDown = async (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && input.trim()) {
      e.preventDefault();
      const existing = suggestions.find(s => s.name.toLowerCase() === input.trim().toLowerCase());
      if (existing) {
        addTag(existing);
      } else {
        const created = await createTag(input.trim(), projectId);
        if (created) addTag(created);
      }
    }
    if (e.key === 'Backspace' && !input && value.length > 0) {
      removeTag(value[value.length - 1].id);
    }
  };

  return (
    <div ref={containerRef} className="relative">
      <div className="flex flex-wrap items-center gap-1.5 px-3 py-2 bg-white/[0.04] border border-white/[0.1] rounded-xl min-h-[42px] focus-within:border-white/[0.2] transition-colors">
        {value.map(tag => (
          <span
            key={tag.id}
            className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium bg-white/[0.08] text-zinc-300 border border-white/[0.08]"
          >
            <span className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: tag.color ?? '#6b7280' }} />
            {tag.name}
            <button type="button" onClick={() => removeTag(tag.id)} className="ml-0.5 hover:text-white transition-colors">
              <X size={10} />
            </button>
          </span>
        ))}
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={e => { setInput(e.target.value); setShowSuggestions(true); }}
          onFocus={() => setShowSuggestions(true)}
          onKeyDown={handleKeyDown}
          placeholder={value.length === 0 ? 'Adicionar tags...' : ''}
          className="flex-1 min-w-[80px] bg-transparent text-sm text-white placeholder:text-zinc-600 outline-none"
        />
      </div>
      {showSuggestions && suggestions.length > 0 && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-zinc-900 border border-white/[0.1] rounded-xl shadow-xl z-50 max-h-[200px] overflow-y-auto">
          {suggestions.map(tag => (
            <button
              key={tag.id}
              type="button"
              onClick={() => addTag(tag)}
              className="flex items-center gap-2 w-full px-3 py-2 text-left text-sm text-zinc-300 hover:bg-white/[0.06] transition-colors"
            >
              <span className="w-2 h-2 rounded-full" style={{ backgroundColor: tag.color ?? '#6b7280' }} />
              {tag.name}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

