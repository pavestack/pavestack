import "@testing-library/jest-dom";

// Node >= 22 defines an experimental `localStorage` global that is undefined
// unless node runs with --localstorage-file; it shadows jsdom's Storage in
// vitest's global proxy. Install a spec-shaped in-memory Storage so code and
// tests using window.localStorage behave the same as in a browser.
function createStorage(): Storage {
  let store = new Map<string, string>();
  const storage = {
    getItem: (key: string) => (store.has(key) ? store.get(key)! : null),
    setItem: (key: string, value: string) => {
      store.set(String(key), String(value));
    },
    removeItem: (key: string) => {
      store.delete(key);
    },
    clear: () => {
      store = new Map();
    },
    key: (index: number) => Array.from(store.keys())[index] ?? null,
    get length() {
      return store.size;
    },
  };
  return storage as Storage;
}

for (const name of ["localStorage", "sessionStorage"] as const) {
  Object.defineProperty(globalThis, name, {
    value: createStorage(),
    writable: true,
    configurable: true,
  });
}

// jsdom's matchMedia doesn't survive vitest's global proxy either; provide a
// minimal implementation (tests that care spy on it and mock the result).
if (typeof globalThis.matchMedia !== "function") {
  Object.defineProperty(globalThis, "matchMedia", {
    value: (query: string): MediaQueryList =>
      ({
        matches: false,
        media: query,
        onchange: null,
        addEventListener: () => {},
        removeEventListener: () => {},
        addListener: () => {},
        removeListener: () => {},
        dispatchEvent: () => false,
      }) as unknown as MediaQueryList,
    writable: true,
    configurable: true,
  });
}
