// Idiomatic JS wrapper around the express utilities WebAssembly adapter.
//
//   import { loadExpress } from './express.mjs';
//   const ex = await loadExpress();          // browser (fetch) or Node
//   ex.ms('2h');                             // 7200000
//   ex.bytesFormat(1536);                    // '1.5KB'
//   ex.uuidV4();                             // random v4 uuid
//   ex.slugify('Hello World');               // 'hello-world'
//   ex.isEmail('a@b.com');                   // true
//
// The same Go implementations that power the express module run here via wasm.
// Only the portable, pure UTILITY subpackages are exposed — not the HTTP server.

async function ensureGo() {
  if (typeof globalThis.Go === 'function') return;
  if (typeof window === 'undefined') {
    // Node: wasm_exec.js is a classic script that assigns globalThis.Go.
    const { readFileSync } = await import('node:fs');
    const { fileURLToPath } = await import('node:url');
    const path = fileURLToPath(new URL('./wasm_exec.js', import.meta.url));
    const { runInThisContext } = await import('node:vm');
    runInThisContext(readFileSync(path, 'utf8'));
  } else {
    await import('./wasm_exec.js');
  }
}

async function readWasm(wasmPath) {
  if (typeof window === 'undefined') {
    const { readFileSync } = await import('node:fs');
    const { fileURLToPath } = await import('node:url');
    const p = wasmPath ?? fileURLToPath(new URL('./express.wasm', import.meta.url));
    return readFileSync(p);
  }
  const res = await fetch(wasmPath ?? new URL('./express.wasm', import.meta.url));
  return new Uint8Array(await res.arrayBuffer());
}

export async function loadExpress(wasmPath) {
  await ensureGo();
  const go = new globalThis.Go();
  const bytes = await readWasm(wasmPath);
  const { instance } = await WebAssembly.instantiate(bytes, go.importObject);
  go.run(instance); // long-running; resolves when the module exits (it won't)
  const g = globalThis.__mgo_express;
  if (!g) throw new Error('express wasm did not register __mgo_express');

  return {
    // ms
    ms: (str) => g.ms(String(str)),
    msFormat: (millis) => g.msFormat(Number(millis)),
    // bytes
    bytes: (str) => g.bytes(String(str)),
    bytesFormat: (n) => g.bytesFormat(Number(n)),
    // uuid
    uuidV4: () => g.uuidV4(),
    uuidValidate: (s) => g.uuidValidate(String(s)),
    // nanoid
    nanoid: () => g.nanoid(),
    // slugify (lowercases by default, like the common npm usage)
    slugify: (text, lower = true) => g.slugify(String(text), Boolean(lower)),
    // validatorjs
    isEmail: (s) => g.isEmail(String(s)),
    isURL: (s) => g.isURL(String(s)),
    // pluralize
    plural: (word) => g.plural(String(word)),
    // titlecase
    titleCase: (s) => g.titleCase(String(s)),
  };
}
