import axios from "axios";

export function interpolate(template: string, params: Record<string, string> = {}) {
  return template.replace(/{{(.*?)}}/g, (_, key) => params[key.trim()] ?? null)
}

export function extractEndpointKeys(endpoint: string): string[] {
  const regex = /\{\{([^{}]+)\}\}/g;
  return Array.from(endpoint.matchAll(regex), m => m[1].trim());
}

type EndpointPart =
  | { type: "static"; value: string, name: string }
  | { type: "param"; value: string, name: string };

export function parseEndpoint(endpoint: string): EndpointPart[] {
  const regex = /\{\{[^{}]+\}\}/g;
  const parts: EndpointPart[] = [];
  let lastIndex = 0;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(endpoint)) !== null) {
    const staticPart = endpoint.slice(lastIndex, match.index);
    if (staticPart) {
      parts.push({ type: "static", value: staticPart, name: "" });
    }

    const paramName = match[0].slice(2, -2).trim();
    parts.push({ type: "param", value: match[0], name: paramName }); // mantém os {{ }}
    lastIndex = regex.lastIndex;
  }

  // adiciona o que sobra no final
  if (lastIndex < endpoint.length) {
    const tail = endpoint.slice(lastIndex);
    if (tail) parts.push({ type: "static", value: tail, name: "" });
  }

  return parts;
}
